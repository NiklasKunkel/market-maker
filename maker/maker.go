package maker

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"github.com/niklaskunkel/market-maker/api"
	"github.com/niklaskunkel/market-maker/config"
	"github.com/niklaskunkel/market-maker/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
)

//Globals
var log = logger.InitLogger()
var orderBook = new(Orders)

type Orders struct {
	Asks	map[string]Order
	Bids 	map[string]Order
}

func MarketMaker(gatecoin *api.GatecoinClient, bands *Bands, CONFIG *config.Config, pair string) {
	//synchronize order book
	err := SynchronizeOrders(gatecoin)
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "error": err.Error()}).Error("Failed to synchronize Orders")
		return
	}
	//get reference price
	refPrice, err := GetFeedPrice(pair, CONFIG)
	if err != nil {
		log.WithFields(logrus.Fields{"pair": pair, "error": err.Error()}).Error("Failed to get feed price")
		return
	}
	CancelExcessOrders(gatecoin, bands.CancellableOrders(GetBuyOrders(), GetSellOrders(), refPrice))
	TopUpBands(gatecoin, bands, refPrice)
	PrintOrderBook(gatecoin)
}

//Updates the in-memory orderbook.
func SynchronizeOrders(gatecoin *api.GatecoinClient) (error) {
	fmt.Printf("synchronizing orderbook...\n")
	resp, err := gatecoin.GetOrders()
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "synchronizeOrders", "error": err.Error()}).Error("Failed to synchronize orders")
		return err
	} else if (resp.Status.Message != "OK") {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "synchronizeOrders", "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Failed to synchronize orders")
		return fmt.Errorf("Failed to synchronize orders due to status message")
	}
	//reset orderbook
	orderBook.Asks = nil
	orderBook.Bids = nil
	orderBook.Asks = make(map[string]Order)
	orderBook.Bids = make(map[string]Order)

	//populate orderbook
	log.Debug("Orderbook:")
	for i, order := range resp.Orders {
		if (order.Side == 0) {
			log.WithFields(logrus.Fields{"orderNum": i, "orderId": order.OrderId, "type": "bid", "price": order.Price, "initialQuantity": order.InitQuantity, "remainingQuantity": order.RemQuantity, "timestamp": order.Date}).Debug()
			orderBook.Bids[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		} else if (order.Side == 1) {
			log.WithFields(logrus.Fields{"orderNum": i, "orderId": order.OrderId, "type": "ask", "price": order.Price, "initialQuantity": order.InitQuantity, "remainingQuantity": order.RemQuantity, "timestamp": order.Date}).Debug()
			orderBook.Asks[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		}
	}
	return nil
}

func CancelExcessOrders(gatecoin *api.GatecoinClient, ordersToCancel []*Order) {
	for _, order := range ordersToCancel {
		resp, err := gatecoin.DeleteOrder(order.OrderId)
		if err != nil {
			log.WithFields(logrus.Fields{"function": "cancelExcessOrders", "orderId": order.OrderId, "error": err.Error()}).Error("Cancelling order failed")
		}
		if resp.Status.ErrorCode != "" || resp.Status.Message != "OK" {
			log.WithFields(logrus.Fields{"function": "cancelExcessOrders", "orderId": order.OrderId, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Cancelling order failed")
		} else {
			log.WithFields(logrus.Fields{"orderId": order.OrderId, "type": order.Side, "price": order.Price, "initialQuantity": order.InitQuantity, "remainingQuantity": order.RemQuantity}).Info("Cancelled Order")
		}
	}
}

func TopUpBands(gatecoin *api.GatecoinClient, bands *Bands, refPrice float64) {
	//create new buy and sell orders in all buy/sell bands
	TopUpBuyBands(gatecoin, GetBuyOrders(), bands.BuyBands, refPrice)
	TopUpSellBands(gatecoin, GetSellOrders(), bands.SellBands, refPrice)
}

func TopUpBuyBands(gatecoin *api.GatecoinClient, orders []*Order, buyBands []BuyBand, refPrice float64) {
	//get balance
 	availableBalances, err := gatecoin.GetBalances("USD")
 	if err != nil {
 		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "topUpBuyBands", "error": err.Error()}).Error("Failed to get balances")
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
 				log.WithFields(logrus.Fields{"client": "Gatecoin", "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableUsdBalance - payAmount}).Info("Creating buy order...")
 				//create order
 				resp, err := gatecoin.CreateOrder("DAIUSD", "bid", strconv.FormatFloat(buyAmount, 'f', 6, 64), strconv.FormatFloat(price, 'f', 6, 64))	//not sure if this is payAmount or buyAmount
 				if err != nil {
 					log.WithFields(logrus.Fields{"client": "Gatecoin", "error": err.Error(), "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableUsdBalance - payAmount}).Error("Creating buy order failed")
 					continue
 				} else if resp.Status.Message != "OK" || resp.OrderId == "" {
 					log.WithFields(logrus.Fields{"client": "Gatecoin", "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode, "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableUsdBalance - payAmount}).Error("Creating buy order failed")
 					continue
 				}
 				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": resp.OrderId, "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableUsdBalance - payAmount}).Info("Created buy order")
 			}
 		}
 		buyOrders = nil
 	}
 	return
}

func TopUpSellBands(gatecoin *api.GatecoinClient, orders []*Order, sellBands []SellBand, refPrice float64) {
	availableBalances, err := gatecoin.GetBalances("DAI")
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "topUpBuyBands", "error": err.Error()}).Error("Failed to get balances")
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
 				log.WithFields(logrus.Fields{"client": "Gatecoin", "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableDaiBalance - payAmount}).Info("Creating sell order...")
 				//create order
 				resp, err := gatecoin.CreateOrder("DAIUSD", "ask", strconv.FormatFloat(buyAmount, 'f', 6, 64), strconv.FormatFloat(price, 'f', 6, 64))	//not sure if this is payAmount or buyAmount
 				fmt.Printf("%+v\n", resp)
 				if err != nil {
 					log.WithFields(logrus.Fields{"client": "Gatecoin", "error": err.Error(), "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableDaiBalance - payAmount}).Error("Creating sell order failed")
 					continue
 				} else if resp.Status.Message != "OK" || resp.OrderId == "" {
 					log.WithFields(logrus.Fields{"client": "Gatecoin", "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode, "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableDaiBalance - payAmount}).Error("Creating buy order failed")
 					continue
 				}
 				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": resp.OrderId, "pair": "DAIUSD", "amount": buyAmount, "price": price, "remainingBalance": availableDaiBalance - payAmount}).Info("Created sell order")
 			}
 		}
 		sellOrders = nil
 	}
 	return
}

