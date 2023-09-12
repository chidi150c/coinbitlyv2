package binanceapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"coinbitly.com/helper"
)

// Candlestick represents a single candlestick data
type Candlestick struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}


// GetQuoteAndBaseBalances retrieves the quote and base balances for a given trading pair.
func getQuoteAndBaseBalances(symbol, baseURL, apiVersion, apiKey, secretKey string) (float64, float64, error) {
    // Fetch exchange information to determine quote and base assets
    exchangeInfo, err := fetchExchangeInfo(symbol, baseURL, apiVersion, apiKey)
    if err != nil {
        return 0, 0, err
    }

    // Find the trading pair in the exchange information
    var quoteAsset, baseAsset string
    for _, symbolInfo := range exchangeInfo["symbols"].([]interface{}) {
        symbolMap := symbolInfo.(map[string]interface{})
        if symbolMap["symbol"] == symbol {
            quoteAsset = symbolMap["quoteAsset"].(string)
            baseAsset = symbolMap["baseAsset"].(string)
            break
        }
    }

    if quoteAsset == "" || baseAsset == "" {
        return 0, 0, errors.New("Symbol not found in exchange information")
    }

    // Fetch account information to get the balances
    account, err := getAccountInfo(apiKey, baseURL, apiVersion, secretKey)
    if err != nil {
        return 0, 0, err
    }

    // Find the balances for quote and base assets
    var quoteBalance, baseBalance float64
    for _, balance := range account.Balances {
        if balance.Asset == quoteAsset {
            quoteBalance = helper.ParseStringToFloat(balance.Free)
        } else if balance.Asset == baseAsset {
            baseBalance = helper.ParseStringToFloat(balance.Free)
        }
    }

    return quoteBalance, baseBalance, nil
}

// FetchHistoricalCandlesticks fetches historical candlestick data for the given symbol and time interval
func fetchHistoricalCandlesticks(symbol, baseURL, apiVersion, apiKey, interval string, startTime, endTime int64) ([]Candlestick, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}

	// Convert time to milliseconds
	startTimeUnix := startTime * 1000
	endTimeUnix := endTime * 1000

	// Construct the API URL
	url := fmt.Sprintf("%s/%s/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d", baseURL, apiVersion, symbol, interval, startTimeUnix, endTimeUnix)
	// Send the HTTP GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the API request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request client.Get(%s) failed with status: %s", url, resp.Status)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response into a slice of candlesticks
	var candlesticks [][]interface{}
	err = json.Unmarshal(body, &candlesticks)
	if err != nil {
		return nil, err
	}

	// Convert the raw candlestick data to Candlestick structs
	var result []Candlestick
	for _, cs := range candlesticks {
		timestamp, _ := cs[0].(float64)
		open, _ := cs[1].(string)
		high, _ := cs[2].(string)
		low, _ := cs[3].(string)
		close, _ := cs[4].(string)
		volume, _ := cs[5].(string)

		candlestick := Candlestick{
			Timestamp: int64(timestamp),
			Open:      helper.ParseStringToFloat(open),
			High:      helper.ParseStringToFloat(high),
			Low:       helper.ParseStringToFloat(low),
			Close:     helper.ParseStringToFloat(close),
			Volume:    helper.ParseStringToFloat(volume),
		}

		result = append(result, candlestick)
	}
	return result, nil
}

