package main

import (
	"fmt"
	"log"

	"coinbitly.com/strategies"
)

func main() {
	loadFrom := "HitBTC"
    ts, err := strategies.NewTradingSystem(loadFrom)
	if err != nil {
		log.Fatal("Error initializing trading system:", err)
		return
	}
		
	// Perform backtesting with the given data and parameters.
	fmt.Println()
	ts.Backtest(loadFrom)
 
	// // Define EMA periods for the crossover strategy as multiple sets
	// emaPeriods := [][]int{{6, 8}, {6, 12}, {6, 13}, {6, 14}, {6, 15}, {6, 16},{6, 17}, {6, 18}, {6, 19}, {6, 20},
	// {12, 26},{12, 27},{12,28}, {12, 29}, {12, 30},{12, 31},{12,32}, {12, 33}, {12, 34},
	// 					}

	// // Calculate profit for each set of EMA periods
	// for _, periods := range emaPeriods {
	// 	EMAProfit, MACDProfit, RSIProfit := ts.EvaluateEMAParameters(ts.ClosingPrices, periods)
	// 	fmt.Printf("Periods: %d,%d EMAProfit: %.2f, MACDProfit %.2f, RSIProfit %.2f\n", periods[0], periods[1], EMAProfit, MACDProfit, RSIProfit)
	// }
}