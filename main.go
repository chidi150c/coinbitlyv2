package main

import (
	"fmt"
	"log"

	"coinbitly.com/strategies"
)

func main() {
	// Initialize the trading system.
	ts := &strategies.TradingSystem{
		ShortPeriod:        12, // Define moving average short period for the strategy.
		LongPeriod:         26, // Define moving average long period for the strategy.
	}

    err := ts.Initialize()
	if err != nil {
		log.Fatal("Error initializing trading system:", err)
		return
	}
		
	// Perform backtesting with the given data and parameters.
	strategies.Backtest(ts)

	// Output the final capital after backtesting.
	fmt.Printf("\nFinal Capital after Backtesting: %.2f\n", ts.CurrentBalance)
	fmt.Printf("\nFinal Capital after Backtesting: %v\n", ts.Signals)
}