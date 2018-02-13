package maker

import(
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/niklaskunkel/market-maker/api"
)

type auth struct {
	Key		string	`json:"apiKey"`
	Secret	string 	`json:"apiSecret"`
}

func setupGatecoinClient(t *testing.T) (*api.GatecoinClient) {
	credentials := new(auth)
	absPath, _ := filepath.Abs("/Users/nkunkel/Programming/Go/src/github.com/niklaskunkel/market-maker/config.json")
	raw, err := ioutil.ReadFile(absPath)
	assert.Nil(t, err)
	err = json.Unmarshal(raw, credentials)
	assert.Nil(t, err)
	client := api.NewGatecoinClient(credentials.Key, credentials.Secret)
	return client
}

func Test_Maker_SynchronizeOrders(t  *testing.T) {
	gatecoin := setupGatecoinClient(t)
	err := synchronizeOrders(gatecoin)
	assert.Nil(t, err)
}

func Test_Maker_GetFeedPrice1(t *testing.T) {
	refPrice, err := getFeedPrice("DAIUSD")
	assert.Nil(t, err)
	assert.Equal(t, refPrice, 1.0)
}

func Test_Maker_GetFeedPrice2(t *testing.T) {
	refPrice, err := getFeedPrice("ETHDAI")
	assert.Nil(t, err)
	assert.NotZero(t, refPrice)
}

func Test_Maker_GetMedian1(t *testing.T) {
	set := []float64{10.0, 100.0, 200.0, 500.0}
	median := getMedian(set)
	assert.Equal(t, median, 150.0)
}

func Test_Maker_GetMedian2(t *testing.T) {
	set := []float64{500.0, 10.0, 200.0, 100.0}
	median := getMedian(set)
	assert.Equal(t, median, 150.0)
}

//marketMaker
//cancelExcessOrders
//topUpBands
//topUpBuyBands
//topUpSellBands
//cancelAllOrders
//getOrders
//getBuyOrders
//getSellOrders
//getTotalOrderAmount
//synchronizeOrders
