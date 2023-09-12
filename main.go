package main

import (
	"encoding/csv"
	"log"
	"time"

	"os"
	"os/signal"
	"syscall"

	"coinbitly.com/config"
	"coinbitly.com/server"
	"coinbitly.com/strategies"
)

//Initialize Trading System:
func main() {
	

	//You specify the source of Exch API (e.g., "HitBTC", "Binance", "BinanceTestnet")
	loadExchFrom := "BinanceTestnet" 
	//You specify the source of DataBase (e.g., "InfluxDB")
	loadDBFrom :=  "InfluxDB"	
	//You specify whether you're performing live trading or not 
	liveTrading := true

	config := config.NewExchangeConfigs()[loadExchFrom]
    //You're initializing your trading system using the strategies.NewTradingSystem function. 
	ts, err := strategies.NewTradingSystem(config.BaseCurrency, liveTrading, loadExchFrom, loadDBFrom)
	if err != nil {
		log.Fatal("Error initializing trading system:", err)
		return
	}

	// Open or create a log file for appending
	logFile, err := os.OpenFile("./webclient/assets/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening or creating log file:", err)
		os.Exit(1)
	}
	
	// Create a logger that writes to the log file
	ts.Log = log.New(logFile, "", log.LstdFlags)

	// Save data to CSV file
    appfile, err := os.OpenFile("./webclient/assets/dataPoint.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }  
	
    ts.CSVWriter = csv.NewWriter(appfile)
	// Write headers to the CSV file
	headers := []string{
		"Count","Strategy","ShortPeriod","LongPeriod", "ShortEMA", "LongEMA", "TargetProfit",
		"TargetStopLoss","RiskPositionPercentage","TotalProfitLoss",   
	}
	err = ts.CSVWriter.Write(headers)
	if err != nil {
		log.Fatal(err)
	}

	ts.Log.Println("started.................................")

	//Depending on whether you're performing live trading or not, you're either calling the LiveTrade or Backtesting
	switch liveTrading{
	case true:
		go ts.LiveTrade(loadExchFrom)  //, loadDBFrom)
	default:
		go ts.Backtest(loadExchFrom)  //, loadDBFrom)
	}
	//Setup and Start Web Server:
	//You're setting up the web server by creating a server.NewTradeHandler() and server.NewServer(addrp, th) instance. 
	//Then, you open the server using the server.Open() method.
	addrp := os.Getenv("PORT4")
	hostsite := os.Getenv("HOSTSITE")
	th := server.NewTradeHandler(ts, hostsite)
	server := server.NewServer(addrp, th)

	// Create a channel to listen for OS signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		// ts.Mu.Lock()
		// defer ts.Mu.Unlock()
		// Wait for a termination signal
		log.Println(<-ts.ShutDownCh)

		// Perform any cleanup or shutdown tasks here

		// Close your server gracefully
		if err := server.Close(); err != nil {
			log.Printf("Error while closing server: %v", err)
		}

		// Close your Log file gracefully
		if err := logFile.Close(); err != nil {
			log.Printf("Error while closing Log file: %v", err)
		}

		ts.Mu.Lock()
		ts.CSVWriter.Flush()
		ts.Mu.Unlock()

		// Close your AppData file gracefully
		if err := appfile.Close(); err != nil {
			log.Printf("Error while closing AppData file: %v", err)
		}
		
		log.Println("Server shut down gracefully.")
		time.Sleep(time.Second * 6) 
		os.Exit(0)
	}()

	//Start the webserver
	if err := server.Open(); err != nil {
		log.Fatalf("Unable to Open Server for listen and serve: %v", err)
	}
}