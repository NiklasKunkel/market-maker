package api

import(
	"io/ioutil"
	"encoding/json"
	"os"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/niklaskunkel/market-maker/config"
)

//CONSTANTS
const API_PUBLIC_RATE_LIMIT = 1000 * time.Millisecond
const API_PRIVATE_RATE_LIMIT = 1000 * time.Millisecond

func SetupGatecoinClient(t *testing.T) (*GatecoinClient) {
	credentials := new(config.Auth)
	goPath, ok := os.LookupEnv("GOPATH")
	assert.True(t, ok)
	credentialsPath := goPath + "/src/github.com/niklaskunkel/market-maker/credentials.json"
	raw, err := ioutil.ReadFile(credentialsPath)
	assert.Nil(t, err)
	err = json.Unmarshal(raw, credentials)
	assert.Nil(t, err)
	client := NewGatecoinClient(credentials.Key, credentials.Secret)
	return client
}

func Test_Api_GetTickers(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PUBLIC_RATE_LIMIT)
	resp, err := gatecoin.GetTickers()
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	assert.NotEmpty(t, resp.Tickers)
	//many token pairs have 0 volume and price data so we check only one token pair likely to have high volume
	ticker := resp.Tickers[0]
	assert.NotEqual(t, ticker.Pair, "")
	assert.NotZero(t, ticker.Open)
	assert.NotZero(t, ticker.Last)
	assert.NotZero(t, ticker.LastQ)
	assert.NotZero(t, ticker.High)
	assert.NotZero(t, ticker.Low)
	assert.NotZero(t, ticker.Volume)
	assert.NotZero(t, ticker.Volumn)
	assert.NotZero(t, ticker.Bid)
	assert.NotZero(t, ticker.BidQ)
	assert.NotZero(t, ticker.Ask)
	assert.NotZero(t, ticker.AskQ)
	assert.NotZero(t, ticker.Vwap)
	assert.NotEqual(t, ticker.Last, "")
}

func Test_Api_GetMarketDepth(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PUBLIC_RATE_LIMIT)
	resp, err := gatecoin.GetMarketDepth("BTCUSD")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	assert.NotEmpty(t, resp.Asks)
	assert.NotEmpty(t, resp.Bids)
	for _, ask := range resp.Asks {
		assert.NotZero(t, ask.Price)
		assert.NotZero(t, ask.Volume)
	}
	for _, bid := range resp.Bids {
		assert.NotZero(t, bid.Price)
		assert.NotZero(t, bid.Volume)
	}
}

func Test_Api_GetTransactions(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PUBLIC_RATE_LIMIT)
	resp, err := gatecoin.GetTransactions("BTCUSD")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	assert.NotEmpty(t, resp.Transactions)
	for _, tx := range resp.Transactions {
		assert.NotZero(t, tx.Id)
		assert.NotEqual(t, tx.Time, "")
		assert.NotZero(t, tx.Price)
		assert.NotZero(t, tx.Quantity)
		assert.NotEqual(t, tx.Pair, "")
		//assert.NotEqual(t, tx.Way, "")	//some orders do not have these fields
		//assert.NotEqual(t, tx.AskId, "")	//some orders do not have these fields
		//assert.NotEqual(t, tx.BidId, "")	//some orders do not have these fields
	}
}

func Test_Api_GetBalances(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PRIVATE_RATE_LIMIT)
	resp, err := gatecoin.GetBalances()
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	for _, token := range resp.Balances {
		assert.NotEqual(t, token.Currency, "")
		//assert.NotZero(t, resp.Balance)	//most of our balances are zero
	}
}

func Test_Api_GetBalance(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PRIVATE_RATE_LIMIT)
	resp, err := gatecoin.GetBalance("DAI")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	assert.NotEqual(t, resp.Balance.Currency, "")
	assert.NotZero(t, resp.Balance.Balance)
	assert.True(t, resp.Balance.IsDigital)
}

var OrderId = ""

func Test_Api_CreateOrder(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PRIVATE_RATE_LIMIT)
	resp, err := gatecoin.CreateOrder("DAIUSD", "bid", "1", "0.01")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	assert.NotEqual(t, resp.OrderId, "")
	OrderId = resp.OrderId
}

func Test_Api_GetOrders(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PRIVATE_RATE_LIMIT)
	resp, err := gatecoin.GetOrders()
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	for _, order := range resp.Orders {
		assert.NotEqual(t, order.Code, "")
		assert.NotEqual(t, order,OrderId, "")
		//assert.NotZero(t, order.Side)		//0 is valid response, so no way to check from default val of 0
		assert.NotZero(t, order.Price)
		assert.NotZero(t, order.InitQuantity)
		assert.NotZero(t, order.RemQuantity)
		assert.NotZero(t, order.Status)
		assert.NotEqual(t, order.StatusDesc, "")
		//assert.NotZero(t, order.TxSeqNo)	//0 is valid response, so no way to check from default val of 0
		assert.NotEqual(t, order.Date, "")
	}
}

func Test_Api_GetOrder(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PRIVATE_RATE_LIMIT)
	resp, err := gatecoin.GetOrder(OrderId)
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
	order := resp.Order
	assert.NotEqual(t, order.Code, "")
	assert.NotEqual(t, order,OrderId, "")
	//assert.NotZero(t, order.Side)			//0 is valid response, so no way to check from default val of 0
	assert.NotZero(t, order.Price)
	assert.NotZero(t, order.InitQuantity)
	assert.NotZero(t, order.RemQuantity)
	assert.NotZero(t, order.Status)
	assert.NotEqual(t, order.StatusDesc, "")
	//assert.NotZero(t, order.TxSeqNo)		//0 is valid response, so no way to check from default val of 0
	assert.NotEqual(t, order.Date, "")
}

func Test_Api_DeleteOrder(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(API_PRIVATE_RATE_LIMIT)
	resp, err := gatecoin.DeleteOrder(OrderId)
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp.Status.Message)
}