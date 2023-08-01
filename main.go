package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	"reflect"
	"coinbitly.com/binanceapi"
	"coinbitly.com/config"
	"coinbitly.com/hitbtcapi"
	"coinbitly.com/influxdb"
	"coinbitly.com/model"
)

//getExchanges retrieves each exhange configuration from the exchangesConfig and return them as slice of echange 
func getExchanges(exchangesConfig map[string]*config.ExchConfig)[]model.ExchServices{
	binanceExch, err := binanceapi.NewExchServices(exchangesConfig)
	if err != nil{
		log.Fatalf("ERROR: %v", err)
	}
	hitbtcEch, err := hitbtcapi.NewExchServices(exchangesConfig)
	if err != nil{
		log.Fatalf("ERROR: %v", err)
	}
	return []model.ExchServices{
		binanceExch,
		hitbtcEch,
	}
}

func main() {
	//Intiallizing the Exchanges
	exchangesConfig := config.NewExchangesConfig()

	//Get instance of an Exchanges
	exchanges := getExchanges(exchangesConfig)

	// Initialize and Connect to InfluxDB
	client, ctx, err := influxdb.NewDB()
	if err != nil {
		fmt.Println("Error Connecting to InfluxDB:", err)
		return
	}
	
	// Close the InfluxDB client before exiting the application
	defer client.Close()

	interval := "M1"	
	// Define the time interval for historical data (e.g., last 30 days)
	startTime := time.Now().Add(-1 * time.Hour).Unix() // 30 days ago
	endTime := time.Now().Unix()
	Symbols := []string{"BTCUSDT"}
	// Fetch and store of each symbol in a loop
	for _, exchange := range exchanges{
		for _, symbol := range Symbols {
			// Fetch historical candlestick data
			fmt.Println("Historical Candlestick Data for", reflect.TypeOf(exchange), symbol, ":")
			candlesticks, err := exchange.FetchHistoricalCandlesticks(symbol, interval, startTime, endTime)
			if err != nil {
				fmt.Println("Error Fetching historical candlestick data for", symbol, ":", err)
				continue
			}
			// Iterate through the fetched candlesticks and store them in InfluxDB
			for _, cs := range candlesticks {
				// Create a new MarketData instance for the symbol
				marketData := influxdb.NewMarketData(cs.ExchName, symbol)
				// Convert the candlestick data to MarketData struct
				marketData.BidPrice = cs.Close
				marketData.AskPrice = cs.Close
				marketData.BidQuantity = cs.Close
				marketData.AskQuantity = cs.Close
				marketData.Exchange = cs.ExchName		
				marketData.Symbol = symbol
				marketData.Price = cs.Close
				marketData.Volume = cs.Volume
				marketData.Timestamp = time.Unix(int64(cs.Timestamp)/1000, 0)

				// Print the values of marketData for debugging purposes
				fmt.Println("Market Data for" , symbol, ":", marketData)

				// Write the data point to InfluxDB
				err = influxdb.WriteDataPoint(client, ctx, marketData)
				if err != nil {
					fmt.Println("Error writing data point:", err)
					return
				}
			}
		}	
	}
	fmt.Println("Data point written successfully!")
	// Wait for a termination signal to gracefully exit
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan

	influxdb.ReadDataPoints(client, ctx)
}

