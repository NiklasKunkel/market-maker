package api

import(
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	APIHostUrl = "https://api.gatecoin.com"
	APIUserAgent = "MakerDAO Market-Maker"
)

var publicMethods = []string {
	"LiveTickers",
	"MarketDepth",
	"Transactions",
}

var privateMethods = []string {
	"Balance/Balances",
	"Trade/Orders",
	"ElectronicWallet/Withdrawals",
}

type GatecoinClient struct {
	key 	string			//Gatecoin API Key
	secret 	string			//Gatecoin Secret Key
	client 	*http.Client 	
}

func NewGatecoinClient(key, secret string) (*GatecoinClient) {
	client := &http.Client{}
	return &GatecoinClient{key, secret, client}
}

/////////////////////////////////////////////////////////////////////////
//                          REQUEST CONSTRUCTION                       //
/////////////////////////////////////////////////////////////////////////
func (gatecoin *GatecoinClient) queryPublic(params []string, typ interface{}) (interface{}, error) {
	//check if valid command
	cmd := params[0]
	if !IsStringInSlice(cmd, publicMethods) {
		return nil, fmt.Errorf("[Gatecoin] (queryPublic) The method %s is not in the supported Public Methods list.", cmd)
	}
	//format request URL w/ path and URL parameters
	reqURL, _ := url.Parse(APIHostUrl)
	reqURL.Path = "/Public"
	for _, param := range params {
		reqURL.Path += "/" + param
	}

	//set type of request
	requestType := "GET"
	resp, err := gatecoin.doRequest(reqURL, requestType, nil, []byte{}, typ)
	return resp, err
}

func (gatecoin *GatecoinClient) queryPrivate(requestType string, params []string, data []byte, responseType interface{}) (interface{}, error) {
	cmd := params[0]
	//check if valid command
	if !IsStringInSlice(cmd, privateMethods) {
		return nil, fmt.Errorf("[Gatecoin] (queryPrivate) The method %s is not in the supposed Private Methods list.", cmd)
	}

	//Set url for request
	reqURL, _ := url.Parse(APIHostUrl)
	for _, param := range params {
		if param != "" {
			reqURL.Path += "/" + param
		}
	}

	//set content type
	var contentType string
	if requestType != "GET" {
		contentType = "application/json"
	}

	//set nonce
	nonce := strconv.FormatInt(time.Now().Unix(), 10) + ".000"

	//construct message
	msg := fmt.Sprintf("%s%s%s%s",requestType, reqURL, contentType, nonce)
	fmt.Printf("Message = %s\n", msg)

	//Create signature using secret
	signature := createSignature(msg, gatecoin.secret)

	//Add api key and encrypted signature to headers
	headers := map[string]string {
		"API_PUBLIC_KEY": gatecoin.key,
		"API_REQUEST_SIGNATURE": signature,
		"API_REQUEST_DATE": nonce,
	}

	resp, err := gatecoin.doRequest(reqURL, requestType, headers, data, responseType)
	return resp, err
}

func (gatecoin *GatecoinClient) doRequest(reqURL *url.URL, requestType string, headers map[string]string, data []byte, responseType interface{}) (interface{}, error) {
	//Create request
	req, err := http.NewRequest(requestType, reqURL.String(), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("[Gatecoin] (doRequest) Could not execute request to %s (%s)", reqURL, err.Error())
	}

	//Add headers to request
	req.Header.Set("User-Agent", APIUserAgent)
	req.Header.Set("Accept", "application/json")

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	//debug
	//Print request prior to execution
	fmt.Printf("\nREQUEST = %+v\n", req)

	//Execute request
	resp, err := gatecoin.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[Gatecoin] (doRequest) Could not execute request to %s! (%s)", reqURL, err.Error())
	}
	defer resp.Body.Close() 

	//Read response
	body, err := ioutil.ReadAll(resp.Body)

	//debug
	//Print response for request
	fmt.Printf("\nRESPONSE = %s\n",body)

	if err != nil {
		return nil, fmt.Errorf("[Gatecoin] (doRequest) Failed to parse response for query to %s! (%s)", reqURL, err.Error())
	} 

	//Convert JSON to ErrorResponse struct to check if API returned error
	apiError := ResponseStatus{}
	json.Unmarshal(body, &apiError)
	if apiError.Message != "OK" {
		return nil, fmt.Errorf("Error Code %s, %s", apiError.ErrorCode, apiError.Message)
	}

	//Convert JSON to Response struct
	err = json.Unmarshal(body, &responseType)
	if err != nil {
		return nil, fmt.Errorf("[Gatecoin] (doRequest) Failed to convert response into JSON for query to %s! (%s)", reqURL, err.Error())
	}
	return responseType, nil
}
/////////////////////////////////////////////////////////////////////////
//                          PUBLIC API METHODS                         //
/////////////////////////////////////////////////////////////////////////

