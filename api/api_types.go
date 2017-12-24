package api

type ResponseStatus struct {
	ErrorCode 	string 			`json:"errorCode,omitempty"`
	Message 	string 			`json:"message,omitempty"`
}

type TickersResponse struct {
	Tickers 	[]Ticker		`json:"tickers,omitempty"`
	Status 		ResponseStatus	`json:"responseStatus,omitempty"`
}

type Ticker struct {
	Pair 	string 		`json:"currencyPair,omitempty"`
	Open 	float64 	`json:"open,omitmepty"`
	Last 	float64 	`json:"last,omitmepty"`
	LastQ 	float64 	`json:"lastQ,omitmepty"`
	High 	float64 	`json:"high,omitmepty"`
	Low 	float64 	`json:"low,omitmepty"`
	Volume 	float64 	`json:"volume,omitmepty"`
	Volumn 	float64 	`json:"volumn,omitmepty"`
	Bid 	float64 	`json:"bid,omitmepty"`
	BidQ 	float64 	`json:"bidQ,omitmepty"`
	ask 	float64 	`json:"ask,omitmepty"`
	askQ 	float64 	`json:"askQ,omitmepty"`
	Vwap 	float64 	`json:"vwap,omitmepty"`
	Time 	int64 		`json:"createDateTime,omitempty"`
}

type MarketDepthResponse struct {
	Asks 	[]Offer		`json:"asks,omitempty"`
	Bids 	[]Offer 	`json:"bids,omitempty"`
	Status 	ResponseStatus 	`json:"responseStatus,omitempty"`
}

type Offer 	struct {
	Price 	float64  	`json:"price,omitempty"`
	Volume 	float64 	`json:"volume,omitempty"`
}

type TransactionResponse struct {
	Transactions []Transaction 	`json:"transactions,omitempty"`
	Status 	ResponseStatus 		`json:"responseStatus,omitempty"`
}

type Transaction struct {
	Id 			int64 		`json:"transactionId,omitempty"`
	Time 		int64		`json:"transactionTime,omitempty"`
	Price 		float64 	`json:"price,omitempty"`
	Quantity 	float64 	`json:"quantity,omitempty"`
	Pair 		string 		`json:"currencyPair,omitempty"`
	Way 		string 		`json:"way,omitempty"`
	AskId 		string 		`json:"askOrderId,omitempty"`
	BidId 		string 		`json:"bidOrderId,omitempty"`
}

type BalancesResponse struct {
	Balances 	[]Balance 		`json:"balances,omitempty"`
	Status 		ResponseStatus 	`json:"responseStatus,omitempty"`
}

type Balance struct {
	Currency 			string 		`json:"currency,omitempty"`
	Balance 			float64 	`json:"balance,omitempty"`
	AvailableBalance 	float64 	`json:"availableBalance,omitempty"`
	PendingIncoming 	float64 	`json:"pendingIncoming,omitempty"`
	PendingOutgoing 	float64 	`json:"pendingOutgoing,omitempty"`
	OpenOrder 			int64 		`json:"openOrder,omitempty"`
	IsDigital 			bool 		`json:"isDigital,omitempty"`
}

type CreateOrderResponse struct {
	OrderId 	string 			`json:"clOrderId,omitempty"`
	Status 		ResponseStatus 	`json:"responseStatus,omitempty"`
}

type GetOrderResponse struct {
	Orders 	[]Order 		`json:"orders,omitempty"`
	Status 	ResponseStatus 	`json:"responseStatus,omitempty"`
}

type Order struct {
	Code 			string 	`json:"code,omitempty"`
	OrderId 		string 	`json:"clOrderId,omitemptu"`
	Side 			int64 	`json:"side,omitempty"`
	Price 			float64 `json:"price,omitempty"`
	InitQuantity 	float64 `json:"initialQuantity,omitempty"`
	RemQuantity 	float64 `json:"remainingQuantity,omitempty"`
	Status 			int64 	`json:"status,omitempty"`
	StatusDesc 		string 	`json:"statusDesc,omitempty"`
	TxSeqNo 		int64 	`json:"tranSeqNo,omitempty"`
	Type 			int64 	`json:"type,omitempty"`
	Date 			string 	`json:"date,omitempty"`
}

type DeleteOrderResponse struct {
	Status 	ResponseStatus 	`json:"responseStatus,omitempty"`
}

type WithdrawResponse struct {
	Status 	ResponseStatus 	`json:"responseStatus,omitempty"`
}