// TickerData represents the ticker data structure
type TickerData struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}
// fetchTickerData fetches real-time of a given symbol
func fetchTickerData(symbol, baseURL, apiVersion, apiKey string) (*TickerData, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	url := fmt.Sprintf("%s/%s/ticker/price?symbol=%s", baseURL, apiVersion, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add API Key to the request header (if required)
	req.Header.Add("X-MBX-APIKEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request client.Get(%s) failed with status: %s", url, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ticker TickerData
	err = json.Unmarshal(body, &ticker)
	if err != nil {
		return nil, err
	}

	return &ticker, nil
}


// 24hrTickerData represents the 24-hour ticker data structure
type Ticker24hrData struct {
	Symbol      string `json:"symbol"`
	LastPrice   string `json:"lastPrice"`
	PriceChange string `json:"priceChange"`
	Volume      string `json:"volume"`
}
// fetch24hrTickerData fetches 24-hour price change statistics for the given symbol
func fetch24hrTickerData(symbol, baseURL, apiVersion, apiKey string) (*Ticker24hrData, error) {// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	url := fmt.Sprintf("%s/%s/ticker/24hr?symbol=%s", baseURL, apiVersion, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add API Key to the request header (if required)
	req.Header.Add("X-MBX-APIKEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request client.Get(%s) failed with status: %s", url, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ticker Ticker24hrData
	err = json.Unmarshal(body, &ticker)
	if err != nil {
		return nil, err
	}

	return &ticker, nil
}

// OrderBookData represents the order book data structure
type OrderBookData struct {
	Bids [][2]string `json:"bids"` // [price, quantity]
	Asks [][2]string `json:"asks"` // [price, quantity]
}
// fetchOrderBook fetches order book depth for the given symbol
func fetchOrderBook(symbol, baseURL, apiVersion, apiKey string, limit int) (*OrderBookData, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	url := fmt.Sprintf("%s/%s/depth?symbol=%s&limit=%d", baseURL, apiVersion, symbol, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add API Key to the request header (if required)
	req.Header.Add("X-MBX-APIKEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var orderBook OrderBookData
	err = json.Unmarshal(body, &orderBook)
	if err != nil {
		return nil, err
	}

	return &orderBook, nil
}

type ExchangeInfo struct {
	Symbols []struct {
		Symbol     string `json:"symbol"`
		Status     string `json:"status"`
		BaseAsset  string `json:"baseAsset"`
		BaseAssetPrecision int `json:"baseAssetPrecision"`
		QuoteAsset string `json:"quoteAsset"`
		QuotePrecision int `json:"quotePrecision"`
		Filters []struct {
			FilterType string `json:"filterType"`
			MinQty string `json:"minQty"`
		} `json:"filters"`
	} `json:"symbols"`
}
// fetchExchangeInfo fetches exchange information including minimum order quantity
func fetchExchangeInfo(symbol, baseURL, apiVersion, apiKey string) (map[string]interface{}, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	url := fmt.Sprintf("%s/%s/exchangeInfo", baseURL, apiVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add API Key to the request header (if required)
	req.Header.Add("X-MBX-APIKEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response
	var exchangeInfo map[string]interface{}
	err = json.Unmarshal(body, &exchangeInfo)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	return exchangeInfo, nil

}

type Fill struct {
    Price            string `json:"price"`
    Qty              string `json:"qty"`
    Commission       string `json:"commission"`
    CommissionAsset  string `json:"commissionAsset"`
    TradeID          int    `json:"tradeId"`
}

type Response struct {
    Symbol                   string  `json:"symbol"`
    OrderID                  int     `json:"orderId"`
    OrderListID              int     `json:"orderListId"`
    ClientOrderID            string  `json:"clientOrderId"`
    TransactTime             int64   `json:"transactTime"`
    Price                    string  `json:"price"`
    OrigQty                  string  `json:"origQty"`
    ExecutedQty              string  `json:"executedQty"`
    CumulativeQuoteQty       string  `json:"cummulativeQuoteQty"` // Note: "cummulative" is a typo, it should be "cumulative"
    Status                   string  `json:"status"`
    TimeInForce              string  `json:"timeInForce"`
    Type                     string  `json:"type"`
    Side                     string  `json:"side"`
    WorkingTime              int64   `json:"workingTime"`
    Fills                    []Fill  `json:"fills"`
    SelfTradePreventionMode  string  `json:"selfTradePreventionMode"`
}


// placeOrder places an order
func placeOrder(symbol, side, orderType, timeInForce, price, quantity,  baseURL, apiVersion, apiKey, secretKey string) (Response, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	url := fmt.Sprintf("%s/%s/order", baseURL, apiVersion)
	params := map[string]string{
		"symbol":      symbol,
		"side":        side,
		"type":        orderType,
		"timeInForce": timeInForce,
		"price":       price,
		"quantity":    quantity,
	}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return Response{}, err
	}	
	query := req.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()
	sign(req, apiKey, secretKey) // Sign the request with API key and secret
	// logRequestDetails(req)
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		logResponseDetails(resp)
		return Response{}, err
	}
	return response, nil
}

// sign signs the HTTP request with HMAC-SHA256 authentication
func sign(req *http.Request, apiKey, secretKey string) {
	query := req.URL.Query()

	// Add a timestamp parameter
	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	query.Add("timestamp", timestamp)

	// Sort and encode the query parameters
	params := query.Encode()

	// Generate the HMAC-SHA256 signature
	signature := generateSignature(params, secretKey)

	// Set the signature parameter
	query.Set("signature", signature)

	// Update the query string with the signature
	req.URL.RawQuery = query.Encode()

	// Add the API key and signature headers
	req.Header.Add("X-MBX-APIKEY", apiKey)
	req.Header.Add("X-MBX-SIGNATURE", signature)
}


type AccountInfo struct {
    Balances []Balance `json:"balances"`
}

type Balance struct {
    Asset  string `json:"asset"`
    Free   string `json:"free"`
    Locked string `json:"locked"`
}

// generateSignature generates the HMAC-SHA256 signature
func generateSignature(data, secret string) string {
	hmacSha256 := hmac.New(sha256.New, []byte(secret))
	hmacSha256.Write([]byte(data))
	return hex.EncodeToString(hmacSha256.Sum(nil))
}
func getAccountInfo(apiKey, baseURL, apiVersion, secretKey string) (AccountInfo, error) {
	// Create an HTTP client
	client := &http.Client{}

	// Prepare the request to fetch account information
	url := fmt.Sprintf("%s/%s/account", baseURL, apiVersion)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// logRequestDetails(req)
		return AccountInfo{}, err
	}


	sign(req, apiKey, secretKey) // Sign the request with API key and secret

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return AccountInfo{}, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// logResponseDetails(resp)
		return AccountInfo{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}
	// Parse the response JSON
	var accountInfo AccountInfo
	err = json.NewDecoder(resp.Body).Decode(&accountInfo)
	if err != nil {
		return AccountInfo{}, err
	}

	return accountInfo, nil
}

// Function to log request details
func logRequestDetails(req *http.Request) {
    fmt.Println("Request URL:", req.URL.String())
    fmt.Println("Request Method:", req.Method)
    fmt.Println("Request Headers:")
    for key, values := range req.Header {
        fmt.Printf("%s: %s\n", key, values)
    }
    fmt.Println("Request Body:")
    if req.Body != nil {
        buf := new(bytes.Buffer)
        buf.ReadFrom(req.Body)
        fmt.Println(buf.String())
    }
}
func logResponseDetails(resp *http.Response) {
    fmt.Println("Response Status:", resp.Status)
    fmt.Println("Response Headers:")
    for key, values := range resp.Header {
        fmt.Printf("%s: %s\n", key, values)
    }
    fmt.Println("Response Body:")
    buf := new(bytes.Buffer)
    buf.ReadFrom(resp.Body)
    fmt.Println(buf.String())
}