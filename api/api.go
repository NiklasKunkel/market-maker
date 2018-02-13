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
	"github.com/niklaskunkel/market-maker/logger"
	"github.com/sirupsen/logrus"
)

//Constants
const (
	APIHostUrl = "https://api.gatecoin.com"
	APIUserAgent = "MakerDAO Market-Maker"
)

//Globals
var log = logger.InitLogger()

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

//Type Structs
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
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "queryPublic", "Command": cmd}).Error("Command is not in supported Public Commands list")
		return nil, fmt.Errorf("Unsupported Public Method")
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
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "queryPrivate", "Command": cmd}).Error("Command is not in supported Private Commands list")
		return nil, fmt.Errorf("Unsupported Private Method.", cmd)
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
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "doRequest", "requestType": requestType, "requestURL": reqURL.String(), "data": data, "error": err.Error()}).Error("Failed to create new request")
		return nil, err
	}

	//Add headers to request
	req.Header.Set("User-Agent", APIUserAgent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	//Log copy of request to debug
	log.WithFields(logrus.Fields{"client": "Gatecoin", "request": req}).Debug("Request being sent")

	//Execute request
	resp, err := gatecoin.client.Do(req)
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "doRequest", "requestType": requestType, "requestURL": reqURL.String(), "data": data, "request": req, "error": err.Error()}).Error("Failed to execute request")
		return nil, err
	}
	defer resp.Body.Close() 

	//Read response
	body, err := ioutil.ReadAll(resp.Body)

	//Log copy of response to debug
	log.WithFields(logrus.Fields{"client": "Gatecoin", "response": body}).Debug("Response received")

	//Check if parsing response failed
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "doRequest", "body": resp.Body, "request": req, "error": err.Error()}).Error("Failed to parse response for query")
		return nil, err
	}

	//Convert JSON to ErrorResponse struct to check if API returned error
	apiError := ErrorResponse{}
	err = json.Unmarshal(body, &apiError)
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "doRequest", "body": body, "request": req, "error": err.Error()}).Error("Failed to convert JSON response into error struct")
		return nil, err
	}
	if apiError.Status.Message != "OK" {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "doRequest", "errorCode": apiError.Status.ErrorCode}).Error(apiError.Status.Message)
		return nil, fmt.Errorf(apiError.Status.Message)
	}

	//Convert JSON to Response struct
	err = json.Unmarshal(body, &responseType)
	if err != nil {
		log.WithFields(logrus.Fields{"client": "Gatecoin", "function": "doRequestuest", "body": body, "request": req, "error": err.Error()}).Error("Failed to convert JSON response into struct")
		return nil, err
		return nil, fmt.Errorf("[Gatecoin] (doRequest) Failed to convert JSON response into struct for query to %s! (%s)", reqURL, err.Error())
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

//TODO - add TransactionID as query parameter ?TransactionId=BK11538053033
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

func (gatecoin *GatecoinClient) CreateOrder(pair string, way string, amount string, price string) (*CreateOrderResponse, error) {
	//compose order obj
	order := NewOrder{pair, way, amount, price}
	//convert to json string
	orderJson, err := json.Marshal(order)
	fmt.Printf("\nOrder JSON string = %s\n", orderJson)
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

func (gatecoin *GatecoinClient) GetOrders() (*GetOrderResponse, error) {
	resp, err := gatecoin.queryPrivate(
		"GET",
		[]string{"Trade/Orders"},
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