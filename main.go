package main

import (
	"log"
	
	"os"
	"os/signal"
	"syscall"

	"coinbitly.com/strategies"
	"coinbitly.com/server"
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
		<-signalCh

		log.Println("Received termination signal. Shutting down...")

		// Perform any cleanup or shutdown tasks here

		// Close your server gracefully
		if err := server.Close(); err != nil {
			log.Printf("Error while closing server: %v", err)
		}

		log.Println("Server shut down gracefully.")
		os.Exit(0)
	}()

	//Start the webserver
	if err := server.Open(); err != nil {
		log.Fatalf("Unable to Open Server for listen and serve: %v", err)
	}
}