package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
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

	// Create a channel to receive the SIGINT signal
	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, os.Interrupt, syscall.SIGTERM)
	// Create a channel to signal the main goroutine to exit
	doneCh := make(chan struct{})
	// Go routine to handle the SIGINT signal
	go func() {
		<-sigintCh
		fmt.Println("\nReceived SIGINT signal. Flushing data and exiting...")

		// Flush the data to InfluxDB before exiting the application
		client.WriteAPI(influxdb.InfluxDBOrgID, "newmarket_data").Flush()

		// Give some time for the data to be flushed
		time.Sleep(1 * time.Second)

		// Exit the application
		os.Exit(0)
		
		// Signal the main goroutine to exit
		doneCh <- struct{}{}
	}()

	Exchange := "Binance"
	interval := "5m"
	// List of cryptocurrency symbols to fetch data for
	symbols := []string{"BTCUSDT"}

	// Define the time interval for historical data (e.g., last 30 days)
	startTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	endTime := time.Now()

	// Fetch and store historical market data for each symbol in a loop
	for _, symbol := range symbols {
		// Fetch historical candlestick data
		fmt.Println("Historical Candlestick Data for", symbol, ":")
		candlesticks, err := binanceapi.FetchHistoricalCandlesticks(symbol, interval, startTime, endTime)
		if err != nil {
			fmt.Println("Error Fetching historical candlestick data for", symbol, ":", err)
			continue
		}

		// Iterate through the fetched candlesticks and store them in InfluxDB
		for _, cs := range candlesticks {
			// Convert the candlestick data to MarketData struct
			marketData := influxdb.NewMarketData(Exchange, symbol)
			marketData.Price = cs.Close
			marketData.Volume = cs.Volume
			marketData.Timestamp = time.Unix(int64(cs.Timestamp)/1000, 0)

			// Print the values of marketData for debugging purposes
			fmt.Println("Market Data for", symbol, ":", marketData)

			// Write the data point to InfluxDB
			err = influxdb.WriteDataPoint(client, marketData)
			if err != nil {
				fmt.Println("Error writing data point:", err)
				return
			}
		}
	}
	fmt.Println("Data point written successfully!")
	<-doneCh
	influxdb.DeleteBucket(client)
}
