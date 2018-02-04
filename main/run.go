package main 

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
	"github.com/niklaskunkel/market-maker/api"
	"github.com/niklaskunkel/market-maker/maker"
)

func main() {
	//Pair to Trade
	PAIR := "ETHDAI"

	//Load Credentials
	CREDENTIALS := new(auth)
	ReadConfig(CREDENTIALS)

	//Create Gatecoin API Client
	client := api.NewGatecoinClient(CREDENTIALS.Key, CREDENTIALS.Secret)
	
	//Load Bands
	bands := new(maker.Bands)
	err := bands.LoadBands()
	if err != nil {
		fmt.Printf("Error loading bands: %s\n", err.Error())
	}
	//Print Bands
	bands.PrintBands()
	//Verify Bands
	err = bands.VerifyBands()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	//Execute market maker on interval
	scheduler(func() {maker.MarketMaker(client, bands, PAIR)}, 5 * time.Second)

	//TO DO - create real test scripts for these
	/*
	//Test Public Queries
	ticker, err := client.GetTickers()
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Printf("%+v", ticker)

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

type auth struct {
	Key		string	`json:"apiKey"`
	Secret	string 	`json:"apiSecret"`
}

func ReadConfig(credentials *auth) {
	absPath, _ := filepath.Abs("./src/github.com/niklaskunkel/market-maker/config.json")
	raw, err := ioutil.ReadFile(absPath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(raw, credentials)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("api key = %s\napi secret = %s\n", credentials.Key, credentials.Secret)
	return
}

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

/*NOTES
Create Analytics
	timer on API calls so we can get average for latency and identify spikes, maybe use for dynamic adjusting frequency of scheduler later
	internal tracking of order amounts sold and bought

Split output into multiple logs
	action log - track all actions and timestamp in a log - order made - order cancelled - timestamped
		These should be in easy to follow format, probably JSON of GetOrder(id)
		JSON format would help for parsing to create analytics later
	raw logs - shows all raw printed output from intermediate function - useful for debugging production failures
	analytics log - interim solution before creating database module for storing analytics
		API latency
		Daily Sold
		Daily Bought
		Daily Profit
		Net Sold
		Net Bought
		Net Profit

Dont have orderbook be a global and just have it be initialized in tupUpBands() and then passed to synchronizeOrders and topUpBuyBands and topUpSellBands

In excessiveOrders() or in includes() need to add a check for order.Side, otherwise you will have bids which get included in sell band orders because of their price.
*/