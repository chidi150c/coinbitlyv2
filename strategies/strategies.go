package strategies

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"

	"sync"

	"coinbitly.com/config"
	"coinbitly.com/hitbtcapi"
	"coinbitly.com/influxdb"
	"coinbitly.com/model"
	"github.com/apourchet/investment/lib/ema"
	"github.com/pkg/errors"
)

// TradingSystem struct: The TradingSystem struct represents the main trading
// system and holds various parameters and fields related to the strategy,
// trading state, and performance.
type TradingSystem struct {
	mu              sync.Mutex // Mutex for protecting concurrent writes
	HistoricalData  []model.Candle
	ClosingPrices   []float64
	Container1      []float64
	Container2      []float64
	Timestamps      []int64
	Signals         []string
	APIServices 	model.APIServices
	TransactionCost float64
	Slippage        float64
	InitialCapital  float64
	PositionSize    float64
	EntryPrice      float64
	InTrade         bool
	QuoteBalance    float64
	BaseBalance     float64
	RiskCost        float64
	DataPoint       int
	CurrentPrice    float64
	EntryQuantity   float64
	// StopLossRecover        float64
	// RiskFactor 			   float64
}

// NewTradingSystem(): This function initializes the TradingSystem and fetches
// historical data from the exchange using the GetCandlesFromExch() function.
// It sets various strategy parameters like short and long EMA periods, RSI period,
// MACD signal period, Bollinger Bands period, and more.
func NewTradingSystem(loadFrom string) (*TradingSystem, error) {
	// Initialize the trading system.
	ts := &TradingSystem{}
	// ts.RiskFactor = 2.0           // Define 1% slippage
	ts.TransactionCost = 0.001 // Define 0.1% transaction cost
	ts.Slippage = 0.01
	ts.InitialCapital = 1000.0          //Initial Capital for simulation on backtesting
	ts.QuoteBalance = ts.InitialCapital //continer to hold the balance
	// ts.StopLossRecover = math.MaxFloat64 //

	// Fetch historical data from the exchange
	err := ts.UpdateHistoricalData(loadFrom)
	if err != nil {
		return &TradingSystem{}, err
	}
	// Perform any other necessary initialization steps here
	ts.Signals = make([]string, len(ts.ClosingPrices))   // Holder of generated trading signals
	ts.Timestamps = make([]int64, len(ts.ClosingPrices)) // Holder of generated trading signals

	return ts, nil
}

func NewAppData() *model.AppData {
	md := &model.AppData{
		ShortPeriod:     6,  // Define moving average short period for the strategy.
		LongPeriod:      16, // Define moving average long period for the strategy.
		ShortMACDPeriod: 12,
		LongMACDPeriod:  29,
	}
	md.Count = 0
	md.Scalping = ""
	md.TargetProfit = 5.0
	md.TargetStopLoss = 3.0
	md.RiskPositionPercentage = 0.10 // Define risk management parameter 5% balance
	md.RSIPeriod = 14                // Define RSI period parameter. 12,296,16
	md.StochRSIPeriod = 3
	md.SmoothK = 3
	md.SmoothD = 3
	md.StRSIOverbought = 0.8 // Define overbought for generating RSI signals
	md.StRSIOversold = 0.2
	md.RSIOverbought = 0.6      // Define overbought for generating RSI signals
	md.RSIOversold = 0.4        // Define oversold for generating RSI signals
	md.SignalMACDPeriod = 9     // Define MACD period parameter.
	md.BollingerPeriod = 20     // Define Bollinger Bands parameter.
	md.BollingerNumStdDev = 2.0 // Define Bollinger Bands parameter.
	md.Strategy = "MACD"
	md.StrategyCombLogic = ""
	md.TotalProfitLoss = 0.0
	return md
}

