package maker

import(
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/niklaskunkel/market-maker/api"
	"github.com/niklaskunkel/market-maker/config"
)

func SetupGatecoinClient(t *testing.T) (*api.GatecoinClient) {
	credentials := new(config.Auth)
	goPath, ok := os.LookupEnv("GOPATH")
	assert.True(t, ok)
	credentialsPath := goPath + "/src/github.com/niklaskunkel/market-maker/credentials.json"
	raw, err := ioutil.ReadFile(credentialsPath)
	assert.Nil(t, err)
	err = json.Unmarshal(raw, credentials)
	assert.Nil(t, err)
	client := api.NewGatecoinClient("GATECOIN", credentials.Key, credentials.Secret)
	return client
}

func Test_Maker_SynchronizeOrders1(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	err := SynchronizeOrders(gatecoin)
	assert.Nil(t, err)
}

func Test_Maker_SynchronizeOrders2(t *testing.T) {
	gatecoin := SetupGatecoinClient(t)
	time.Sleep(1000 * time.Millisecond)
	err := SynchronizeOrders(gatecoin)
	assert.Nil(t, err)
	time.Sleep(1000 * time.Millisecond)
	err = SynchronizeOrders(gatecoin)
	assert.Nil(t, err)
}

func Test_Maker_GetFeedPrice1(t *testing.T) {
	configuration := new(config.Config)
	config.LoadConfig(configuration)
	refPrice, err := GetFeedPrice("DAIUSD", configuration)
	assert.Nil(t, err)
	assert.Equal(t, refPrice, 1.0)
}

func Test_Maker_GetFeedPrice2(t *testing.T) {
	configuration := new(config.Config)
	config.LoadConfig(configuration)
	refPrice, err := GetFeedPrice("ETHDAI", configuration)
	assert.Nil(t, err)
	assert.NotZero(t, refPrice)
}

func Test_Maker_GetMedian1(t *testing.T) {
	set := []float64{10.0, 100.0, 200.0, 500.0}
	median := GetMedian(set)
	assert.Equal(t, median, 150.0)
}

func Test_Maker_GetMedian2(t *testing.T) {
	set := []float64{500.0, 10.0, 200.0, 100.0}
	median := GetMedian(set)
	assert.Equal(t, median, 150.0)
}

//MarketMaker
//CancelExcessOrders
//TopUpBands
//TopUpBuyBands
//TopUpSellBands
//CancelAllOrders
//GetOrders
//GetBuyOrders
//GetSellOrders
//getTotalOrderAmount