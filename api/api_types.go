package api

type ResponseStatus struct {
	ErrorCode 	string 			`json:"errorCode"`
	Message 	string 			`json:"message"`
}

type ErrorResponse struct {
	Status 	ResponseStatus 		`json:"responseStatus"`
}

type TickersResponse struct {
	Tickers 	[]Ticker		`json:"tickers"`
	Status 		ResponseStatus 	`json:"responseStatus"`
}

type Ticker struct {
	Pair 	string 		`json:"currencyPair"`
	Open 	float64 	`json:"open,omitmepty"`
	Last 	float64 	`json:"last,omitmepty"`
	LastQ 	float64 	`json:"lastQ,omitmepty"`
	High 	float64 	`json:"high,omitmepty"`
	Low 	float64 	`json:"low,omitmepty"`
	Volume 	float64 	`json:"volume,omitmepty"`
	Volumn 	float64 	`json:"volumn,omitmepty"`
	Bid 	float64 	`json:"bid,omitmepty"`
	BidQ 	float64 	`json:"bidQ,omitmepty"`
	Ask 	float64 	`json:"ask,omitmepty"`
	AskQ 	float64 	`json:"askQ,omitmepty"`
	Vwap 	float64 	`json:"vwap,omitmepty"`
	Time 	string 		`json:"createDateTime"`
}

type MarketDepthResponse struct {
	Asks 	[]Offer		`json:"asks"`
	Bids 	[]Offer 	`json:"bids"`
	Status 	ResponseStatus 	`json:"responseStatus"`
}

type Offer 	struct {
	Price 	float64  	`json:"price"`
	Volume 	float64 	`json:"volume"`
}

type TransactionsResponse struct {
	Transactions []Transaction 	`json:"transactions"`
	Status 	ResponseStatus 		`json:"responseStatus"`
}

type Transaction struct {
	Id 			int64 		`json:"transactionId"`
	Time 		string		`json:"transactionTime"`
	Price 		float64 	`json:"price"`
	Quantity 	float64 	`json:"quantity"`
	Pair 		string 		`json:"currencyPair"`
	Way 		string 		`json:"way"`
	AskId 		string 		`json:"askOrderId"`
	BidId 		string 		`json:"bidOrderId"`
}

type BalancesResponse struct {
	Balances 		[]Balance 		`json:"balances"`
	Status 			ResponseStatus 	`json:"responseStatus"`
}

type BalanceResponse struct {
	Balance 	Balance 		`json:"balance"`
	Status 		ResponseStatus 	`json:"responseStatus"`
}

type Balance struct {
	Currency 			string		`json:"currency"`
	Balance 			float64		`json:"balance"`
	AvailableBalance 	float64		`json:"availableBalance"`
	PendingIncoming 	float64		`json:"pendingIncoming"`
	PendingOutgoing 	float64		`json:"pendingOutgoing"`
	OpenOrder 			float64		`json:"openOrder"`
	IsDigital 			bool		`json:"isDigital"`
}

type CreateOrderResponse struct {
	OrderId 	string			`json:"clOrderId"`
	Status 		ResponseStatus 	`json:"responseStatus"`
}

type NewOrder struct {
	Pair 	string	`json:"Code"`
	Way 	string	`json:"Way"`
	Amount 	string	`json:"Amount"`
	Price 	string	`json:"Price"`
}

type GetOrdersResponse struct {
	Orders 	[]Order 		`json:"orders"`
	Status 	ResponseStatus 	`json:"responseStatus"`
}

type GetOrderResponse struct {
	Order 	Order			`json:"order"`
	Status 	ResponseStatus 	`json:"responseStatus"`
}

type Order struct {
	Code 			string 	`json:"code"`
	OrderId 		string 	`json:"clOrderId"`
	Side 			int64 	`json:"side"`
	Price 			float64 `json:"price"`
	InitQuantity 	float64 `json:"initialQuantity"`
	RemQuantity 	float64 `json:"remainingQuantity"`
	Status 			int64 	`json:"status"`
	StatusDesc 		string 	`json:"statusDesc"`
	TxSeqNo 		int64 	`json:"tranSeqNo"`
	Type 			int64 	`json:"type"`
	Date 			string 	`json:"date"`
}

type KillOrderResponse struct {
	Status 	ResponseStatus 	`json:"responseStatus"`
}

type WithdrawResponse struct {
	Status 	ResponseStatus 	`json:"responseStatus"`
}