// Backtest(): This function simulates the backtesting process using historical
// price data. It iterates through the closing prices, checks for entry and exit
// conditions, and executes trades accordingly. It also tracks trading performance
// and updates the current balance after each trade.
func (ts *TradingSystem) Backtest(loadFrom string) {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)
	//Count,Strategy,StrategyCombLogic,ShortPeriod,LongPeriod,ShortMACDPeriod,LongMACDPeriod,
	//SignalMACDPeriod,RSIPeriod,StochRSIPeriod,SmoothK,SmoothD,RSIOverbought,RSIOversold,
	//StRSIOverbought,StRSIOversold,BollingerPeriod,BollingerNumStdDev,TargetProfit,
	//TargetStopLoss,RiskPositionPercentage,Scalping,
	backT := []*model.AppData{ //StochRSI
		{1, "MACD", "", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {2, "RSI", "", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {3, "EMA", "", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {4, "StochR StrategiesOnly", "", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {5, "Bollinger DataStrategies", "", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {6, "MACD,RSI", "AND", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {7, "RSI,MACD", "AND", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {8, "EMA,MACD", "AND", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {9, "StochR,MACD", "AND", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {10, "Bollinger,MACD", "AND", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {11, "MACD,RSI", "OR", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {12, "RSI,MACD", "OR", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {13, "EMA,MACD", "OR", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {14, "StochR,MACD", "OR", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
		// {15, "Bollinger,MACD", "OR", 6, 16, 12, 29, 9, 14, 14, 3, 3, 0.6, 0.4, 0.7, 0.2, 20, 2.0, 1.0, 0.5, 0.1, "", 0.0},
	}
	AppDatatoCSV(backT)
	for _, md := range backT {
		// Initialize variables for tracking trading performance.
		tradeCount := 0
		// Simulate the backtesting process using historical price data.
		for ts.DataPoint, ts.CurrentPrice = range ts.ClosingPrices {
			// Execute the trade if entry conditions are met.
			if (!ts.InTrade) && (ts.EntryRule(md)) { //&& (ts.CurrentPrice <= ts.StopLossRecover)
				// Record entry price for calculating profit/loss and stoploss later.
				ts.EntryPrice = ts.CurrentPrice

				// Execute the buy order using the ExecuteStrategy function.
				err := ts.ExecuteStrategy(md, "Buy")
				if err != nil {
					fmt.Println("Error executing buy order:", err)
					ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position
					continue
				}

				// Mark that we are in a trade.
				ts.InTrade = true
				// ts.StopLossRecover = math.MaxFloat64

				tradeCount++
				fmt.Printf("- BUY at %v Quant: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f NoBuyLoss: %.8f PosPcent: %.8f DataPt: %d\n", ts.CurrentPrice, ts.EntryQuantity, ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, 0.0, md.RiskPositionPercentage, ts.DataPoint)
				// Close the trade if exit conditions are met.
			} else if ts.InTrade {
				switch md.Scalping {
				case "UseTA":
					if ts.ExitRule(md) {
						// Execute the sell order using the ExecuteStrategy function.
						err := ts.ExecuteStrategy(md, "Sell")
						if err != nil {
							if fmt.Sprintf("%v", err) == "Stoploss Triggered" {

								ts.InTrade = false

								tradeCount++
								// fmt.Println(err, ": StoplossRecover", ts.StopLossRecover)
								continue
							}
							// fmt.Println("Error:", err, " at:", ts.CurrentPrice, ", Stoploss below::", ts.StopLossPrice)

							ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position

							continue
						}
					}
				default:
					// Execute the sell order using the ExecuteStrategy function.
					err := ts.ExecuteStrategy(md, "Sell")
					if err != nil {
						if fmt.Sprintf("%v", err) == "Stoploss Triggered" {

							ts.InTrade = false

							tradeCount++
							// fmt.Println(err, ": StoplossRecover", ts.StopLossRecover)
							continue
						}
						// fmt.Println("Error:", err, " at:", ts.CurrentPrice, ", Stoploss below::", ts.StopLossPrice)

						ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position

						continue
					}

				}
				fmt.Printf("- SELL at %v Quant: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f Loss: %.8f PosPcent: %.8f DataPt: %d\n", ts.CurrentPrice, ts.EntryQuantity, ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, md.TargetStopLoss, md.RiskPositionPercentage, ts.DataPoint)
				// Mark that we are no longer in a trade.
				ts.InTrade = false
				tradeCount++
				// Implement risk management and update capital after each trade
				// For simplicity, we'll skip this step in this basic example
			} else {
				ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position
			}

			if  loadFrom != "InfluxDB" {
				err := ts.APIServices.WriteCandleToDB(ts.HistoricalData[ts.DataPoint])
				if (err != nil) && (!strings.Contains(fmt.Sprintf("%v", err), "Skipping write")) {
					log.Fatalf("Error: writing to influxDB: %v", err)
				}
			}
		}
		//Check if there is still asset remainning and sell off
		if ts.BaseBalance > 0.0 {
			// After sell off Update the quote and base balances after the trade.

			// Calculate profit/loss for the trade.
			exitPrice := ts.CurrentPrice
			tradeProfitLoss := CalculateProfitLoss(ts.EntryPrice, exitPrice, ts.BaseBalance)
			transactionCost := ts.TransactionCost * exitPrice * ts.BaseBalance
			slippageCost := ts.Slippage * exitPrice * ts.BaseBalance

			// Store profit/loss for the trade.

			tradeProfitLoss -= transactionCost + slippageCost
			md.TotalProfitLoss += tradeProfitLoss

			ts.QuoteBalance += (ts.BaseBalance * exitPrice) - transactionCost - slippageCost
			ts.BaseBalance -= ts.BaseBalance
			ts.Timestamps = append(ts.Timestamps, ts.Timestamps[len(ts.Timestamps)-1])
			ts.ClosingPrices = append(ts.ClosingPrices, ts.ClosingPrices[len(ts.ClosingPrices)-1])
			ts.Signals = append(ts.Signals, "Sell")

			fmt.Printf("- SELL-Off at %v Quant: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f LossCal: %.8f PosPcent: %.8f DataPt: %d\n", ts.CurrentPrice, ts.BaseBalance, ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, tradeProfitLoss, md.RiskPositionPercentage, ts.DataPoint)
		}

		// Print the overall trading performance after backtesting.
		fmt.Printf("\nBacktesting Summary: %d Strategy: %s Combination: %s\n", md.Count, md.Strategy, md.StrategyCombLogic)
		fmt.Printf("Total Trades: %d, ", tradeCount)
		fmt.Printf("Total Profit/Loss: %.2f, ", md.TotalProfitLoss)
		fmt.Printf("Final Capital: %.2f, ", ts.QuoteBalance)
		fmt.Printf("Final Asset: %.8f ", ts.BaseBalance)
		var err error

		// if len(ts.Container1) > 0 && (!strings.Contains(md.Strategy, "Bollinger")) && (!strings.Contains(md.Strategy, "EMA")) && (!strings.Contains(md.Strategy, "MACD")) {
		// 	err = CreateLineChartWithSignals(ts.Timestamps, ts.ClosingPrices, ts.Signals, "DataOnly")
		// 	err = CreateLineChartWithSignals(ts.Timestamps, ts.Container1, ts.Signals, "StrategyOnly")
		// } else if len(ts.Container1) > 0 && strings.Contains(md.Strategy, "Bollinger") {
		// 	err = CreateLineChartWithSignalsV3(ts.Timestamps, ts.ClosingPrices, ts.Container1, ts.Container2, ts.Signals, "DataStrategies")
		// } else if len(ts.Container1) > 0 && strings.Contains(md.Strategy, "EMA") {
		// 	err = CreateLineChartWithSignalsV3(ts.Timestamps, ts.ClosingPrices, ts.Container1, ts.Container2, ts.Signals, "DataStrategies")
		// } else if len(ts.Container1) > 0 && strings.Contains(md.Strategy, "MACD") {
		// 	err = CreateLineChartWithSignals(ts.Timestamps, ts.ClosingPrices, ts.Signals, "DataOnly")
		// 	err = CreateLineChartWithSignalsV2(ts.Timestamps, ts.Container1, ts.Container2, ts.Signals, "StrategiesOnly")
		// } else {
		// 	err = CreateLineChartWithSignals(ts.Timestamps, ts.ClosingPrices, ts.Signals, "DataOnly")
		// }
		if err != nil {
			fmt.Println("Error creating Line Chart with signals:", err)
			return
		}
		<-sigchnl
		fmt.Println()
		fmt.Println()
		fmt.Println("Next Set Starts Below:")
	}
}

// TechnicalAnalysis(): This function performs technical analysis using the
// calculated moving averages (short and long EMA), RSI, MACD line, and Bollinger
// Bands. It determines the buy and sell signals based on various strategy rules.
func (ts *TradingSystem) TechnicalAnalysis(md *model.AppData) (buySignal, sellSignal bool) {
	// Calculate moving averages (MA) using historical data.
	longEMA, shortEMA, timeStamps, err := ts.CandleExponentialMovingAverage(md, 0, 0)
	if err != nil {
		log.Fatalf("Error: in TechnicalAnalysis Unable to get EMA: %v", err)
	}
	// Calculate Relative Strength Index (RSI) using historical data.
	rsi := CalculateRSI(ts.ClosingPrices, md.RSIPeriod)

	// Calculate Stochastic RSI and SmoothK, SmoothD values
	stochRSI, smoothKRSI := StochasticRSI(rsi, md.StochRSIPeriod, md.SmoothK, md.SmoothD)

	// Calculate MACD and Signal Line using historical data.
	macdLine, signalLine, macdHistogram := ts.CalculateMACD(md, md.LongMACDPeriod, md.ShortMACDPeriod)

	// Calculate Bollinger Bands using historical data.
	_, upperBand, lowerBand := CalculateBollingerBands(ts.ClosingPrices, md.BollingerPeriod, md.BollingerNumStdDev)

	// Determine the buy and sell signals based on the moving averages, RSI, MACD line, and Bollinger Bands.
	if len(stochRSI) > 1 && len(smoothKRSI) > 1 && len(shortEMA) > 1 && len(longEMA) > 1 && len(rsi) > 0 && len(macdLine) > 1 && len(signalLine) > 1 && len(upperBand) > 0 && len(lowerBand) > 0 {
		// Generate buy/sell signals based on conditions
		count := 0
		switch md.StrategyCombLogic {
		case "AND":
			buySignal, sellSignal = true, true
			if strings.Contains(md.Strategy, "StochR") && (ts.DataPoint < len(stochRSI)) && ts.DataPoint >= (md.StochRSIPeriod+md.SmoothK+md.SmoothD-1) {
				count++
				buySignal = buySignal && ((stochRSI[ts.DataPoint-1] < smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] >= smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint] < md.StRSIOversold))
				sellSignal = sellSignal && ((stochRSI[ts.DataPoint-1] >= smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] < smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint] > md.StRSIOverbought))
				// buySignal = buySignal && (stochRSI[ts.DataPoint] > ts.StRSIOverbought && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] > ts.StRSIOverbought)
				// sellSignal = sellSignal && (stochRSI[ts.DataPoint] < md.StRSIOversold && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] < md.StRSIOversold)
			}
			if strings.Contains(md.Strategy, "EMA") && ts.DataPoint > 0 {
				count++
				buySignal = buySignal && (shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint])
				sellSignal = sellSignal && (shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint])
			}
			if strings.Contains(md.Strategy, "RSI") && ts.DataPoint > 0 && ts.DataPoint <= len(rsi) {
				count++
				buySignal = buySignal && (rsi[ts.DataPoint-1] < md.RSIOversold)
				sellSignal = sellSignal && (rsi[ts.DataPoint-1] > md.RSIOverbought)
			}
			if strings.Contains(md.Strategy, "MACD") && ts.DataPoint > 0 {
				count++
				buySignal = buySignal && (macdLine[ts.DataPoint-1] <= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] > signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] > 0)
				sellSignal = sellSignal && (macdLine[ts.DataPoint-1] >= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] < signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] < 0)
			}
			if strings.Contains(md.Strategy, "Bollinger") && ts.DataPoint > 0 && ts.DataPoint < len(upperBand) {
				count++
				buySignal = buySignal && (ts.ClosingPrices[len(ts.ClosingPrices)-1] < lowerBand[ts.DataPoint])
				sellSignal = sellSignal && (ts.ClosingPrices[len(ts.ClosingPrices)-1] > upperBand[ts.DataPoint])
			}
		case "OR":
			if strings.Contains(md.Strategy, "StochR") && (ts.DataPoint < len(stochRSI)) && ts.DataPoint >= (md.StochRSIPeriod+md.SmoothK+md.SmoothD-1) {
				count++
				buySignal = buySignal || ((stochRSI[ts.DataPoint-1] < smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] >= smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint] < md.StRSIOversold))
				sellSignal = sellSignal || ((stochRSI[ts.DataPoint-1] >= smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] < smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint] < md.StRSIOversold))
				// buySignal = buySignal || (stochRSI[ts.DataPoint] > ts.StRSIOverbought && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] > ts.StRSIOverbought)
				// sellSignal = sellSignal || (stochRSI[ts.DataPoint] < md.StRSIOversold && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] < md.StRSIOversold)
			}
			if strings.Contains(md.Strategy, "EMA") && ts.DataPoint > 0 {
				count++
				buySignal = buySignal || (shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint])
				sellSignal = sellSignal || (shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint])
			}
			if strings.Contains(md.Strategy, "RSI") && ts.DataPoint > 0 && ts.DataPoint <= len(rsi) {
				count++
				buySignal = buySignal || (rsi[ts.DataPoint-1] < md.RSIOversold)
				sellSignal = sellSignal || (rsi[ts.DataPoint-1] > md.RSIOverbought)
			}
			if strings.Contains(md.Strategy, "MACD") && ts.DataPoint > 0 {
				count++
				buySignal = buySignal || (macdLine[ts.DataPoint-1] <= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] > signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] > 0)
				sellSignal = sellSignal || (macdLine[ts.DataPoint-1] >= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] < signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] < 0)
			}
			if strings.Contains(md.Strategy, "Bollinger") && ts.DataPoint > 0 && ts.DataPoint < len(upperBand) {
				count++
				buySignal = buySignal || (ts.ClosingPrices[len(ts.ClosingPrices)-1] < lowerBand[ts.DataPoint])
				sellSignal = sellSignal || (ts.ClosingPrices[len(ts.ClosingPrices)-1] > upperBand[ts.DataPoint])
			}
		default:
			if strings.Contains(md.Strategy, "StochR") && (ts.DataPoint < len(stochRSI)) && ts.DataPoint >= (md.StochRSIPeriod+md.SmoothK+md.SmoothD-1) {
				count++
				ts.Container1 = stochRSI
				ts.Container2 = smoothKRSI
				buySignal = (stochRSI[ts.DataPoint-1] < smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] >= smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint] < md.StRSIOversold)
				sellSignal = (stochRSI[ts.DataPoint-1] >= smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] < smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint] < md.StRSIOversold)
				// buySignal = stochRSI[ts.DataPoint] > ts.StRSIOverbought && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] > ts.StRSIOverbought)
				// sellSignal = stochRSI[ts.DataPoint] < md.StRSIOversold && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] < md.StRSIOversold)
			}
			if strings.Contains(md.Strategy, "EMA") && ts.DataPoint > 0 {
				count++
				ts.Container1 = shortEMA
				ts.Container2 = longEMA
				buySignal = shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint]
				sellSignal = shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint]
			}
			if strings.Contains(md.Strategy, "RSI") && ts.DataPoint > 0 && ts.DataPoint <= len(rsi) {
				count++
				ts.Container1 = rsi
				buySignal = rsi[ts.DataPoint-1] < md.RSIOversold
				sellSignal = rsi[ts.DataPoint-1] > md.RSIOverbought
			}
			if strings.Contains(md.Strategy, "MACD") && ts.DataPoint > 0 {
				count++
				ts.Container1 = signalLine
				ts.Container2 = macdLine
				buySignal = macdLine[ts.DataPoint-1] <= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] > signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] > 0
				sellSignal = macdLine[ts.DataPoint-1] >= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] < signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] < 0
			}
			if strings.Contains(md.Strategy, "Bollinger") && ts.DataPoint > 0 && ts.DataPoint < len(upperBand) {
				count++
				ts.Container1 = lowerBand
				ts.Container2 = upperBand
				buySignal = buySignal || ts.ClosingPrices[len(ts.ClosingPrices)-1] < lowerBand[ts.DataPoint]
				sellSignal = sellSignal || ts.ClosingPrices[len(ts.ClosingPrices)-1] > upperBand[ts.DataPoint]
			}
		}

		if count == 0 {
			buySignal, sellSignal = false, false
		}
		_ = timeStamps
		// PlotEMA(timeStamps, stochRSI, smoothKRSI)
		// PlotRSI(stochRSI[290:315], smoothKRSI[290:315])

		// fmt.Printf("EMA: %v, RSI: %v, MACD: %v, Bollinger: %v buySignal: %v, sellSignal: %v", md.Strategy.UseEMA, md.Strategy.UseRSI, md.Strategy.UseMACD, md.Strategy.UseBollinger, buySignal, sellSignal)
	}

	return buySignal, sellSignal
}