func GetFeedPrice(pair string, config *config.Config) (float64, error) {
	if (strings.ToUpper(pair) == "DAIUSD") {
		return 1.00, nil
	} else if (strings.ToUpper(pair) == "ETHDAI") {
		cmd := config.SetzerPath
		exchanges := []string{"gemini", "gdax", "kraken"}
		prices := []float64{}
		for _, exchange := range exchanges {
			out, err := exec.Command(cmd, "price", exchange).Output()
			if err != nil {
				log.WithFields(logrus.Fields{"function": "GetFeedPrice", "exchange": exchange, "pair": pair, "error": err.Error(), "output": string(out)}).Error("Setzer failed to fetch price")
        		continue
    		}
    		log.WithFields(logrus.Fields{"function": "GetFeedPrice", "exchange": exchange, "pair": pair, "price": string(out)}).Debug("Setzer price fetch returned price")
    		price, err := strconv.ParseFloat(string(out[:len(out) - 2]), 64)
    		if err != nil {
    			log.WithFields(logrus.Fields{"function": "GetFeedPrice", "exchange": exchange, "pair": pair, "price": string(out), "error": err.Error()}).Error("Failed to parse price from string to float")
    			continue
    		}
    		log.WithFields(logrus.Fields{"function": "GetFeedPrice", "exchange": exchange, "pair": pair, "unparsedPrice": string(out), "parsedPrice": price}).Debug("Parsing setzer price returned")
    		prices = append(prices, price)
    	}
    	if (len(prices) == 0) {
    		log.WithFields(logrus.Fields{"function": "GetFeedPrice", "pair": pair}).Error("No valid price sources")
    		return 0, fmt.Errorf("No valid price sources")
    	}
    	median := GetMedian(prices)
    	log.WithFields(logrus.Fields{"function": "GetFeedPrice", "prices": prices, "median": median}).Debug("Median price of feed prices is...")
    	return median, nil
	}
	return 0, fmt.Errorf("no valid price sources\n")
}

func GetMedian(prices []float64) (float64) {
	sum := 0.0
	length := len(prices)
	sort.Float64s(prices)
	for _, price := range prices[1:(length - 1)] {
		sum += price
	}
	return sum / float64(length - 2)
}

func CancelAllOrders(gatecoin *api.GatecoinClient) {
	SynchronizeOrders(gatecoin)
	log.WithFields(logrus.Fields{"client": "Gatecoin"}).Info("Cancelling all orders...")
	for id, _ := range orderBook.Bids {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": id}).Info("Cancelling order...")
		resp, err := gatecoin.DeleteOrder(id)
		if err != nil {
			log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "error": err.Error()}).Error("Failed to cancel order")
		} else if resp.Status.Message != "OK" {
			log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Failed to cancel order")
		}
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Info("Cancelled Order")
	}
	for id, _ := range orderBook.Asks {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": id}).Info("Cancelling order...")
		resp, err := gatecoin.DeleteOrder(id)
		if err != nil {
			log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "error": err.Error()}).Error("Failed to cancel order")
		} else if resp.Status.Message != "OK" {
			log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Failed to cancel order")
		}
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Info("Cancelled Order")
	}
}

func GetOrders() (*Orders) {
	return orderBook
}

func GetBuyOrders() (bids []*Order) {
	for _, bid := range orderBook.Bids {
		bids = append(bids, &bid)
	}
	return bids
}

func GetSellOrders() (asks []*Order) {
	for _, ask := range orderBook.Asks {
		asks = append(asks, &ask)
	}
	return asks
}

func GetTotalOrderAmount(orders []*Order) (sum float64) {
	for _, order := range orders {
		sum += order.RemQuantity
	}
	return sum
}

func PrintOrderBook(gatecoin *api.GatecoinClient) (error) {
	err := SynchronizeOrders(gatecoin)
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