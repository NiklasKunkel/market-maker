package main 

import(
	"fmt"
	"github.com/niklaskunkel/market-maker/api"
)

func main() {
	client := api.NewGatecoinClient("key", "secret")
	ticker, err := client.GetTickers()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%+v", ticker)
	return
}