// ExecuteStrategy executes the trade based on the provided trade action and current price.
// The tradeAction parameter should be either "Buy" or "Sell".
func (ts *TradingSystem) ExecuteStrategy(md *model.AppData, tradeAction string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	// Calculate position size based on the fixed percentage of risk per trade.
	isStopLossExecuted := ts.RiskManagement(md) // make sure entry price is set before calling risk management
	if isStopLossExecuted {
		return errors.New("Stoploss Triggered")
	}
	switch tradeAction {
	case "Buy":
		if ts.InTrade {
			return fmt.Errorf("cannot execute a buy order while already in a trade")
		}

		// Update the current balance after the trade (considering transaction fees and slippage).
		transactionCost := ts.TransactionCost * ts.CurrentPrice * ts.PositionSize
		slippageCost := ts.Slippage * ts.CurrentPrice * ts.PositionSize

		//check if there is enough Quote (USDT) for the buy transaction
		if ts.QuoteBalance < (ts.PositionSize*ts.CurrentPrice)+transactionCost+slippageCost {
			return fmt.Errorf("cannot execute a buy order due to insufficient QuoteBalance: %.8f (ts.PositionSize %.8f* ts.CurrentPrice %.8f) = %.8f + transactionCost %.8f+ slippageCost %.8f", ts.QuoteBalance, ts.PositionSize, ts.CurrentPrice, ts.PositionSize*ts.CurrentPrice, transactionCost, slippageCost)
		}

		// Record Signal for plotting graph later
		ts.Signals[ts.DataPoint] = "Buy"

		// Update the quote and base balances after the trade.
		ts.QuoteBalance -= (ts.PositionSize * ts.CurrentPrice) + transactionCost + slippageCost
		ts.BaseBalance += ts.PositionSize

		// Mark that we are in a trade.
		ts.InTrade = true

		// Record entry price for calculating profit/loss later.
		ts.EntryQuantity = ts.PositionSize

		return nil

	case "Sell":
		if !ts.InTrade {
			return fmt.Errorf("cannot execute a sell order without an existing trade")
		}

		// Calculate profit/loss for the trade.
		exitPrice := ts.CurrentPrice
		tradeProfitLoss := CalculateProfitLoss(ts.EntryPrice, exitPrice, ts.EntryQuantity)
		transactionCost := ts.TransactionCost * exitPrice * ts.EntryQuantity
		slippageCost := ts.Slippage * exitPrice * ts.EntryQuantity

		if (ts.BaseBalance < ts.EntryQuantity) || (tradeProfitLoss < transactionCost+slippageCost+md.TargetProfit) {
			if ts.BaseBalance < ts.EntryQuantity {
				return fmt.Errorf("cannot execute a sell order insufficient BaseBalance: %.8f needed up to: %.8f", ts.BaseBalance, ts.EntryQuantity)
			}
			return fmt.Errorf("cannot execute a sell order without profit up to: %.8f, this trade profit: %.8f", md.TargetProfit, tradeProfitLoss)
		}

		tradeProfitLoss -= transactionCost + slippageCost

		// Mark that we are no longer in a trade.
		ts.InTrade = false

		// Store profit/loss for the trade.
		md.TotalProfitLoss += tradeProfitLoss

		// Update the quote and base balances after the trade.
		ts.QuoteBalance += (ts.EntryQuantity * exitPrice) - transactionCost - slippageCost
		ts.BaseBalance -= ts.EntryQuantity

		// Record Signal for plotting graph later
		ts.Signals[ts.DataPoint] = "Sell"

		//reduce riskPosition by a factor
		// md.RiskPositionPercentage /= ts.RiskFactor
		// ts.RiskStopLossPercentage /= ts.RiskFactor

		return nil

	default:
		return fmt.Errorf("invalid trade action: %s", tradeAction)
	}
}

