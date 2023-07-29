package binanceapi

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"time"
	"encoding/json"
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
