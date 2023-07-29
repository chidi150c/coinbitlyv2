package main

import (
	"fmt"
	"strconv"
	"time"

	"coinbitly.com/binanceapi"
	"coinbitly.com/influxdb"
)

func main() {
	// Initialize and Connect to InfluxDB
	client, err := influxdb.NewDB()
	if err != nil {
		fmt.Println("Error Connecting to InfluxDB:", err)
		return
	}
	
	// Close the InfluxDB client before exiting the application
	defer client.Close()

	// Create a write API
	writeDB := client.WriteAPI(influxdb.InfluxDBOrgID, "market_data") // Use "market_data" as the measurement name
	Exchange := "Binance"
	// List of cryptocurrency symbols to fetch data for
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "XRPUSDT"}

	// Fetch and store market data for each symbol in a loop
	for _, symbol := range symbols {
		// Fetch and display real-time market data
		fmt.Println("Real-Time Market Data for", symbol, ":")
			ticker, err := binanceapi.FetchAndDisplayTickerData(symbol)
		if err != nil {
			fmt.Println("Error Fetching real-time market data for", symbol, ":", err)
			continue
		}
		fmt.Println()

		// Fetch and display 24-hour price change statistics
		fmt.Println("24-Hour Price Change Statistics for", symbol, ":")
		ticker24C, err := binanceapi.FetchAndDisplay24hrTickerData(symbol)
		if err != nil {
			fmt.Println("Error Fetching 24-hour price change statistics for", symbol, ":", err)
				continue
		}
		fmt.Println()

		// Fetch and display order book depth
		fmt.Println("Order Book Depth for", symbol, ":")
		orderBook, err := binanceapi.FetchAndDisplayOrderBook(symbol, 5) // Fetch top 5 bids and asks
		if err != nil {
			fmt.Println("Error Fetching order book depth for", symbol, ":", err)
			continue
		}

		// Convert the price from string to float64
		price, err := strconv.ParseFloat(ticker.Price, 64)
		if err != nil {
			fmt.Println("Error converting price to float64 for", symbol, ":", err)
			continue
		}

		// Convert the ticker24C.Volume from string to float64
		volume, err := strconv.ParseFloat(ticker24C.Volume, 64)
		if err != nil {
			fmt.Println("Error converting volume to float64 for", symbol, ":", err)
			continue
		}

		// Create a new MarketData instance for the symbol
		marketData := influxdb.NewMarketData(Exchange, symbol)
		
		// Iterate through the top 5 bids and asks and store them in MarketData struct
		for i := 0; i < 5; i++ {
			// Convert the bid and ask prices from string to float64
			bidPrice, err := strconv.ParseFloat(orderBook.Bids[i][0], 64)
			if err != nil {
				fmt.Println("Error converting bid price to float64 for", symbol, ":", err)
				continue
			}
			askPrice, err := strconv.ParseFloat(orderBook.Asks[i][0], 64)
			if err != nil {
				fmt.Println("Error converting ask price to float64 for", symbol, ":", err)
				continue
			}

			// Convert the bid and ask quantities from string to float64
			bidQuantity, err := strconv.ParseFloat(orderBook.Bids[i][1], 64)
			if err != nil {
				fmt.Println("Error converting bid quantity to float64 for", symbol, ":", err)
				continue
			}
			askQuantity, err := strconv.ParseFloat(orderBook.Asks[i][1], 64)
			if err != nil {
				fmt.Println("Error converting ask quantity to float64 for", symbol, ":", err)
				continue
			}

			// Assign the converted values to the MarketData struct
			marketData.BidPrice = bidPrice
			marketData.AskPrice = askPrice
			marketData.BidQuantity = bidQuantity
			marketData.AskQuantity = askQuantity
			marketData.Price = price
			marketData.Volume = volume
			marketData.Exchange = Exchange
	
			// Assign the converted price to the MarketData struct
			marketData.Timestamp = time.Now()

			// Write the data point to InfluxDB
			err = influxdb.WriteDataPoint(writeDB, marketData)
			if err != nil {
				fmt.Println("Error writing data point:", err)
				return
			}
		}
	}	
	fmt.Println("Data point written successfully!")
}

