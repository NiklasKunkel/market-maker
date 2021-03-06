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
	"github.com/niklaskunkel/market-maker/registry"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
)

//Globals
var log = logger.InitLogger()
var orderBook = make(OrderBook)

type Order struct {
	Code 			string
	OrderId 		string
	Side 			int64
	Price 			float64
	InitQuantity 	float64
	RemQuantity 	float64
	Status 			int64
	StatusDesc 		string
	TxSeqNo 		int64
	Type 			int64
	Date 			string
}

type Orders struct {
	Asks	map[string]Order
	Bids 	map[string]Order

}
type OrderBook map[string]map[string]*Orders


func MarketMaker(gatecoin *api.GatecoinClient, CONFIG *config.Config) {
	//load up Bands
	allBands := make(AllBands)
	if(!allBands.LoadBands()) {
		return
	}
	//synchronize order book
	err := SynchronizeOrders(gatecoin)
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "error": err.Error()}).Error("Failed to synchronize Orders")
		return
	}
	//iterate through active trading pairs
	for _, tokenPair := range CONFIG.ActivePairs {
		//get reference price
		refPrice, err := GetFeedPrice(tokenPair, CONFIG)
		if err != nil {
			log.WithFields(logrus.Fields{"client": "Gatecoin", "pair": tokenPair, "error": err.Error()}).Error("Failed to get feed price")
			continue
		}
		CancelExcessOrders(gatecoin, allBands[tokenPair].CancellableOrders(GetBuyOrders(tokenPair), GetSellOrders(tokenPair), refPrice))
		TopUpBands(gatecoin, tokenPair, allBands[tokenPair], refPrice)
		PrintOrderBook(gatecoin)
	}
}

//Updates the in-memory orderbook.
func SynchronizeOrders(gatecoin *api.GatecoinClient) (error) {
	//reset orderbook
	for _, quoteMap := range orderBook {
		for _, orderTypes := range quoteMap {
			orderTypes.Asks = nil
			orderTypes.Bids = nil
			orderTypes.Asks = make(map[string]Order)
			orderTypes.Bids = make(map[string]Order)
		}
	}

	log.WithFields(logrus.Fields{"client": "Gatecoin"}).Info("Synchronizing orderbook...")
	resp, err := gatecoin.GetOrders()
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "SynchronizeOrders", "error": err.Error()}).Error("Failed to synchronize orders")
		return err
	} else if (resp.Status.Message != "OK") {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "SynchronizeOrders", "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Failed to synchronize orders")
		return fmt.Errorf("Failed to synchronize orders due to invalid status message")
	}

	//populate orderbook
	for i, order := range resp.Orders {
		//initialize orderbool
		base, quote := registry.LookupTokenPair(order.Code)
		if (orderBook[base] == nil) {
			log.Debug("Initializing quote to *Orders map")
			orderBook[base] = make(map[string]*Orders)
		}
		if (orderBook[base][quote] == nil) {
			orderBook[base][quote] = &Orders{Asks: make(map[string]Order), Bids: make(map[string]Order)}
		}

		//divide orders into asks and bids
		if (order.Side == 0) {
			log.WithFields(logrus.Fields{"orderNum": i, "pair": order.Code, "orderId": order.OrderId, "type": "bid", "price": order.Price, "initialQuantity": order.InitQuantity, "remainingQuantity": order.RemQuantity, "timestamp": order.Date}).Debug()
			//insert into orderbook
			(orderBook[base])[quote].Bids[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		} else if (order.Side == 1) {
			log.WithFields(logrus.Fields{"orderNum": i, "pair": order.Code, "orderId": order.OrderId, "type": "ask", "price": order.Price, "initialQuantity": order.InitQuantity, "remainingQuantity": order.RemQuantity, "timestamp": order.Date}).Debug()
			//insert into orderbook
			(orderBook[base])[quote].Asks[order.OrderId] = Order{order.Code, order.OrderId, order.Side, order.Price, order.InitQuantity, order.RemQuantity, order.Status, order.StatusDesc, order.TxSeqNo, order.Type, order.Date}
		}
	}
	return nil
}

