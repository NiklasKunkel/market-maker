package main 

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"github.com/niklaskunkel/market-maker/api"
)

func main() {
	var CREDENTIALS auth
	ReadConfig(&CREDENTIALS)

	client := api.NewGatecoinClient(CREDENTIALS.Key, CREDENTIALS.Secret)

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
	balances, err := client.GetBalances("")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Printf("%+v", balances)

	resp, err := client.CreateOrder("BTCUSD", "bid", "1", "0.001")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("%+v\n", resp)

	order, err := client.GetOrder()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("%+v\n", order)

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