func (gatecoin *GatecoinClient) GetTickers() (*TickersResponse, error) {
	resp, err := gatecoin.queryPublic(
		[]string{"LiveTickers"},
		&TickersResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*TickersResponse), nil
}

func (gatecoin *GatecoinClient) GetMarketDepth(pair string) (*MarketDepthResponse, error) {
	//Make request
	resp, err := gatecoin.queryPublic(
		[]string{"MarketDepth", pair},
		&MarketDepthResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*MarketDepthResponse), nil
}

func (gatecoin *GatecoinClient) GetTransactions(pair string) (*TransactionsResponse, error) {
	resp, err := gatecoin.queryPublic(
		[]string{"Transactions", pair},
		&TransactionsResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*TransactionsResponse), nil
}

/////////////////////////////////////////////////////////////////////////
//                          PRIVATE API METHODS                        //
/////////////////////////////////////////////////////////////////////////

func (gatecoin *GatecoinClient) GetBalances(pair string) (*BalancesResponse, error) {
	resp, err := gatecoin.queryPrivate(
		"GET",
		[]string{"Balance/Balances", pair},
		[]byte{},
		&BalancesResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*BalancesResponse), nil
}

//UNFINISHED - args and formatting
func (gatecoin *GatecoinClient) CreateOrder(pair string, way string, amount float64, price float64) (*CreateOrderResponse, error) {
	//compose order obj
	order := NewOrder{pair, way, amount, price}
	//convert to json string
	orderJson, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	resp, err := gatecoin.queryPrivate(
		"POST",
		[]string{"Trade/Orders"},
		orderJson,
		&CreateOrderResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*CreateOrderResponse), nil
}

func (gatecoin *GatecoinClient) GetOrder(id string) (*GetOrderResponse, error) {
	resp, err := gatecoin.queryPrivate(
		"GET",
		[]string{"Trade/Orders", id},
		[]byte{},
		&GetOrderResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*GetOrderResponse), nil
} 

func (gatecoin *GatecoinClient) DeleteOrder(id string) (*KillOrderResponse, error) {
	resp, err := gatecoin.queryPrivate(
		"DELETE",
		[]string{"Trade/Orders", id},
		[]byte{},
		&KillOrderResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*KillOrderResponse), nil
}

//UNFINISHED - args and formatting 
func (gatecoin *GatecoinClient) Withdraw(currency string) (*WithdrawResponse, error) {
	resp, err := gatecoin.queryPrivate(
		"POST",
		[]string{"ElectronicWallet/Withdrawals", currency},
		[]byte{},
		&WithdrawResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*WithdrawResponse), nil
}

/////////////////////////////////////////////////////////////////////////
//                              ENCRYPTION                             //
/////////////////////////////////////////////////////////////////////////

///Creates a hmac hash with sha512
func getHMacSha256(msg string, secret string) []byte {
	hmac := hmac.New(sha256.New,[]byte(secret))
	hmac.Write([]byte(strings.ToLower(msg)))
	return hmac.Sum(nil)
}

//Creates signature for our HTTP requests
func createSignature(msg string, secret string) string {
	sum := getHMacSha256(msg, secret)
	hashInBase64 := base64.StdEncoding.EncodeToString(sum)
	//fmt.Printf("encrypting message: %s\nsecret in bytes: %+v\nsecret string: %s\nsha-256 []byte: %+v\nsha-256 string: %s\nbase-64 string: %s\n", strings.ToLower(msg), []byte(secret), secret, sum, string(sum[:]), hashInBase64)
	return hashInBase64
}

/////////////////////////////////////////////////////////////////////////
//                          UTILITY METHODS                            //
/////////////////////////////////////////////////////////////////////////

//Verifies if given term is in a list of strings
func IsStringInSlice(term string, list []string) bool {
	for _, found := range list {
		if term == found {
			return true
		}
	}
	return false
}