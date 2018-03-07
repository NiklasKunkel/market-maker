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

//Schedules process to execute on interval
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
	client := api.NewGatecoinClient("GATECOIN", CREDENTIALS.Key, CREDENTIALS.Secret)

	//Execute market maker on interval
	scheduler(func() {maker.MarketMaker(client, CONFIG)}, 20 * time.Second)
	return
}