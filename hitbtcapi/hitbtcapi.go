package hitbtcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
// Add more structs as needed for different types of responses.

// FetchHistoricalCandlesticks fetches historical candlestick data for the given symbol and time interval
func fetchHistoricalCandlesticks(symbol, baseURL, apiVersion, apiKey, interval string, startTime, endTime int64) ([]Candlestick, error) {
    // Initialize HTTP client with a custom transport to handle rate limiting
	client := &http.Client{
		Transport: helper.RateLimitedTransport{Base: http.DefaultTransport},
	}

    // Connect to the HitBTC websocket endpoint"
	url := fmt.Sprintf("%s/%s/public/candles/%s?period=%s&from=%d&to=%d", baseURL, apiVersion, symbol, interval, startTime, endTime)
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
    
    // Parse the candlestick data from the received message
    var candles []Candlestick
    err = json.Unmarshal(body, &candles)
    if err != nil {
        log.Println("Error parsing candlestick data:", err)
        return nil, err
    }

    // Process the received candlestick data
    for _, candle := range candles {
        fmt.Printf(
            "HITBTC CANDLE\n Timestamp: %s, Open: %f, Close: %f, High: %f, Low: %f, Volume: %f, Candle Count: %d\n",
            candle.Timestamp, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume, candle.CandleCount,
        )
    }
    return candles, nil    
}
