package main

import (
	"log"
	"time"

	"os"
	"os/signal"
	"syscall"

	"coinbitly.com/server"
	"coinbitly.com/strategies"
)

//Initialize Trading System:
func main() {
	

	//You specify the source of data (e.g., "HitBTC", "Binance" or "InfluxDB")
	loadFrom := "Binance" 
	
	//You specify whether you're performing live trading 
	liveTrading := true

    //You're initializing your trading system using the strategies.NewTradingSystem function. 
	ts, err := strategies.NewTradingSystem(liveTrading, loadFrom)
	if err != nil {
		log.Fatal("Error initializing trading system:", err)
		return
	}

	//Perform Trading:
	//Depending on whether you're performing live trading or not, you're either calling the LiveTrade or Backtesting
	switch liveTrading{
	case true:
		go ts.LiveTrade(loadFrom)
	default:
		go ts.Backtest(loadFrom)
	}
	
	// Open or create a log file for appending
	logFile, err := os.OpenFile("./webclient/assets/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening or creating log file:", err)
		os.Exit(1)
	}
	
	// Create a logger that writes to the log file
	logger := log.New(logFile, "", log.LstdFlags)

	// Your application logic goes here

	logger.Println("App started. Press Ctrl+C to exit.")

	ts.Log = logger

	//Setup and Start Web Server:
	//You're setting up the web server by creating a server.NewTradeHandler() and server.NewServer(addrp, th) instance. 
	//Then, you open the server using the server.Open() method.
	addrp := os.Getenv("PORT")
	th := server.NewTradeHandler()
	server := server.NewServer(addrp, th)

	// Create a channel to listen for OS signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		// Wait for a termination signal
		log.Println(<-ts.ShutDownCh)

		// Perform any cleanup or shutdown tasks here

		// Close your server gracefully
		if err := server.Close(); err != nil {
			log.Printf("Error while closing server: %v", err)
		}

		// Close your data file gracefully
		if err := logFile.Close(); err != nil {
			log.Printf("Error while closing App file: %v", err)
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