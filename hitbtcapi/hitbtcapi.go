package hitbtcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"coinbitly.com/helper"
)

type Candlestick struct {
	Timestamp   time.Time `json:"timestamp"`
	Open        string    `json:"open"`
	Close       string    `json:"close"`
	High        string    `json:"high"`
	Low         string    `json:"low"`
	Volume      string    `json:"volume"`
	CandleCount int64     `json:"count"`
}

type Balance struct {
	Currency           string  `json:"currency"`
	Available          string `json:"available"`
	Reserved           string `json:"reserved"`
	ReservedMargin     string `json:"reserved_margin"`
	CrossMarginReserved string `json:"cross_margin_reserved"`
}

type OrderResponse struct {
	ID               string `json:"id"`
	ClientOrderID    string `json:"clientOrderId"`
	Symbol           string `json:"symbol"`
	Side             string `json:"side"`
	Status           string `json:"status"`
	Type             string `json:"type"`
	TimeInForce      string `json:"timeInForce"`
	Quantity         string `json:"quantity"`
	Price            string `json:"price"`
	CumQuantity      string `json:"cumQuantity"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
	StopPrice        string `json:"stopPrice"`
	ExpireTime       string `json:"expireTime"`
	OriginalRequest  string `json:"originalRequest"`
	LastUpdateTime   string `json:"lastUpdateTime"`
	ReportType       string `json:"reportType"`
	TradeID          string `json:"tradeId"`
	TradeQuantity    string `json:"tradeQuantity"`
	TradePrice       string `json:"tradePrice"`
	TradeFee         string `json:"tradeFee"`
	TradeTimestamp   string `json:"tradeTimestamp"`
	TradeIsBuyer     bool   `json:"tradeIsBuyer"`
	TradeIsMaker     bool   `json:"tradeIsMaker"`
	TradeFeeCurrency string `json:"tradeFeeCurrency"`
}

type Order struct {
	ID               int64   `json:"id"`
	ClientOrderID    string  `json:"client_order_id"`
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	Status           string  `json:"status"`
	Type             string  `json:"type"`
	TimeInForce      string  `json:"time_in_force"`
	Quantity         float64 `json:"quantity"`
	Price            float64 `json:"price"`
	QuantityCumulative float64 `json:"quantity_cumulative"`
}

func getOrderByID(orderID, baseURL, apiVersion, apiKey, apiSecret string) (OrderResponse, error) {
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	var orderResp OrderResponse
	// Initialize HTTP client with a custom transport to handle rate limiting
	url := fmt.Sprintf("%s/%s/order/%s", baseURL, apiVersion, orderID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return orderResp, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(apiKey, apiSecret)

	resp, err := client.Do(req)
	if err != nil {
		return orderResp, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return orderResp, err
	}

	if resp.StatusCode != http.StatusOK {
		return orderResp, fmt.Errorf("API returned non-OK status: %s\nResponse: %s", resp.Status, string(body))
	}

	err = json.Unmarshal(body, &orderResp)
	if err != nil {
		return orderResp, err
	}

	return orderResp, nil
}

func cancelOrderByID(orderID, baseURL, apiVersion, apiKey, apiSecret string) error {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	url := fmt.Sprintf("%s/%s/order/%s", baseURL, apiVersion, orderID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(apiKey, apiSecret)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API returned non-OK status: %s\nResponse: %s", resp.Status, string(body))
	}

	return nil
}

func checkBalance(currency string, amount float64, baseURL string, apiVersion string, apiKey string, apiSecret string) bool {
	balance, err := getBalance(currency, baseURL, apiVersion, apiKey, apiSecret)
	if err != nil {
		fmt.Println("Error checking balance:", err)
		return false
	}

	if helper.ParseStringToFloat(balance.Available) >= amount {
		return true
	}

	return false
}


func getBalance(currency, baseURL, apiVersion, apiKey, apiSecret string) (Balance, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}

	url := baseURL + "/"+ apiVersion + "/spot/balance"
	if currency != "" {
		url += "/" + currency
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Balance{}, err
	}
	req.SetBasicAuth(apiKey, apiSecret)

	resp, err := client.Do(req)
	if err != nil {
		return Balance{}, err
	}
	defer resp.Body.Close()

	var balance Balance
	err = json.NewDecoder(resp.Body).Decode(&balance)
	if err != nil {
		return Balance{}, err
	}

	return balance, nil
}

func getBalances(baseURL, apiVersion, apiKey, apiSecret string) ([]Balance, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}

	url := baseURL + "/"+ apiVersion + "/spot/balance"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiKey, apiSecret)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var balances []Balance
	err = json.NewDecoder(resp.Body).Decode(&balances)
	if err != nil {
		return nil, err
	}

	return balances, nil
}

func getActiveOrders(symbol, baseURL, apiVersion, apiKey, apiSecret string) ([]Order, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}
	var orders []Order
	base := fmt.Sprintf("%s/%s/spot/", baseURL, apiVersion)
	url := base + "order"
	if symbol != "" {
		url += "?symbol=" + symbol
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return orders, err
	}
	req.SetBasicAuth(apiKey, apiSecret)

	resp, err := client.Do(req)
	if err != nil {
		return orders, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return orders, err
	}

	err = json.Unmarshal(body, &orders)
	if err != nil {
		return orders, err
	}

	return orders, nil
}

func createSpotOrder(symbol, side string, quantity, price float64, baseURL, apiVersion, apiKey, apiSecret string) (Order, error) {
	// Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}

	var order Order
	base := fmt.Sprintf("%s/%s/spot/", baseURL, apiVersion)
	urlb := base + "order"
	data := fmt.Sprintf(`{
		"symbol": "%s",
		"side": "%s",
		"quantity": "%f",
		"price": "%f",
		"timeInForce": "GTC",
		"type": "limit"
	}`, symbol, side, quantity, price)

	req, err := http.NewRequest("POST", urlb, strings.NewReader(data))
	if err != nil {
		return order, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(apiKey, apiSecret)

	resp, err := client.Do(req)
	if err != nil {
		return order, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return order, err
	}

	if resp.StatusCode != http.StatusOK {
		// Handle Error Response
		return order, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}

	err = json.Unmarshal(body, &order)
	if err != nil {
		return order, err
	}

	return order, nil
}

// Add more structs as needed for different types of responses.

// FetchHistoricalCandlesticks fetches historical candlestick data for the given symbol and time interval
func fetchHistoricalCandlesticks(symbol, baseURL, apiVersion, apiKey, interval string, startTime, endTime int64) ([]Candlestick, error) {
    // Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}

    // Calculate the CandleCount
	candleCount := helper.CalculateCandleCount(startTime, endTime, int64(helper.IntervalToDuration(interval)))
	fmt.Println("Interval:", interval, "CandleCount:", candleCount)
    // Connect to the HitBTC endpoint"
	url := fmt.Sprintf("%s/%s/public/candles/%s?period=%s&limit=%d", baseURL, apiVersion, symbol, interval, candleCount)
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
    
    // Parse the candlestick data from the received message
    var candles []Candlestick
    err = json.Unmarshal(body, &candles)
    if err != nil {
        log.Println("Error parsing candlestick data:", err)
        return nil, err
    }
    return candles, nil    
}
