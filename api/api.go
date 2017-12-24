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
	//"strconv"
	"strings"
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
	"Balance/Balances",				//GET
	"Trade/Orders",					//GET/POST/DELETE
	"ElectronicWallet/Withdrawals",	//POST
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
func (gatecoin *GatecoinClient) queryPublic(values url.Values, typ interface{}) (interface{}, error) {
	//check if valid command
	if !IsStringInSlice(values.Get("command"), publicMethods) {
		return nil, fmt.Errorf("[Gatecoin] (queryPublic) The method %s is not in the supported Public Methods list.", values.Get("command"))
	}
	//format request URL w/ path and URL parameters
	reqURL, _ := url.Parse(APIHostUrl)
	reqURL.Path = "/Public"
	reqURL.RawQuery = values.Encode()
	values = nil

	//set type of request
	requestType := "GET"
	resp, err := gatecoin.doRequest(reqURL, requestType, values, nil, typ)
	return resp, err
}

/*
func (gatecoin *GatecoinClient) queryPrivate(values url.Values, responseType interface{}) (interface{}, error) {
	cmd := values.Get("command)")
	//check if valid command
	if !IsStringInSlice(cmd, privateMethods) {
		return nil, fmt.Errorf("[Gatecoin] (queryPrivate) The method %s is not in the supposed Private Methods list.", cmd)
	}

	//Set url for request
	reqURL, _ := url.Parse(APIHostUrl)	//sets scheme to https - host to gatecoin.com
	reqURL.Path = "/" + cmd

	//UNFINISHED
	return 
}
*/

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
		url.Values{ "command": []string{"LiveTickers"}},
		&TickersResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*TickersResponse), nil
}

func (gatecoin *GatecoinClient) GetMarketDepth(pair string) (*MarketDepthResponse, error) {
	//Make request
	resp, err := gatecoin.queryPublic(
		url.Values { "command": []string{"MarketDepth"}},
		&MarketDepthResponse{})
	if err != nil {
		return nil, err
	}
	return resp.(*MarketDepthResponse), nil
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