func CancelExcessOrders(gatecoin *api.GatecoinClient, ordersToCancel []*Order) {
	for _, order := range ordersToCancel {
		resp, err := gatecoin.DeleteOrder(order.OrderId)
		if err != nil {
			log.WithFields(logrus.Fields{"function": "CancelExcessOrders", "orderId": order.OrderId, "error": err.Error()}).Error("Cancelling order failed")
		}
		if resp.Status.ErrorCode != "" || resp.Status.Message != "OK" {
			log.WithFields(logrus.Fields{"function": "CancelExcessOrders", "orderId": order.OrderId, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Cancelling order failed")
		} else {
			log.WithFields(logrus.Fields{"orderId": order.OrderId, "type": order.Side, "price": order.Price, "initialQuantity": order.InitQuantity, "remainingQuantity": order.RemQuantity}).Info("Cancelled Order")
			//remove order from internal orderbook
			base, quote := registry.LookupTokenPair(order.Code)
			if (order.Side == 0) {
				delete(orderBook[base][quote].Bids, order.OrderId)
			} else if (order.Side == 1) {
				delete(orderBook[base][quote].Bids, order.OrderId) 
			}
		}
	}
}

func TopUpBands(gatecoin *api.GatecoinClient, tokenPair string, bands Bands, refPrice float64) {
	//create new buy and sell orders in all buy/sell bands
	TopUpBuyBands(gatecoin, tokenPair, GetBuyOrders(tokenPair), bands.BuyBands, refPrice)
	TopUpSellBands(gatecoin, tokenPair, GetSellOrders(tokenPair), bands.SellBands, refPrice)
}

func TopUpBuyBands(client *api.GatecoinClient, tokenPair string, orders []*Order, buyBands []BuyBand, refPrice float64) {
	//lookup token pair components
	_, quote := registry.LookupTokenPair(tokenPair)
	//get balance of quote token
 	availableBalance, err := client.GetBalance(quote)
 	if err != nil {
 		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "TopUpBuyBands", "token": quote, "error": err.Error()}).Error("Failed to get balances")
 		return
	}
	availableQuoteBalance := availableBalance.Balance.AvailableBalance
	precision := registry.LookupGatecoinTokenPairPrecision(tokenPair)

 	inBandBuyOrders := []*Order{}
 	//iterate through buy bands 
 	for _, buyBand := range buyBands {
 		//iterate through all buy orders for tokenPair
 		for _, order := range orders {
 			//check if buy order is included in band
 			if buyBand.Includes(order.Price, refPrice) {
 				//add to in-band buy order list
 				inBandBuyOrders = append(inBandBuyOrders, order)
 			}
 		}
 		//get total amount of all buy orders in band denominated in quote token except for gatecoin which uses base token
 		totalAmount := buyBand.TotalAmount(inBandBuyOrders)
 		//special case code for Gatecoin b/c Gatecoin uses base token for amounts of bids and asks
 		if (client.Name == "GATECOIN") {
 			//price denominated in quote / base
	 		price := buyBand.AvgPrice(refPrice)
	 		//convert total amount to be denominated in quote token using average price of band (approximation b/c gatecoin sucks)
	 		totalAmount = price * totalAmount
	 		//if total order amount is below minimum band threshold
	 		if (totalAmount < buyBand.MinAmount) {
	 			//get order parameters
	 			//amount to pay denominated in quote token
	 			payAmount := math.Min(buyBand.AvgAmount - totalAmount, availableQuoteBalance)
	 			//amount to buy denominated in base token
	 			buyAmount := payAmount / price
	 			//verify order parameters
	 			if ((payAmount >= buyBand.DustCutoff) && (payAmount > float64(0)) && (buyAmount > float64(0))) {
	 				//lookup Gatecoin token pair syntax
	 				gatecoinTokenPair := registry.LookupGatecoinTokenPairName(tokenPair)
	 				//adjust amount and price with precision limits for each exchange
	 				adjustedAmount := strconv.FormatFloat(buyAmount, 'f', precision.BIDAMOUNTPRECISION, 64)
	 				adjustedPrice := strconv.FormatFloat(price, 'f', precision.BIDPRICEPRECISION, 64)
	 				potentialRemainingQuoteBalance := availableQuoteBalance - payAmount
	 				//log attempted order creation
	 				log.WithFields(logrus.Fields{"client": "Gatecoin", "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "potentialRemainingQuoteBalance": potentialRemainingQuoteBalance}).Info("Creating buy order...")
	 				//create order - amount denominated in base token
	 				resp, err := client.CreateOrder(gatecoinTokenPair, "bid", adjustedAmount, adjustedPrice)
	 				//check if order creation failed
	 				if err != nil {
	 					log.WithFields(logrus.Fields{"client": "Gatecoin", "error": err.Error(), "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "potentialRemainingQuoteBalance": potentialRemainingQuoteBalance}).Error("Creating buy order failed")
	 					continue
	 				} else if resp.Status.Message != "OK" || resp.OrderId == "" {
	 					log.WithFields(logrus.Fields{"client": "Gatecoin", "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode, "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "potentialRemainingBalance": potentialRemainingQuoteBalance}).Error("Creating buy order failed")
	 					continue
	 				}
	 				availableQuoteBalance = potentialRemainingQuoteBalance
	 				//log successful order creation
	 				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": resp.OrderId, "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "remainingQuoteBalance": availableQuoteBalance}).Info("Created buy order")
	 			}
	 		}
	 		inBandBuyOrders = nil
	 	}
 	}
 	return
}

func TopUpSellBands(gatecoin *api.GatecoinClient, tokenPair string, orders []*Order, sellBands []SellBand, refPrice float64) {
	//lookup token pair components
	base, _ := registry.LookupTokenPair(tokenPair)
	//get balance of base token
	availableBalance, err := gatecoin.GetBalance(base)
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "TopUpBuyBands", "error": err.Error()}).Error("Failed to get balances")
		return
	}
	availableBaseBalance := availableBalance.Balance.AvailableBalance
	precision := registry.LookupGatecoinTokenPairPrecision(tokenPair)

	inBandSellOrders := []*Order{}
 	//iterate through sell bands 
 	for _, sellBand := range sellBands {
 		//iterate through all sell orders
 		for _, order := range orders {
 			//check if sell order is included in band 
 			if sellBand.Includes(order.Price, refPrice) {
 				//add to in-band sell order list
 				inBandSellOrders = append(inBandSellOrders, order)
 			}
 		}
 		//get total amount of all sell orders in band
 		totalAmount := sellBand.TotalAmount(inBandSellOrders)
 		//if total order amount is below minimum band threshold
 		if (totalAmount < sellBand.MinAmount) {
 			//get order parameters
 			//price denominated in quote / base
 			price := sellBand.AvgPrice(refPrice)
 			//amount to pay denominated in base token
 			payAmount := math.Min(sellBand.AvgAmount - totalAmount, availableBaseBalance)
 			//amount to buy denominated in quote token
 			buyAmount := payAmount * price
 			//verify order parameters
 			if ((payAmount >= sellBand.DustCutoff) && (payAmount > float64(0)) && (buyAmount > float64(0))) {
 				//lookup Gatecoin token pair syntax
 				gatecoinTokenPair := registry.LookupGatecoinTokenPairName(tokenPair)
 				//adjust order amount for precision allowed by Gatecoin API
 				adjustedAmount := strconv.FormatFloat(payAmount, 'f', precision.ASKAMOUNTPRECISION, 64)
 				//adjust order price for precision allowed by Gatecoin API
 				adjustedPrice := strconv.FormatFloat(price, 'f', precision.ASKPRICEPRECISION, 64)
 				//Log order creation
 				log.WithFields(logrus.Fields{"client": "Gatecoin", "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "potentialRemainingBalance": availableBaseBalance - payAmount}).Info("Creating sell order...")
 				//create order - amount denominated in base token
 				resp, err := gatecoin.CreateOrder(gatecoinTokenPair, "ask", adjustedAmount, adjustedPrice)
 				fmt.Printf("%+v\n", resp)
 				if err != nil {
 					log.WithFields(logrus.Fields{"client": "Gatecoin", "error": err.Error(), "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "potentialRemainingBalance": availableBaseBalance - payAmount}).Error("Creating sell order failed")
 					continue
 				} else if resp.Status.Message != "OK" || resp.OrderId == "" {
 					log.WithFields(logrus.Fields{"client": "Gatecoin", "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode, "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "potentialRemainingBalance": availableBaseBalance - payAmount}).Error("Creating buy order failed")
 					continue
 				}
 				availableBaseBalance -= payAmount
 				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": resp.OrderId, "pair": gatecoinTokenPair, "amount": adjustedAmount, "price": adjustedPrice, "remainingBalance": availableBaseBalance - payAmount}).Info("Created sell order")
 			}
 		}
 		inBandSellOrders = nil
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
    		price, err := strconv.ParseFloat(string(out[:len(out) - 1]), 64)
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
	} else if (strings.ToUpper(pair) == "MKRBTC") {
		return 0, fmt.Errorf("No valid price sources")
	} else if (strings.ToUpper(pair) == "MKRETH") {
		return 0, fmt.Errorf("No valid price sources")
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
	for _, quoteSet := range orderBook {
		for _, orders := range quoteSet {
			for id, _ := range orders.Bids {
				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": id}).Info("Cancelling order...")
				resp, err := gatecoin.DeleteOrder(id)
				if err != nil {
					log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "error": err.Error()}).Error("Failed to cancel order")
				} else if resp.Status.Message != "OK" {
					log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Failed to cancel order")
				}
				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": id}).Info("Cancelled Order")
			}
			for id, _ := range orders.Asks {
				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": id}).Info("Cancelling order...")
				resp, err := gatecoin.DeleteOrder(id)
				if err != nil {
					log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "error": err.Error()}).Error("Failed to cancel order")
				} else if resp.Status.Message != "OK" {
					log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Failed to cancel order")
				}
				log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": id}).Info("Cancelled Order")
			}
		}
	}
}

