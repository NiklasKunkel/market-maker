package maker

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"github.com/niklaskunkel/market-maker/api"
)


//keeps track of all active orders
var orderBook = new(Orders)

type Orders struct {
	Asks	map[string]Order
	Bids 	map[string]Order
}

func topUpBands(gatecoin *api.GatecoinClient, bands Bands, targetPrice float64) {
	//synchronize order book
	err := synchronizeOrders(gatecoin)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	//create new buy and sell orders in all buy/sell bands
	topUpBuyBands(gatecoin, getBuyOrders(), bands.BuyBands, targetPrice)
	topUpSellBands(gatecoin, getSellOrders(), bands.SellBands, targetPrice)
}

func topUpBuyBands(gatecoin *api.GatecoinClient, orders []*Order, buyBands []BuyBand, targetPrice float64) {
	//get balance
 	availableBalances, err := gatecoin.GetBalances("USD")
 	if err != nil {
		fmt.Printf("[GATECOIN] Failed to get balances due to: %s\n", err.Error())
	}
	availableUsdBalance := availableBalances.Balances[0].AvailableBalance

 	buyOrders := []*Order{}
 	//iterate through bandsb 
 	for _, buyBand := range buyBands {
 		//get all buy orders in this band
 		for _, order := range orders {
 			if buyBand.Includes(order.Price, targetPrice) {
 				buyOrders = append(buyOrders, order)
 			}
 		}
 		//get totalAmount of orders
 		totalAmount := buyBand.TotalAmount(buyOrders)
 		//if total order amount is below minimum band threshold
 		if (totalAmount < buyBand.MinAmount) {
 			//get order parameters
 			price := buyBand.AvgPrice(targetPrice)
 			payAmount := math.Min(buyBand.AvgAmount - totalAmount, availableUsdBalance)
 			buyAmount := payAmount / price
 			//verify order parameters
 			if ((payAmount >= buyBand.DustCutoff) && (payAmount > float64(0)) && (buyAmount > float64(0))) {
 				//Log order creation
 				fmt.Printf("[GATECOIN] Creating Buy Order for pair: %s, for amount: %f, at price: %f. Remaining balance is: %f\n", "DAIUSD", buyAmount, price, availableUsdBalance - payAmount)
 				//create order
 				resp, err := gatecoin.CreateOrder("DAIUSD", "bid", strconv.FormatFloat(buyAmount, 'f', 6, 64), strconv.FormatFloat(price, 'f', 6, 64))	//not sure if this is payAmount or buyAmount
 				fmt.Printf("%+v", resp)
 				if err != nil {
 					fmt.Printf("Failed to create Gatecoin buy order due to error: %s\n", err.Error())
 					continue
 				}
 			}
 		}
 		buyOrders = nil
 	}
 	return
}

func topUpSellBands(gatecoin *api.GatecoinClient, orders []*Order, sellBands []SellBand, targetPrice float64) {
	availableBalances, err := gatecoin.GetBalances("DAI")
	if err != nil {
		fmt.Printf("[GATECOIN] Failed to get balances due to: %s\n", err.Error())
	}
	availableDaiBalance := availableBalances.Balances[0].AvailableBalance 

	sellOrders := []*Order{}
 	//iterate through bandsb 
 	for _, sellBand := range sellBands {
 		//get all buy orders in this band
 		for _, order := range orders {
 			if sellBand.Includes(order.Price, targetPrice) {
 				sellOrders = append(sellOrders, order)
 			}
 		}
 		//get totalAmount of orders
 		totalAmount := sellBand.TotalAmount(sellOrders)
 		//if total order amount is below minimum band threshold
 		if (totalAmount < sellBand.MinAmount) {
 			//get order parameters
 			price := sellBand.AvgPrice(targetPrice)
 			payAmount := math.Min(sellBand.AvgAmount - totalAmount, availableDaiBalance)
 			buyAmount := payAmount / price
 			//verify order parameters
 			if ((payAmount >= sellBand.DustCutoff) && (payAmount > float64(0)) && (buyAmount > float64(0))) {
 				//Log order creation
 				fmt.Printf("[GATECOIN] Creating Sell Order for pair: DAIUSD, for amount: %f, at price: %f. Remaining balance is: %f\n", buyAmount, price, availableDaiBalance - payAmount)
 				//create order
 				resp, err := gatecoin.CreateOrder("DAIUSD", "ask", strconv.FormatFloat(buyAmount, 'f', 6, 64), strconv.FormatFloat(price, 'f', 6, 64))	//not sure if this is payAmount or buyAmount
 				fmt.Printf("%+v\n", resp)
 				if err != nil {
 					fmt.Printf("Failed to create Gatecoin sell order due to error: %s\n", err.Error())
 					continue
 				}
 			}
 		}
 		sellOrders = nil
 	}
 	return
}

func cancelAllOrders(gatecoin *api.GatecoinClient) {
	synchronizeOrders(gatecoin)
	fmt.Printf("Cancelling all orders...\n")
	for id, _ := range orderBook.Bids {
		fmt.Printf("Cancelling Order %s...\n", id)
		resp, err := gatecoin.DeleteOrder(id)
		if err != nil {
			fmt.Printf("Error: Failed to cancel order %s due to: %s\n", id, err.Error())
		}
		fmt.Printf("%s", resp.Status.Message)
	}
	for id, _ := range orderBook.Asks {
		fmt.Printf("Cancelling Order %s...\n", id)
		resp, err := gatecoin.DeleteOrder(id)
		if err != nil {
			fmt.Printf("Error: Failed to cancel order %s due to: %s\n", id, err.Error())
		}
		fmt.Printf("%s", resp.Status.Message)
	}
}

func getRefPrice(pair string) (float32) {
	if (strings.ToLower(pair) == "dai") {
		return 1.00
	}
	return 0
}

func getOrders() (*Orders) {
	return orderBook
}

func getBuyOrders() (bids []*Order) {
	for _, bid := range orderBook.Bids {
		bids = append(bids, &bid)
	}
	return bids
}

func getSellOrders() (asks []*Order) {
	for _, ask := range orderBook.Asks {
		asks = append(asks, &ask)
	}
	return asks
}

func getTotalOrderAmount(orders []*Order) (sum float64) {
	for _, order := range orders {
		sum += order.RemQuantity
	}
	return sum
}

func getFeedPrice(currency string) {
	
}

//Updates the in-memory orderbook.
func synchronizeOrders(gatecoin *api.GatecoinClient) (error) {
	resp, err := gatecoin.GetOrders()
	if err != nil {
		return fmt.Errorf("[GATECOIN] Failed to synchronize orders due to: %s\n", err.Error)
	} else if (resp.Status.Message != "OK") {
		return fmt.Errorf("[GATECOIN] Failed to synchronize orders due to: %+v\n", resp.Status)
	}
	//reset orderbook
	orderBook.Asks  = nil
	orderBook.Bids = nil

	//populate orderbook
	for _, order := range resp.Orders {
		if (order.Side == 0) {
			orderBook.Bids[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		} else if (order.Side == 1) {
			orderBook.Asks[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		}
	}
	return nil
}

//NOTES
//Keep track of all actions in a log - order made - order cancelled
	//These should be in easy to follow format, probably JSON of GetOrder(id)
	//JSON format would help for parsing to create analytics later

//Keep an internal orderbook, before doing anything call synchronizeOrders() in order to update the internal orderbook.
//maybe dont have orderbook be a global and just have it be initialized in tupUpBands() and then passed to synchronizeOrders and topUpBuyBands and topUpSellBands

