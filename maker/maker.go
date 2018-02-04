package maker

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
	"github.com/niklaskunkel/market-maker/api"
	"github.com/olekukonko/tablewriter"
)


//keeps track of all active orders
var orderBook = new(Orders)

type Orders struct {
	Asks	map[string]Order
	Bids 	map[string]Order
}

func MarketMaker(gatecoin *api.GatecoinClient, bands *Bands, pair string) {
	//synchronize order book
	err := synchronizeOrders(gatecoin)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	//get reference price
	refPrice, err := getFeedPrice(pair)
	if err != nil {
		return
	}
	cancelExcessOrders(gatecoin, bands.CancellableOrders(getBuyOrders(), getSellOrders(), refPrice))
	topUpBands(gatecoin, bands, refPrice)
	PrintOrderBook(gatecoin)
}

//Updates the in-memory orderbook.
func synchronizeOrders(gatecoin *api.GatecoinClient) (error) {
	fmt.Printf("synchronizing orderbook...\n")
	resp, err := gatecoin.GetOrders()
	if err != nil {
		return fmt.Errorf("[GATECOIN] Failed to synchronize orders due to: %s\n", err.Error)
	} else if (resp.Status.Message != "OK") {
		return fmt.Errorf("[GATECOIN] Failed to synchronize orders due to: %+v\n", resp.Status)
	}
	//reset orderbook
	orderBook.Asks = nil
	orderBook.Bids = nil
	orderBook.Asks = make(map[string]Order)
	orderBook.Bids = make(map[string]Order)

	//populate orderbook
	fmt.Printf("Orderbook:\n")
	for i, order := range resp.Orders {
		if (order.Side == 0) {
			fmt.Printf("Order #%d: Bid - OrderId = %s - Price = %f - Initial Quantity = %f - Remaining Quantity = %f - Timestamp = %s\n", i, order.OrderId, order.Price, order.InitQuantity, order.RemQuantity, order.Date)
			orderBook.Bids[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		} else if (order.Side == 1) {
			fmt.Printf("Order #%d: Ask - OrderId = %s - Price = %f - Initial Quantity = %f - Remaining Quantity = %f - Timestamp = %s\n", i, order.OrderId, order.Price, order.InitQuantity, order.RemQuantity, order.Date)
			orderBook.Asks[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		}
	}
	return nil
}

func cancelExcessOrders(gatecoin *api.GatecoinClient, ordersToCancel []*Order) {
	for _, order := range ordersToCancel {
		resp, err := gatecoin.DeleteOrder(order.OrderId)
		if err != nil {
			fmt.Printf("Cancelling order %s failed due to: %s\n", order.OrderId, err.Error())
		}
		if resp.Status.ErrorCode != "" || resp.Status.Message != "OK" {
			fmt.Printf("Cancelling order %s failed with error code: %s and message: %s\n", resp.Status.ErrorCode, resp.Status.Message)
		} else {
			fmt.Printf("[%s] Cancelled order %s: - Type: %d -  Price: %f - Initial Quantity: %f - Remaining Quantity: %f\n", time.Now().String(), order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, )
		}
	}
}

func topUpBands(gatecoin *api.GatecoinClient, bands *Bands, refPrice float64) {
	//create new buy and sell orders in all buy/sell bands
	topUpBuyBands(gatecoin, getBuyOrders(), bands.BuyBands, refPrice)
	topUpSellBands(gatecoin, getSellOrders(), bands.SellBands, refPrice)
}

func topUpBuyBands(gatecoin *api.GatecoinClient, orders []*Order, buyBands []BuyBand, refPrice float64) {
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
 			if buyBand.Includes(order.Price, refPrice) {
 				buyOrders = append(buyOrders, order)
 			}
 		}
 		//get totalAmount of orders
 		totalAmount := buyBand.TotalAmount(buyOrders)
 		//if total order amount is below minimum band threshold
 		if (totalAmount < buyBand.MinAmount) {
 			//get order parameters
 			price := buyBand.AvgPrice(refPrice)
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

func topUpSellBands(gatecoin *api.GatecoinClient, orders []*Order, sellBands []SellBand, refPrice float64) {
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
 			if sellBand.Includes(order.Price, refPrice) {
 				sellOrders = append(sellOrders, order)
 			}
 		}
 		//get totalAmount of orders
 		totalAmount := sellBand.TotalAmount(sellOrders)
 		//if total order amount is below minimum band threshold
 		if (totalAmount < sellBand.MinAmount) {
 			//get order parameters
 			price := sellBand.AvgPrice(refPrice)
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

func getFeedPrice(pair string) (float64, error) {
	if (strings.ToUpper(pair) == "DAIUSD") {
		return 1.00, nil
	} else if (strings.ToUpper(pair) == "ETHDAI") {
		cmd := "/Users/nkunkel/Programming/Tools/setzer/bin/setzer"
		args := []string{"gemini", "gdax", "kraken"}
		prices := []float64{}
		for _, arg := range args {
			out, err := exec.Command(cmd, "price", arg).Output()
			if err != nil {
        		fmt.Printf("%s\n", err.Error())
        		fmt.Printf("%s\n", string(out))
        		continue
    		}
    		fmt.Printf("Cmd Returned: %s : %s", arg, string(out))
    		price, err := strconv.ParseFloat(string(out[:len(out) - 2]), 64)
    		if err != nil {
    			fmt.Printf("%s", err.Error())
    			continue
    		}
    		fmt.Printf("Parsing Returned: %s : %f\n", arg, price)
    		prices = append(prices, price)
    	}
    	if (len(prices) == 0) {
    		return 0, fmt.Errorf("No valid price sources\n")
    	}
    	median := getMedian(prices)
    	fmt.Printf("Median Price = %f\n", median)
    	return median, nil
	}
	return 0, fmt.Errorf("no valid price sources\n")
}

func getMedian(prices []float64) (float64) {
	sum := 0.0
	length := len(prices)
	sort.Float64s(prices)
	for _, price := range prices[1:(length - 1)] {
		sum += price
	}
	return sum / float64(length - 2)
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

func PrintOrderBook(gatecoin *api.GatecoinClient) (error) {
	err := synchronizeOrders(gatecoin)
	if err != nil {
		return err
	}
	data := [][]string{}
	for _, order := range orderBook.Asks {
		data = append(data, []string{"Ask", order.OrderId, strconv.FormatFloat(order.Price, 'f', 6, 64), strconv.FormatFloat(order.InitQuantity, 'f', 6, 64), strconv.FormatFloat(order.RemQuantity, 'f', 6, 64), order.Date})
	}
	for _, order := range orderBook.Bids {
		data = append(data, []string{"Bid", order.OrderId, strconv.FormatFloat(order.Price, 'f', 6, 64), strconv.FormatFloat(order.InitQuantity, 'f', 6, 64), strconv.FormatFloat(order.RemQuantity, 'f', 6, 64), order.Date})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Order Type", "Order ID", "Price", "Initial Quantity", "Remaining Quantity", "Timestamp"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.AppendBulk(data)
	table.Render()
	return nil
}

func PrintAmountSold(gatecoin *api.GatecoinClient) (error) {
	//TODO
	return nil
}