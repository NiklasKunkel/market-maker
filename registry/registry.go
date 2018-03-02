package registry

import(
	"strings"
	"time"
	"github.com/niklaskunkel/market-maker/logger"
	"github.com/sirupsen/logrus"
)

//Globals
var log = logger.InitLogger()

//////////////////////////////////////////////////////
//                   Data Structures                //
//////////////////////////////////////////////////////
type TokenPair struct {
	BASETOKEN 	string
	QUOTETOKEN 	string
}

type ExchangeList struct {
	GATECOIN	ExchangeTokenInfo
	ETHFINEX	ExchangeTokenInfo
}

type ExchangeTokenInfo struct {
	TOKENPAIRNAME 	string
	PRECISION 		Precision
}

type Precision struct {
	BIDPRICEPRECISION	int
	ASKPRICEPRECISION 	int
	BIDAMOUNTPRECISION 	int 
	ASKAMOUNTPRECISION 	int
}

type ApiTimeout struct {
	PUBLICTIMEOUT			int64
	PRIVATETIMEOUT			int64
	LastPublicExecution 	int64
	LastPrivateExecution	int64
}
//////////////////////////////////////////////////////
//                   Registry Data                  //
//////////////////////////////////////////////////////
var TokenPairRegistry = map[string]TokenPair {
	"DAIUSD": TokenPair{"DAI", "USD"},
	"ETHBTC": TokenPair{"ETH", "BTC"},
	"ETHDAI": TokenPair{"ETH", "DAI"},
	"MKRBTC": TokenPair{"MKR", "BTC"},
	"MKRETH": TokenPair{"MKR", "ETH"},
}

var ExchangeTokenPairRegistry = map[string]ExchangeList {
	"DAIUSD": ExchangeList{GATECOIN: ExchangeTokenInfo{TOKENPAIRNAME: "DAIUSD", PRECISION: Precision{BIDPRICEPRECISION: 10, ASKPRICEPRECISION: 10, BIDAMOUNTPRECISION: 10, ASKAMOUNTPRECISION: 10}}},
	"ETHBTC": ExchangeList{GATECOIN: ExchangeTokenInfo{TOKENPAIRNAME: "ETHBTC", PRECISION: Precision{BIDPRICEPRECISION: 10, ASKPRICEPRECISION: 10, BIDAMOUNTPRECISION: 10, ASKAMOUNTPRECISION: 10}}},
	"ETHDAI": ExchangeList{GATECOIN: ExchangeTokenInfo{TOKENPAIRNAME: "ETHDAI", PRECISION: Precision{BIDPRICEPRECISION: 2, ASKPRICEPRECISION: 10, BIDAMOUNTPRECISION: 10, ASKAMOUNTPRECISION: 10}}},
	"MKRBTC": ExchangeList{GATECOIN: ExchangeTokenInfo{TOKENPAIRNAME: "MKRBTC", PRECISION: Precision{BIDPRICEPRECISION: 10, ASKPRICEPRECISION: 10, BIDAMOUNTPRECISION: 10, ASKAMOUNTPRECISION: 10}}},
	"MKRETH": ExchangeList{GATECOIN: ExchangeTokenInfo{TOKENPAIRNAME: "MKRETH", PRECISION: Precision{BIDPRICEPRECISION: 10, ASKPRICEPRECISION: 10, BIDAMOUNTPRECISION: 10, ASKAMOUNTPRECISION: 10}}},
}

var ExchangeApiTimeoutRegistry = map[string]*ApiTimeout {
	"GATECOIN": &ApiTimeout{PUBLICTIMEOUT: 2000, PRIVATETIMEOUT: 2000, LastPublicExecution: 0, LastPrivateExecution: 0},
	"ETHFINEX": &ApiTimeout{PUBLICTIMEOUT: 1000, PRIVATETIMEOUT: 1000, LastPublicExecution: 0, LastPrivateExecution: 0},
}

//////////////////////////////////////////////////////
//                   Getter Functions               //
//////////////////////////////////////////////////////
func LookupTokenPair(pair string) (string, string) {
	return TokenPairRegistry[pair].BASETOKEN, TokenPairRegistry[pair].QUOTETOKEN
}

func LookupGatecoinTokenPairName(pair string) (string) {
	return ExchangeTokenPairRegistry[pair].GATECOIN.TOKENPAIRNAME
}

func LookupGatecoinTokenPairPrecision(pair string) (Precision) {
	return ExchangeTokenPairRegistry[pair].GATECOIN.PRECISION
}

func MakeTimestamp() (int64) {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetExchangeApiPublicTimeout(exchange string) (int64) {
	if reg, ok := ExchangeApiTimeoutRegistry[strings.ToUpper(exchange)]; ok {
		timeToWait := reg.PUBLICTIMEOUT + reg.LastPublicExecution - MakeTimestamp()
		if (timeToWait <= 0) {
			return 0
		} else {
			return timeToWait
		}
	}
	log.WithFields(logrus.Fields{"function": "GetExchangeApiPublicTimeout", "exchange": exchange}).Error("Could not find exchange in ApiTimeoutRegistry")
	return 0
}

func GetExchangeApiPrivateTimeout(exchange string) (int64) {
	if reg, ok := ExchangeApiTimeoutRegistry[strings.ToUpper(exchange)]; ok {
    	timeToWait := reg.PRIVATETIMEOUT + reg.LastPrivateExecution - MakeTimestamp()
		if (timeToWait <= 0) {
			return 0
		} else {
			return timeToWait
		}
	}
	log.WithFields(logrus.Fields{"function": "GetExchangeApiPrivateTimeout", "exchange": exchange}).Error("Could not find exchange in ApiTimeoutRegistry")
	return 0
}

func SetExchangeApiPublicTimeout(exchange string) {
	exchange = strings.ToUpper(exchange)
	if _, ok := ExchangeApiTimeoutRegistry[exchange]; ok {
		ExchangeApiTimeoutRegistry[exchange].LastPrivateExecution = MakeTimestamp()
		return
	}
	log.WithFields(logrus.Fields{"function": "SetExchangeApiPublicTimeout", "exchange": exchange}).Error("Could not find exchange in ApiTimeoutRegistry")
	return
}

func SetExchangeApiPrivateTimeout(exchange string) {
	exchange = strings.ToUpper(exchange)
	if _, ok := ExchangeApiTimeoutRegistry[exchange]; ok {
		ExchangeApiTimeoutRegistry[exchange].LastPublicExecution = MakeTimestamp()
		return
	}
	log.WithFields(logrus.Fields{"function": "SetExchangeApiPublicTimeout", "exchange": exchange}).Error("Could not find exchange in ApiTimeoutRegistry")
	return
}