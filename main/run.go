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
	//Initialize Logging
	log = logger.InitLogger()

	//Load Config
	CONFIG := new(config.Config)
	config.LoadConfig(CONFIG)

	//Load Credentials
	CREDENTIALS := new(config.Auth)
	config.LoadCredentials(CREDENTIALS)

	//Create Gatecoin API Client
	client := api.NewGatecoinClient(CREDENTIALS.Key, CREDENTIALS.Secret)

	//Execute market maker on interval
	scheduler(func() {maker.MarketMaker(client, CONFIG)}, 5 * time.Second)

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