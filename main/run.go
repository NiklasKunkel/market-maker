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
		fmt.Printf("%s", err)
	}
	fmt.Printf("%+v", ticker)

	marketDepth, err := client.GetMarketDepth("DAIUSD")
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%+v", marketDepth)

	transactions, err := client.GetTransactions("DAIUSD")
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%+v", transactions)

	//Test Private Queries
	balances, err := client.GetBalances("DAI")
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%+v", balances)

	//



	return
}

type auth struct {
	Key		string	`json:"apiKey"`
	Secret	string 	`json:"secretKey"`
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
	return
}