// RiskManagement applies risk management rules to limit potential losses.
// It calculates the stop-loss price based on the fixed percentage of risk per trade and the position size.
// If the current price breaches the stop-loss level, it triggers a sell signal and exits the trade.
func (ts *TradingSystem) RiskManagement(md *model.AppData) bool {

	// Calculate position size based on the fixed percentage of risk per trade.
	ts.RiskCost = ts.QuoteBalance * md.RiskPositionPercentage
	ts.PositionSize = ts.RiskCost / ts.CurrentPrice

	// Calculate profit/loss for the trade.
	exitPrice := ts.CurrentPrice
	tradeProfitLoss := CalculateProfitLoss(ts.EntryPrice, exitPrice, ts.EntryQuantity)
	transactionCost := ts.TransactionCost * exitPrice * ts.EntryQuantity
	slippageCost := ts.Slippage * exitPrice * ts.EntryQuantity
	tradeProfitLoss -= transactionCost + slippageCost

	// Check if the current price breaches the stop-loss level and triggers a sell signal.
	if ts.InTrade && tradeProfitLoss <= -md.TargetStopLoss && (ts.BaseBalance >= ts.EntryQuantity) {
		md.TotalProfitLoss += tradeProfitLoss

		// Update the quote and base balances after the trade.
		ts.QuoteBalance += (ts.EntryQuantity * exitPrice) - transactionCost - slippageCost
		ts.BaseBalance -= ts.EntryQuantity

		// Record Signal for plotting graph later.
		ts.Signals[ts.DataPoint] = "Sell"

		//increase riskPosition by a factor
		// md.RiskPositionPercentage *= ts.RiskFactor

		fmt.Printf("- SELL-StopLoss at %v Quant: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f LossCal: %.8f PosPcent: %.8f DataPt: %d\n",
			ts.CurrentPrice, ts.EntryQuantity, ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, tradeProfitLoss, md.RiskPositionPercentage, ts.DataPoint)

		// Mark that we are no longer in a trade.
		ts.InTrade = false
		// ts.StopLossRecover = ts.CurrentPrice * (1.0 - ts.RiskStopLossPercentage)
		return true
	}
	return false
}