func CancelTokenPairOrders(gatecoin *api.GatecoinClient, pair string) {
	base, quote := registry.LookupTokenPair(pair)
	SynchronizeOrders(gatecoin)
	//Check if token pair exists in orderbook
	if _, ok := orderBook[base]; !ok {
    	return
	}
	if _, ok := orderBook[base][quote]; !ok {
    	return
	}
	//get orders for token pair
	orders := orderBook[base][quote]
	//iterate over buy orders
	for id, _ := range orders.Bids {
		//cancel buy order
		log.WithFields(logrus.Fields{"client": "Gatecoin", "orderId": id}).Info("Cancelling order...")
		resp, err := gatecoin.DeleteOrder(id)
		if err != nil {
			log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "error": err.Error()}).Error("Failed to cancel order")
		} else if resp.Status.Message != "OK" {
			log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Error("Failed to cancel order")
		}
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "CancelAllOrders", "orderId": id, "message": resp.Status.Message, "errorCode": resp.Status.ErrorCode}).Info("Cancelled Order")
	}
	//iterate over sell orders
	for id, _ := range orders.Asks {
		//cancel sell order
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

func GetBuyOrders(tokenPair string) (bids []*Order) {
	base, quote := registry.LookupTokenPair(tokenPair)
	if _, ok := orderBook[base]; !ok {
    	return bids
	}
	if _, ok := orderBook[base][quote]; !ok {
    	return bids
	}
	for _, bid := range orderBook[base][quote].Bids {
		log.WithFields(logrus.Fields{"function": "GetBuyOrders", "bid": bid}).Debug("Pulling bid from orderbook")
		//bids = append(bids, &bid)
		bids = append(bids, &Order{bid.Code, bid.OrderId, bid.Side, bid.Price, bid.InitQuantity, bid.RemQuantity, bid.Status, bid.StatusDesc, bid.TxSeqNo, bid.Type, bid.Date})
	}
	for _, bid := range bids {
		log.WithFields(logrus.Fields{"function": "GetBuyOrders", "bid": bid}).Debug("List of bids compiled from orderbook")
	}
	return bids
}

func GetSellOrders(tokenPair string) (asks []*Order) {
	base, quote := registry.LookupTokenPair(tokenPair)
	if _, ok := orderBook[base]; !ok {
    	return asks
	}
	if _, ok := orderBook[base][quote]; !ok {
    	return asks
	}
	for _, ask := range orderBook[base][quote].Asks {
		log.WithFields(logrus.Fields{"function": "GetSellOrders", "ask": ask}).Debug("Pulling ask from orderbook")
		//asks = append(asks, &ask)
		asks = append(asks, &Order{ask.Code, ask.OrderId, ask.Side, ask.Price, ask.InitQuantity, ask.RemQuantity, ask.Status, ask.StatusDesc, ask.TxSeqNo, ask.Type, ask.Date})

	}
	for _, ask := range asks {
		log.WithFields(logrus.Fields{"function": "GetBuyOrders", "ask": ask}).Debug("List of asks compiled from orderbook")
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
	for _, quoteSet := range orderBook {
		for _, orders := range quoteSet {
			data := [][]string{}
			for _, order := range orders.Asks {
				data = append(data, []string{order.Code, "Ask", order.OrderId, strconv.FormatFloat(order.Price, 'f', 6, 64), strconv.FormatFloat(order.InitQuantity, 'f', 6, 64), strconv.FormatFloat(order.RemQuantity, 'f', 6, 64), order.Date})
			}
			for _, order := range orders.Bids {
				data = append(data, []string{order.Code, "Bid", order.OrderId, strconv.FormatFloat(order.Price, 'f', 6, 64), strconv.FormatFloat(order.InitQuantity, 'f', 6, 64), strconv.FormatFloat(order.RemQuantity, 'f', 6, 64), order.Date})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Pair", "Order Type", "Order ID", "Price", "Initial Quantity", "Remaining Quantity", "Timestamp"})
			table.SetHeaderColor(
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold})
			table.AppendBulk(data)
			table.Render()
		} 
	}
	return nil
}

func PrintBalances(gatecoin *api.GatecoinClient) (error) {
	resp, err := gatecoin.GetBalances()
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "error": err}).Error("Failed to query token balances")
		return err
	}
	data := [][]string{}
	
	for _, balance := range resp.Balances {
		data = append(data, []string{balance.Currency, strconv.FormatFloat(balance.Balance, 'f', 6, 64), strconv.FormatFloat(balance.AvailableBalance, 'f', 6, 64), strconv.FormatFloat(balance.PendingIncoming, 'f', 6, 64), strconv.FormatFloat(balance.PendingOutgoing, 'f', 6, 64), strconv.FormatFloat(balance.OpenOrder, 'f', 6, 64)})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Currency", "Balance", "AvailableBalance", "PendingBalance", "PendingOutgoing", "OpenOrder"})
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