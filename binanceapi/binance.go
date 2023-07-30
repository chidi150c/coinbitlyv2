package binanceapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	baseURL    = "https://api.binance.com"
	apiVersion = "api/v3"
)

var (
	apiKey    = "1CHA2mXswJdHjfossO43t4WRa82HPzFaeZOt2entAgajkAYIUaf55f7CepLt58YK" // Replace with your Binance API key
	secretKey = "0DiulGqjlOuQlQHcVQYLYjKfpkq6Qs2rNxNVHHTtF1s2uy7n5clugQRxTXltjqFj" // Replace with your Binance API secret key
)

// rateLimitedTransport is a custom transport that handles API rate limiting
type rateLimitedTransport struct {
	base http.RoundTripper
}

func (t rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Implement rate limiting logic here (e.g., using a rate limiter)
	// For simplicity, we'll use a simple time.Sleep to limit the rate
	time.Sleep(100 * time.Millisecond)
	return t.base.RoundTrip(req)
}

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

// FetchHistoricalCandlesticks fetches historical candlestick data for the given symbol and time interval
func fetchHistoricalCandlesticks(symbol, interval string, startTime, endTime time.Time) ([]Candlestick, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: rateLimitedTransport{base: http.DefaultTransport},
	}

	// Convert time to milliseconds
	startTimeUnix := startTime.Unix() * 1000
	endTimeUnix := endTime.Unix() * 1000

	// Construct the API URL
	url := fmt.Sprintf("https://api.binance.com/api/v1/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d", symbol, interval, startTimeUnix, endTimeUnix)

	// Send the HTTP GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the API request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
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
			Open:      ParseStringToFloat(open),
			High:      ParseStringToFloat(high),
			Low:       ParseStringToFloat(low),
			Close:     ParseStringToFloat(close),
			Volume:    ParseStringToFloat(volume),
		}

		result = append(result, candlestick)
	}

	return result, nil
}

// ParseStringToFloat parses a string to a float64 value
func ParseStringToFloat(str string) float64 {
	val, _ := strconv.ParseFloat(str, 64)
	return val
}






// OrderBookData represents the order book data structure
type OrderBookData struct {
	Bids [][2]string `json:"bids"` // [price, quantity]
	Asks [][2]string `json:"asks"` // [price, quantity]
}

// fetchTickerData fetches real-time market data for the given symbol
func fetchTickerData(symbol string) (*TickerData, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: rateLimitedTransport{base: http.DefaultTransport},
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
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
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
func fetch24hrTickerData(symbol string) (*Ticker24hrData, error) {// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: rateLimitedTransport{base: http.DefaultTransport},
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
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
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
func fetchOrderBook(symbol string, limit int) (*OrderBookData, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: rateLimitedTransport{base: http.DefaultTransport},
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