// FundamentalAnalysis performs fundamental analysis and generates trading signals.
func (ts *TradingSystem) FundamentalAnalysis() (buySignal, sellSignal bool) {
	// Implement your fundamental analysis logic here.
	// Analyze project fundamentals, team credentials, market news, etc.
	// Determine the buy and sell signals based on the analysis.
	// Return true for buySignal and sellSignal if conditions are met.
	return false, false
}

// EntryRule defines the entry conditions for a trade.
func (ts *TradingSystem) EntryRule(md *model.AppData) bool {
	// Combine the results of technical and fundamental analysis to decide entry conditions.
	technicalBuy, _ := ts.TechnicalAnalysis(md)
	// fundamentalBuy, fundamentalSell := ts.FundamentalAnalysis()
	return technicalBuy
}

// ExitRule defines the exit conditions for a trade.
func (ts *TradingSystem) ExitRule(md *model.AppData) bool {
	// Combine the results of technical and fundamental analysis to decide exit conditions.
	_, technicalSell := ts.TechnicalAnalysis(md)
	// fundamentalBuy, fundamentalSell := ts.FundamentalAnalysis()
	return technicalSell
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) UpdateHistoricalData(loadFrom string) error {
	var (
		err         error
		exch        model.APIServices
		configParam *config.ExchConfig
		ok          bool
	)
	switch loadFrom {
	case "HitBTC":
		// Check if the key "HitBTC" exists in the map
		if configParam, ok = config.NewExchangesConfig()[loadFrom]; ok {
			// Check if REQUIRED "InfluxDB" exists in the map
			if influxDBParam, ok := config.NewExchangesConfig()["InfluxDB"]; ok {
				//Initiallize InfluDB config Parameters and instanciate its struct
				infDB, err := influxdb.NewAPIServices(influxDBParam)
				if err != nil {
					log.Fatalf("Error getting Required services from InfluxDB: %v", err)
				}				
				exch, err = hitbtcapi.NewAPIServices(infDB, configParam)
				if err != nil {
					log.Fatalf("Error getting new exchange services from HitBTC: %v", err)
				}
			} else {
				fmt.Println("Error Internal: Exch Required InfluxDB")
				return errors.New("Required InfluxDB services not contracted")
			}
		} else {
			fmt.Println("Error Internal: Exch name not HitBTC")
			return errors.New("Exch name not HitBTC")
		}
	case "binance":
	case "InfluxDB":
		// Check if the key "InfluxDB" exists in the map
		if configParam, ok = config.NewExchangesConfig()[loadFrom]; ok {
			exch, err = influxdb.NewAPIServices(configParam)
			if err != nil {
				log.Fatalf("Error getting new exchange services from InfluxDB: %v", err)
			}
		} else {
			fmt.Println("Error Internal: Exch name not InfluxDB")
			return errors.New("Exch name not InfluxDB")
		}
	case "csv":
	default:
		return errors.Errorf("Error updating historical data from %s invalid loadFrom tag \"%s\" ", loadFrom, loadFrom)
	}
	ts.HistoricalData, err = exch.FetchCandles(configParam.Symbol, configParam.CandleInterval, configParam.CandleStartTime, configParam.CandleEndTime)
	ts.APIServices = exch
	if err != nil {
		return err
	}
	// Extract the closing prices from the candles data
	for _, candle := range ts.HistoricalData {
		ts.ClosingPrices = append(ts.ClosingPrices, candle.Close)
	}
	return nil
}

