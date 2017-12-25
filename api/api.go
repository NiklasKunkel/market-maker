package api

import(
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
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
	resp, err := gatecoin.doRequest(reqURL, requestType, nil, nil, typ)
	return resp, err
}

func (gatecoin *GatecoinClient) queryPrivate(requestType string, params []string, responseType interface{}) (interface{}, error) {
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

	//Add timestamp
	reqURL.Path += strconv.FormatInt(time.Now().UnixNano(), 10)

	//Create signature using secret
	signature := createSignature(reqURL.Path, gatecoin.secret)

	//Add api key and encrypted signature to headers
	headers := map[string]string {
		"Key": gatecoin.key,
		"Sign": signature,
	}

	resp, err := gatecoin.doRequest(reqURL, requestType, nil, headers, responseType)
	return resp, err
}

func (gatecoin *GatecoinClient) doRequest(reqURL *url.URL, requestType string, values url.Values, headers map[string]string, responseType interface{}) (interface{}, error) {
	//Create request
	req, err := http.NewRequest(requestType, reqURL.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("[Gatecoin] (doRequest) Could not execute request to %s (%s)", reqURL, err.Error())
	}

	//Add headers to request
	req.Header.Set("User-Agent", APIUserAgent)
	req.Header.Set("Accept", "application/json")
	if req.Header.Set("Content-Type", ""); requestType != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}
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
		&BalancesResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*BalancesResponse), nil
}

//UNFINISHED - args and formatting
func (gatecoin *GatecoinClient) CreateOrder() (*CreateOrderResponse, error) {
	resp, err := gatecoin.queryPrivate(
		"POST",
		[]string{"Trade/Orders"},
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
func getHMacSha512(msg string, secret string) []byte {
	hmac := hmac.New(sha512.New, []byte(secret))
	hmac.Write([]byte(strings.ToLower(msg)))
	return hmac.Sum(nil)
}

//Creates signature for our HTTP requests
func createSignature(msg string, secret string) string {
	sum := getHMacSha512(msg, secret)
	return hex.EncodeToString(sum)
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