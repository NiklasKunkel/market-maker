package main 

import(
	"fmt"
	"time"
	"github.com/niklaskunkel/market-maker/api"
	"github.com/niklaskunkel/market-maker/config"
	"github.com/niklaskunkel/market-maker/logger"
	"github.com/niklaskunkel/market-maker/maker"
	"github.com/sirupsen/logrus"
)

//Globals
var log *logrus.Logger

func scheduler(what func(), delay time.Duration) {
	fmt.Printf("Starting scheduled process on interval %d\n", delay)
	ticker := time.NewTicker(delay)
	quit := make(chan bool, 1)
	go func() {
		for {
	       select {
	        case <- ticker.C:
	        	what()
	        case <- quit:
	            ticker.Stop()
	            return
	        }
	    }
	 }()
	 <-quit
}

func main() {
	//Pair to Trade
	PAIR := "ETHDAI"

	//Initialize Logging
	log = logger.InitLogger()

	//Load Config
	CONFIG := new(config.Config)
	config.LoadConfig(CONFIG)

	//Load Credentials
	CREDENTIALS := new(config.Auth)
	config.LoadCredentials(CREDENTIALS)

	//Load Bands
	bands := new(maker.Bands)
	if(!bands.LoadBands()) {
		return
	}

	//Create Gatecoin API Client
	client := api.NewGatecoinClient(CREDENTIALS.Key, CREDENTIALS.Secret)
	

	//Execute market maker on interval
	scheduler(func() {maker.MarketMaker(client, bands, CONFIG, PAIR)}, 5 * time.Second)

	/*
	//TO DO - create real test scripts for these
	//Test Public Queries
	ticker, err := client.GetTickers()
	if err != nil {
		log.Error("%s", err.Error())
	}

	marketDepth, err := client.GetMarketDepth("DAIUSD")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Printf("%+v", marketDepth)

	transactions, err := client.GetTransactions("BTCUSD")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Printf("%+v", transactions)

	//Test Private Queries
	balances, err := client.GetBalances("DAI")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Printf("%+v", balances)

	//Create Order
	resp, err := client.CreateOrder("DAIUSD", "bid", "1", "0.01")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("\n%+v\n", resp)

	//Get Order
	order, err := client.GetOrders()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("\n%+v\n", order)

	//Delete Order
	cancel, err := client.DeleteOrder(resp.OrderId)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("\n%+v\n", cancel)
	*/

	return
}

/*TO DO:
Create Analytics
	timer on API calls so we can get average for latency and identify spikes, maybe use for dynamic adjusting frequency of scheduler later since higher responsiveness to volatility is lower risk
	
		API latency
		Daily Sold
		Daily Bought
		Daily Profit
		Net Sold
		Net Bought
		Net Profit

Dont have orderbook be a global and just have it be initialized in topUpBands() and then passed to synchronizeOrders and topUpBuyBands and topUpSellBands

In excessiveOrders() or in includes() need to add a check for order.Side, otherwise you will have bids which get included in sell band orders because of their price.
*/