// CalculateMACD calculates the Moving Average Convergence Divergence (MACD) and MACD Histogram for the given data and periods.
func (ts *TradingSystem) CalculateMACD(md *model.AppData, LongPeriod, ShortPeriod int) (macdLine, signalLine, macdHistogram []float64) {
	longEMA, shortEMA, _, err := ts.CandleExponentialMovingAverage(md, LongPeriod, ShortPeriod)
	if err != nil {
		log.Fatalf("Error: in CalclateMACD while tring to get EMA")
	}
	// Calculate MACD line
	macdLine = make([]float64, len(ts.ClosingPrices))
	for i := range ts.ClosingPrices {
		macdLine[i] = shortEMA[i] - longEMA[i]
	}

	// Calculate signal line using the MACD line
	signalLine = CalculateExponentialMovingAverage(macdLine, md.SignalMACDPeriod)

	// Calculate MACD Histogram
	macdHistogram = make([]float64, len(ts.ClosingPrices))
	for i := range ts.ClosingPrices {
		macdHistogram[i] = macdLine[i] - signalLine[i]
	}

	return macdLine, signalLine, macdHistogram
}

//CandleExponentialMovingAverage calculates EMA from condles
func (ts *TradingSystem) CandleExponentialMovingAverage(md *model.AppData, LongPeriod, ShortPeriod int) (longEMA, shortEMA []float64, timestamps []int64, err error) {
	var ema55, ema15 *ema.Ema
	if LongPeriod == 0 || ShortPeriod == 0 {
		ema55 = ema.NewEma(alphaFromN(md.LongPeriod))
		ema15 = ema.NewEma(alphaFromN(md.ShortPeriod))
	} else {
		ema55 = ema.NewEma(alphaFromN(LongPeriod))
		ema15 = ema.NewEma(alphaFromN(ShortPeriod))
	}

	longEMA = make([]float64, len(ts.HistoricalData))
	shortEMA = make([]float64, len(ts.HistoricalData))
	timestamps = make([]int64, len(ts.HistoricalData))
	for k, candle := range ts.HistoricalData {
		ema55.Step(candle.Close)
		ema15.Step(candle.Close)
		longEMA[k] = ema55.Compute()
		shortEMA[k] = ema15.Compute()
		timestamps[k] = candle.Timestamp
	}
	return longEMA, shortEMA, timestamps, nil
}

