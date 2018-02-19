package registry

type TokenPair struct {
	BaseToken 	string
	QuoteToken 	string
}

type ExchangeList struct {
	GATECOIN	string
	ETHFINEX	string
}

var TokenPairRegistry = map[string]TokenPair {
	"DAIUSD": TokenPair{"DAI", "USD"},
	"ETHBTC": TokenPair{"ETH", "BTC"},
	"ETHDAI": TokenPair{"ETH", "DAI"},
	"MKRBTC": TokenPair{"MKR", "BTC"},
	"MKRETH": TokenPair{"MKR", "ETH"},
}

var ExchangeTokenPairRegistry = map[string]ExchangeList {
	"DAIUSD": ExchangeList{GATECOIN: "DAIUSD"},
	"ETHBTC": ExchangeList{GATECOIN: "ETHBTC"},
	"ETHDAI": ExchangeList{GATECOIN: "ETHDAI"},
	"MKRBTC": ExchangeList{GATECOIN: "MKRBTC"},
	"MKRETH": ExchangeList{GATECOIN: "MKRETH"},
}

func LookupTokenPair(pair string) (string, string) {
	return TokenPairRegistry[pair].BaseToken, TokenPairRegistry[pair].QuoteToken
}

func LookupGatecoinTokenPair(pair string) (string) {
	return ExchangeTokenPairRegistry[pair].GATECOIN
}