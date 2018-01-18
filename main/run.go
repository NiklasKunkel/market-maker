package main 

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"github.com/niklaskunkel/market-maker/api"
	"github.com/niklaskunkel/market-maker/maker"
)

func main() {
	//Load Credentials
	CREDENTIALS := new(auth)
	ReadConfig(CREDENTIALS)

	//Create Gatecoin API Client
	client := api.NewGatecoinClient(CREDENTIALS.Key, CREDENTIALS.Secret)
	
	//Load Bands
	BANDS := new(maker.Bands)
	err := BANDS.LoadBands()
	if err != nil {
		fmt.Printf("Error loading bands: %s\n", err.Error())
	}
	//Print Bands
	BANDS.PrintBands()
	//Verify Bands
	err = BANDS.VerifyBands()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

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