// CalculateExponentialMovingAverage calculates the Exponential Moving Average (EMA) for the given data and period.
func CalculateExponentialMovingAverage(data []float64, period int) (ema []float64) {
	ema = make([]float64, len(data))
	smoothingFactor := 2.0 / (float64(period) + 1.0)

	// Calculate the initial SMA as the sum of the first 'period' data points divided by 'period'.
	sma := 0.0
	for i := 0; i < period; i++ {
		sma += data[i]
	}
	ema[period-1] = sma / float64(period)

	// Calculate the EMA for the remaining data points.
	for i := period; i < len(data); i++ {
		ema[i] = (data[i]-ema[i-1])*smoothingFactor + ema[i-1]
	}

	return ema
}

// CalculateProfitLoss calculates the profit or loss from a trade.
func CalculateProfitLoss(entryPrice, exitPrice, positionSize float64) float64 {
	return (exitPrice - entryPrice) * positionSize
}

func CalculateSimpleMovingAverage(data []float64, period int) (sma []float64) {
	if period <= 0 || len(data) < period {
		return nil
	}

	sma = make([]float64, len(data)-period+1)
	for i := 0; i <= len(data)-period; i++ {
		sum := 0.0
		for j := i; j < i+period; j++ {
			sum += data[j]
		}
		sma[i] = sum / float64(period)
	}

	return sma
}

