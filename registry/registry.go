package registry

type TokenPair struct {
	BaseToken string
	QuoteToken string
}

var TokenPairRegistry = map[string]TokenPair {
	"DAIUSD": TokenPair{"DAI", "USD"},
	"ETHBTC": TokenPair{"ETH", "BTC"},
	"ETHDAI": TokenPair{"ETH", "DAI"},
	"MKRBTC": TokenPair{"MKR", "BTC"},
	"MKRETH": TokenPair{"MKR", "ETH"},
}

func LookupTokenPair(pair string) (string, string) {
	return TokenPairRegistry[pair].BaseToken, TokenPairRegistry[pair].QuoteToken
}

