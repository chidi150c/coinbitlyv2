package binanceapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"coinbitly.com/helper"
)

// TickerData represents the ticker data structure
type TickerData struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// 24hrTickerData represents the 24-hour ticker data structure
type Ticker24hrData struct {
	Symbol      string `json:"symbol"`
	LastPrice   string `json:"lastPrice"`
	PriceChange string `json:"priceChange"`
	Volume      string `json:"volume"`
}

// Candlestick represents a single candlestick data
type Candlestick struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// OrderBookData represents the order book data structure
type OrderBookData struct {
	Bids [][2]string `json:"bids"` // [price, quantity]
	Asks [][2]string `json:"asks"` // [price, quantity]
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