func CalculateStandardDeviation(data []float64, period int) []float64 {
	if period <= 0 || len(data) < period {
		return nil
	}

	stdDev := make([]float64, len(data)-period+1)
	for i := 0; i <= len(data)-period; i++ {
		avg := CalculateSimpleMovingAverage(data[i:i+period], period)[0]
		variance := 0.0
		for j := i; j < i+period; j++ {
			variance += math.Pow(data[j]-avg, 2)
		}
		stdDev[i] = math.Sqrt(variance / float64(period))
	}

	return stdDev
}

// StochasticRSI calculates the Stochastic RSI for a given RSI data series, period, and SmoothK, SmoothD.
func StochasticRSI(rsi []float64, period, smoothK, smoothD int) ([]float64, []float64) {
	stochasticRSI := make([]float64, len(rsi)-period+1)
	// smoothKRSI := make([]float64, len(stochasticRSI)-smoothK+1)

	for i := period; i < len(rsi); i++ {
		highestRSI := rsi[i-period]
		lowestRSI := rsi[i-period]

		// Find the highest and lowest RSI values over the specified period
		for j := i - period + 1; j <= i; j++ {
			if rsi[j] > highestRSI {
				highestRSI = rsi[j]
			}
			if rsi[j] < lowestRSI {
				lowestRSI = rsi[j]
			}
		}

		// Calculate the Stochastic RSI value
		stochasticRSI[i-period] = (rsi[i] - lowestRSI) / (highestRSI - lowestRSI)
	}

	// Smooth Stochastic RSI using Exponential Moving Averages
	// Smooth Stochastic RSI using Exponential Moving Averages
	emaSmoothK := CalculateExponentialMovingAverage(stochasticRSI, smoothK)
	emaSmoothD := CalculateExponentialMovingAverage(emaSmoothK, smoothD)

	return emaSmoothK, emaSmoothD
}

// CalculateRSI calculates the Relative Strength Index for a given data series and period.
func CalculateRSI(data []float64, period int) []float64 {
	rsi := make([]float64, len(data)-period+1)

	// Calculate initial average gains and losses
	var gainSum, lossSum float64
	for i := 1; i <= period; i++ {
		change := data[i] - data[i-1]
		if change >= 0 {
			gainSum += change
		} else {
			lossSum += -change
		}
	}

	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)

	rsi[0] = 100.0 - (100.0 / (1.0 + avgGain/avgLoss))

	// Calculate RSI for the remaining data points
	for i := period + 1; i < len(data); i++ {
		change := data[i] - data[i-1]

		var gain, loss float64
		if change >= 0 {
			gain = change
		} else {
			loss = -change
		}

		avgGain = (avgGain*(float64(period)-1) + gain) / float64(period)
		avgLoss = (avgLoss*(float64(period)-1) + loss) / float64(period)

		rsi[i-period] = 100.0 - (100.0 / (1.0 + avgGain/avgLoss))
	}

	return rsi
}

// CalculateBollingerBands calculates the middle (SMA) and upper/lower Bollinger Bands for the given data and period.
func CalculateBollingerBands(data []float64, period int, numStdDev float64) ([]float64, []float64, []float64) {
	sma := CalculateSimpleMovingAverage(data, period)
	stdDev := CalculateStandardDeviation(data, period)

	upperBand := make([]float64, len(sma))
	lowerBand := make([]float64, len(sma))

	for i := range sma {
		upperBand[i] = sma[i] + numStdDev*stdDev[i]
		lowerBand[i] = sma[i] - numStdDev*stdDev[i]
	}

	return sma, upperBand, lowerBand
}

func alphaFromN(N int) float64 {
	var n = float64(N)
	return 2. / (n + 1.)
}
