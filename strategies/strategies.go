package strategies

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"

	"coinbitly.com/config"
	"coinbitly.com/binanceapi"
	"coinbitly.com/model"
	"github.com/apourchet/investment/lib/ema"
	"github.com/pkg/errors"
)

// TradingSystem struct: The TradingSystem struct represents the main trading
// system and holds various parameters and fields related to the strategy,
// trading state, and performance.
type TradingSystem struct {
	Strategy           string
	StrategyCombLogic  string
	HistoricalData     []model.Candlestick
	ClosingPrices      []float64
	Timestamps         []int64
	Signals            []string
	ShortPeriod        int
	LongPeriod         int
	LongMACDPeriod     int
	ShortMACDPeriod	   int
	RsiPeriod          int
	StochRSIPeriod     int
	SmoothK            int
	SmoothD            int
	Overbought 		   float64 // Define overbought for generating RSI signals
	Oversold           float64 //  Define oversold thresholds for generating RSI signals
	MacdSignalPeriod   int
	BollingerPeriod    int
	BollingerNumStdDev float64
	TransactionCost    float64
	Slippage           float64
	InitialCapital         float64
	PositionSize           float64
	RiskStopLossPercentage float64
	RiskPositionPercentage float64
	TotalProfitLoss        float64
	EntryPrice             float64
	InTrade                bool
	QuoteBalance           float64
	BaseBalance            float64
	StopLossPrice          float64
	RiskAmount             float64
	DataPoint              int
	CurrentPrice           float64
	EntryQuantity          float64
	StopLossRecover        float64
	RiskFactor 			   float64
}

type backT struct{
	count int
	Strategy string
	StrategyCombLogic string
	ShortPeriod  int     
	LongPeriod  int     
	ShortMACDPeriod  int     
	LongMACDPeriod  int 
	RiskStopLossPercentage float64
	RiskPositionPercentage float64
	RsiPeriod  int
	StochRSIPeriod     int
	SmoothK            int
	SmoothD            int
	MacdSignalPeriod  int 
	BollingerPeriod  int    
	BollingerNumStdDev float64
	Overbought 		   float64 // Define overbought for generating RSI signals
	Oversold           float64 
}

// NewTradingSystem(): This function initializes the TradingSystem and fetches
// historical data from the exchange using the GetCandlesFromExch() function.
// It sets various strategy parameters like short and long EMA periods, RSI period,
// MACD signal period, Bollinger Bands period, and more.
func NewTradingSystem() (*TradingSystem, error) {
	// Initialize the trading system.
	ts := &TradingSystem{
		ShortPeriod: 6, // Define moving average short period for the strategy.
		LongPeriod:  16, // Define moving average long period for the strategy.
		ShortMACDPeriod: 12,
		LongMACDPeriod: 29,	   
	}
	ts.RiskStopLossPercentage = 0.01    // Define risk management parameter 25% stop-loss
	ts.RiskPositionPercentage = 0.10     // Define risk management parameter 5% balance
	ts.RsiPeriod = 14                    // Define RSI period parameter. 12,296,16 
	ts.StochRSIPeriod = 3
	ts.SmoothK = 3
	ts.SmoothD = 3	
	ts.Overbought = 0.8 // Define overbought for generating RSI signals
	ts.Oversold = 0.2 // Define oversold for generating RSI signals
	ts.MacdSignalPeriod = 9              // Define MACD period parameter.
	ts.BollingerPeriod = 20              // Define Bollinger Bands parameter.
	ts.BollingerNumStdDev = 2.0          // Define Bollinger Bands parameter.
	ts.Strategy = "MACD"

	ts.RiskFactor = 2.0           // Define 1% slippage
	ts.TransactionCost = 0.001           // Define 0.1% transaction cost
	ts.Slippage = 0.01        
	ts.InitialCapital = 1000.0           //Initial Capital for simulation on backtesting
	ts.QuoteBalance = ts.InitialCapital  //continer to hold the balance
	ts.StopLossRecover = math.MaxFloat64 //

	// Fetch historical data from the exchange
	err := ts.UpdateHistoricalData()
	if err != nil {
		return &TradingSystem{}, err
	}
	// Perform any other necessary initialization steps here
	ts.Signals = make([]string, len(ts.ClosingPrices))   // Holder of generated trading signals
	ts.Timestamps = make([]int64, len(ts.ClosingPrices)) // Holder of generated trading signals

	return ts, nil
}

func (ts *TradingSystem) ReconfigureTS(v backT) {
	// Reconfigure the trading system.	
	ts.ShortPeriod = v.ShortPeriod
	ts.LongPeriod = v.LongPeriod
	ts.ShortMACDPeriod = v.ShortMACDPeriod
	ts.LongMACDPeriod = v.LongMACDPeriod
	ts.RiskStopLossPercentage = v.RiskStopLossPercentage
	ts.RiskPositionPercentage = v.RiskPositionPercentage
	ts.RsiPeriod = v.RsiPeriod
	ts.MacdSignalPeriod = v.MacdSignalPeriod
	ts.BollingerPeriod = v.BollingerPeriod
	ts.BollingerNumStdDev = v.BollingerNumStdDev
    ts.StochRSIPeriod = v.StochRSIPeriod
    ts.SmoothK = v.SmoothK
    ts.SmoothD    = v.SmoothD
    ts.Overbought = v.Overbought
    ts.Oversold  = v.Oversold
	ts.Strategy = v.Strategy
	ts.StrategyCombLogic = v.StrategyCombLogic
	ts.RiskFactor = 2.0           // Define 1% slippage
	ts.TransactionCost = 0.001           // Define 0.1% transaction cost
	ts.Slippage = 0.01        
	ts.InitialCapital = 1000.0           //Initial Capital for simulation on backtesting
	ts.QuoteBalance = ts.InitialCapital  //continer to hold the balance
	ts.StopLossRecover = math.MaxFloat64 //

	// Extract the closing prices from the candles data
	for _, candle := range ts.HistoricalData {
		ts.ClosingPrices = append(ts.ClosingPrices, candle.Close)
	}
	// Perform any other necessary initialization steps here
	ts.Signals = make([]string, len(ts.ClosingPrices))   // Holder of generated trading signals
	ts.Timestamps = make([]int64, len(ts.ClosingPrices)) // Holder of generated trading signals

	return
}

// Backtest(): This function simulates the backtesting process using historical
// price data. It iterates through the closing prices, checks for entry and exit
// conditions, and executes trades accordingly. It also tracks trading performance
// and updates the current balance after each trade.
func (ts *TradingSystem) Backtest() {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)
//count, ShortPeriod,LongPeriod,ShortMACDPeriod,LongMACDPeriod,RiskStopLossPercentage
//RiskPositionPercentage,RsiPeriod,StochRSIPeriod,SmoothK,SmoothD,MacdSignalPeriod,BollingerPeriod,BollingerNumStdDev     
//Strategy,Overbought,Oversold
	testkit := []backT{  //StochRSI
		// {1,6,16,12,29,0.005,0.10,14,9,20,2.0,"EMA",},
		{2,"MACD","",6,16,12,29,0.15,0.05,14,14,3,3,9,20,2.0,0.7,0.2},
		// {3,6,16,12,29,0.05,0.05,14,9,20,2.0,"MACD",},
		// {4,6,16,12,29,0.1,0.05,14,9,20,2.0,"Bollinger",},
		// {5,6,16,12,29,0.15,0.05,14,9,20,2.0,"EMA",},
		// {6,6,16,12,29,0.5,0.05,14,9,20,2.0,"EMA",},
	}
	for _, v := range testkit{
		ts = &TradingSystem{HistoricalData: ts.HistoricalData}
		ts.ReconfigureTS(v)
		// Initialize variables for tracking trading performance.
		tradeCount := 0
		ts.TotalProfitLoss = 0.0
		// Simulate the backtesting process using historical price data.
		for ts.DataPoint, ts.CurrentPrice = range ts.ClosingPrices {
			// Execute the trade if entry conditions are met.
			if (!ts.InTrade) && (ts.EntryRule() && (ts.CurrentPrice <= ts.StopLossRecover)) {
				// Record entry price for calculating profit/loss and stoploss later.
				ts.EntryPrice = ts.CurrentPrice

				// Execute the buy order using the ExecuteStrategy function.
				err := ts.ExecuteStrategy("Buy")
				if err != nil {
					fmt.Println("Error executing buy order:", err)
					ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position
					continue
				}

				// Mark that we are in a trade.
				ts.InTrade = true
				tradeCount++
				ts.StopLossRecover = math.MaxFloat64
				fmt.Printf("- BUY at %v Quant: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f LossPcent: %.8f PosPcent: %.8f DataPt: %d\n", ts.CurrentPrice, ts.EntryQuantity, ts.QuoteBalance, ts.BaseBalance, ts.TotalProfitLoss, ts.RiskStopLossPercentage, ts.RiskPositionPercentage, ts.DataPoint)

				// Close the trade if exit conditions are met.
			} else if ts.InTrade && ts.ExitRule() {
				// Execute the sell order using the ExecuteStrategy function.
				err := ts.ExecuteStrategy("Sell")
				if err != nil {
					if fmt.Sprintf("%v", err) == "Stoploss Triggered" {
						ts.InTrade = false
						tradeCount++
						fmt.Println(err, ": StoplossRecover", ts.StopLossRecover)
						continue
					}
					// fmt.Println("Error:", err, " at:", ts.CurrentPrice, ", Stoploss below::", ts.StopLossPrice)
					ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position
					continue
				}

				fmt.Printf("- SELL at %v Quant: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f LossPcent: %.8f PosPcent: %.8f DataPt: %d\n", ts.CurrentPrice, ts.EntryQuantity, ts.QuoteBalance, ts.BaseBalance, ts.TotalProfitLoss, ts.RiskStopLossPercentage, ts.RiskPositionPercentage, ts.DataPoint)

				// Mark that we are no longer in a trade.
				ts.InTrade = false
				tradeCount++
				// Implement risk management and update capital after each trade
				// For simplicity, we'll skip this step in this basic example

			} else {
				ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position
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
			ts.TotalProfitLoss += tradeProfitLoss

			ts.QuoteBalance += (ts.BaseBalance * exitPrice) - transactionCost - slippageCost
			ts.BaseBalance -= ts.BaseBalance
			ts.Timestamps = append(ts.Timestamps, ts.Timestamps[len(ts.Timestamps)-1])
			ts.ClosingPrices = append(ts.ClosingPrices, ts.ClosingPrices[len(ts.ClosingPrices)-1])
			ts.Signals = append(ts.Signals, "Sell")
			fmt.Printf("- SELL-Off at %v Quant: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f LossPcent: %.8f PosPcent: %.8f DataPt: %d\n", ts.CurrentPrice, ts.BaseBalance, ts.QuoteBalance, ts.BaseBalance, ts.TotalProfitLoss, ts.RiskStopLossPercentage, ts.RiskPositionPercentage, ts.DataPoint)
		}
		// Print the overall trading performance after backtesting.
		fmt.Printf("\nBacktesting Summary: %d\n",v.count)
		fmt.Printf("Total Trades: %d, ", tradeCount)
		fmt.Printf("Total Profit/Loss: %.2f, ", ts.TotalProfitLoss)
		fmt.Printf("Final Capital: %.2f, ", ts.QuoteBalance)
		fmt.Printf("Final Asset: %.8f ", ts.BaseBalance)
		err := CreateLineChartWithSignals(ts.Timestamps, ts.ClosingPrices, ts.Signals)
		if err != nil {
			fmt.Println("Error creating Line Chart with signals:", err)
			return
		}
		<-sigchnl
	}	
}

// ExecuteStrategy executes the trade based on the provided trade action and current price.
// The tradeAction parameter should be either "Buy" or "Sell".
func (ts *TradingSystem) ExecuteStrategy(tradeAction string) error {
	// Calculate position size based on the fixed percentage of risk per trade.
	isStopLossExecuted := ts.RiskManagement() // make sure entry price is set before calling risk management
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

		// Mark that we are in a trade.
		ts.InTrade = true

		// Record Signal for plotting graph later
		ts.Signals[ts.DataPoint] = "Buy"

		// Update the quote and base balances after the trade.
		ts.QuoteBalance -= (ts.PositionSize * ts.CurrentPrice) + transactionCost + slippageCost
		ts.BaseBalance += ts.PositionSize

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

		if tradeProfitLoss < (transactionCost+slippageCost)*(1+0.5) {
			return fmt.Errorf("cannot execute a sell order without profit")
		} else {
			tradeProfitLoss -= transactionCost + slippageCost
		}

		// Mark that we are no longer in a trade.
		ts.InTrade = false

		// Store profit/loss for the trade.
		ts.TotalProfitLoss += tradeProfitLoss

		// Update the quote and base balances after the trade.
		ts.QuoteBalance += (ts.EntryQuantity * exitPrice) - transactionCost - slippageCost
		ts.BaseBalance -= ts.EntryQuantity

		// Record Signal for plotting graph later
		ts.Signals[ts.DataPoint] = "Sell"

		
		//reduce riskPosition by a factor
		ts.RiskPositionPercentage /= ts.RiskFactor 
		// ts.RiskStopLossPercentage /= ts.RiskFactor 

		return nil

	default:
		return fmt.Errorf("invalid trade action: %s", tradeAction)
	}
}

// RiskManagement applies risk management rules to limit potential losses.
// It calculates the stop-loss price based on the fixed percentage of risk per trade and the position size.
// If the current price breaches the stop-loss level, it triggers a sell signal and exits the trade.
func (ts *TradingSystem) RiskManagement() bool {
	// Calculate stop-loss price based on the fixed percentage of risk per trade.
	ts.StopLossPrice = ts.EntryPrice * (1.0 - ts.RiskStopLossPercentage)

	// Calculate position size based on the fixed percentage of risk per trade.
	ts.RiskAmount = ts.QuoteBalance * ts.RiskPositionPercentage
	ts.PositionSize = ts.RiskAmount / ts.StopLossPrice

	// Check if the current price breaches the stop-loss level and triggers a sell signal.
	if ts.InTrade && ts.CurrentPrice <= ts.StopLossPrice {
		// Calculate profit/loss for the trade.
		exitPrice := ts.CurrentPrice
		tradeProfitLoss := CalculateProfitLoss(ts.EntryPrice, exitPrice, ts.EntryQuantity)
		transactionCost := ts.TransactionCost * exitPrice * ts.EntryQuantity
		slippageCost := ts.Slippage * exitPrice * ts.EntryQuantity
		tradeProfitLoss -= transactionCost + slippageCost
		ts.TotalProfitLoss += tradeProfitLoss
		
		// Update the quote and base balances after the trade.
		ts.QuoteBalance += (ts.EntryQuantity * exitPrice) - transactionCost - slippageCost
		ts.BaseBalance -= ts.EntryQuantity

		// Record Signal for plotting graph later.
		ts.Signals[ts.DataPoint] = "Sell"

		//increase riskPosition by a factor
		ts.RiskPositionPercentage *= ts.RiskFactor
		ts.RiskStopLossPercentage *= ts.RiskFactor 

		fmt.Printf("- SELL-StopLoss at %v Quant: %.8f, Price: %.8f, QBal: %.8f, BBal: %.8f, totalP&L %.2f LossPcent: %.8f PosPcent: %.8f DataPt: %d\n",
			ts.CurrentPrice, ts.EntryQuantity, ts.RiskAmount, ts.QuoteBalance, ts.BaseBalance, ts.TotalProfitLoss, ts.RiskStopLossPercentage, ts.RiskPositionPercentage, ts.DataPoint)

		// Mark that we are no longer in a trade.
		ts.InTrade = false
		ts.StopLossRecover = ts.CurrentPrice * (1.0 - ts.RiskStopLossPercentage)
		return true
	}
	return false
}

// TechnicalAnalysis(): This function performs technical analysis using the
// calculated moving averages (short and long EMA), RSI, MACD line, and Bollinger
// Bands. It determines the buy and sell signals based on various strategy rules.
func (ts *TradingSystem) TechnicalAnalysis() (buySignal, sellSignal bool) {
	// Calculate moving averages (MA) using historical data.
	longEMA, shortEMA, timeStamps, err := ts.CandleExponentialMovingAverage(0, 0)
	if err != nil {
		log.Fatalf("Error: in TechnicalAnalysis Unable to get EMA: %v", err)
	}
	// Calculate Relative Strength Index (RSI) using historical data.
	rsi := CalculateRSI(ts.ClosingPrices, ts.RsiPeriod)

	// Calculate Stochastic RSI and SmoothK, SmoothD values
	stochRSI, smoothKRSI := StochasticRSI(rsi, ts.StochRSIPeriod, ts.SmoothK, ts.SmoothD)

	// Calculate MACD and Signal Line using historical data.
	macdLine, signalLine, macdHistogram := ts.CalculateMACD(ts.LongMACDPeriod, ts.ShortMACDPeriod)

	// Calculate Bollinger Bands using historical data.
	_, upperBand, lowerBand := CalculateBollingerBands(ts.ClosingPrices, ts.BollingerPeriod, ts.BollingerNumStdDev)

	// Determine the buy and sell signals based on the moving averages, RSI, MACD line, and Bollinger Bands.
	if len(stochRSI) > 1 && len(smoothKRSI) > 1 && len(shortEMA) > 1 && len(longEMA) > 1 && len(rsi) > 0 && len(macdLine) > 1 && len(signalLine) > 1 && len(upperBand) > 0 && len(lowerBand) > 0 {
		// Generate buy/sell signals based on conditions
		count := 0
		switch ts.StrategyCombLogic{
		case "AND":
			buySignal, sellSignal = true, true
			if  strings.Contains(ts.Strategy, "StochR") && (ts.DataPoint < len(stochRSI)) && ts.DataPoint >= (ts.StochRSIPeriod + ts.SmoothK + ts.SmoothD - 1){
				count++                
				buySignal = buySignal && ((stochRSI[ts.DataPoint-1] < smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] >= smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint]<ts.Oversold)) 
				sellSignal = sellSignal && ((stochRSI[ts.DataPoint-1] >= smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] < smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint]<ts.Oversold))
				// buySignal = buySignal && (stochRSI[ts.DataPoint] > ts.Overbought && smoothKRSI[ts.DataPoint-ts.StochRSIPeriod-ts.SmoothK] > ts.Overbought)
				// sellSignal = sellSignal && (stochRSI[ts.DataPoint] < ts.Oversold && smoothKRSI[ts.DataPoint-ts.StochRSIPeriod-ts.SmoothK] < ts.Oversold)
			}			
			if strings.Contains(ts.Strategy, "EMA") && ts.DataPoint > 0 {
				count++
				buySignal = buySignal && (shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint])
				sellSignal = sellSignal && (shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint])
			}
			if  strings.Contains(ts.Strategy, "RSI") && ts.DataPoint > 0  && ts.DataPoint <= len(rsi) {
				count++
				buySignal = buySignal && (rsi[ts.DataPoint-1] < 30)
				sellSignal = sellSignal && (rsi[ts.DataPoint-1] > 70)
			}
			if strings.Contains(ts.Strategy, "MACD") && ts.DataPoint > 0{
				count++
				buySignal = buySignal && (macdLine[ts.DataPoint-1] <= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] > signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] > 0)
				sellSignal = sellSignal && (macdLine[ts.DataPoint-1] >= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] < signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] < 0)
			}
			if strings.Contains(ts.Strategy, "Bollinger") && ts.DataPoint > 0  && ts.DataPoint < len(upperBand){
				count++
				buySignal = buySignal && ts.ClosingPrices[len(ts.ClosingPrices)-1] < lowerBand[ts.DataPoint]
				sellSignal = sellSignal && ts.ClosingPrices[len(ts.ClosingPrices)-1] > upperBand[ts.DataPoint]
			}
		case "OR":
			if  strings.Contains(ts.Strategy, "StochR") && (ts.DataPoint < len(stochRSI)) && ts.DataPoint >= (ts.StochRSIPeriod + ts.SmoothK + ts.SmoothD - 1){
				count++                
				buySignal = buySignal || ((stochRSI[ts.DataPoint-1] < smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] >= smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint]<ts.Oversold)) 
				sellSignal = sellSignal || ((stochRSI[ts.DataPoint-1] >= smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] < smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint]<ts.Oversold))
				// buySignal = buySignal || (stochRSI[ts.DataPoint] > ts.Overbought && smoothKRSI[ts.DataPoint-ts.StochRSIPeriod-ts.SmoothK] > ts.Overbought)
				// sellSignal = sellSignal || (stochRSI[ts.DataPoint] < ts.Oversold && smoothKRSI[ts.DataPoint-ts.StochRSIPeriod-ts.SmoothK] < ts.Oversold)
			}			
			if strings.Contains(ts.Strategy, "EMA") && ts.DataPoint > 0 {
				count++
				buySignal = buySignal || (shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint])
				sellSignal = sellSignal || (shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint])
			}
			if  strings.Contains(ts.Strategy, "RSI") && ts.DataPoint > 0  && ts.DataPoint <= len(rsi) {
				count++
				buySignal = buySignal || (rsi[ts.DataPoint-1] < 30)
				sellSignal = sellSignal || (rsi[ts.DataPoint-1] > 70)
			}
			if strings.Contains(ts.Strategy, "MACD") && ts.DataPoint > 0{
				count++
				buySignal = buySignal || (macdLine[ts.DataPoint-1] <= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] > signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] > 0)
				sellSignal = sellSignal || (macdLine[ts.DataPoint-1] >= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] < signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] < 0)
			}
			if strings.Contains(ts.Strategy, "Bollinger") && ts.DataPoint > 0  && ts.DataPoint < len(upperBand){
				count++
				buySignal = buySignal || (ts.ClosingPrices[len(ts.ClosingPrices)-1] < lowerBand[ts.DataPoint])
				sellSignal = sellSignal || (ts.ClosingPrices[len(ts.ClosingPrices)-1] > upperBand[ts.DataPoint])
			}
		default:
			if  strings.Contains(ts.Strategy, "StochR") && (ts.DataPoint < len(stochRSI)) && ts.DataPoint >= (ts.StochRSIPeriod + ts.SmoothK + ts.SmoothD - 1){
				count++                
				buySignal = (stochRSI[ts.DataPoint-1] < smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] >= smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint]<ts.Oversold) 
				sellSignal = (stochRSI[ts.DataPoint-1] >= smoothKRSI[ts.DataPoint-1] && stochRSI[ts.DataPoint] < smoothKRSI[ts.DataPoint]) && (stochRSI[ts.DataPoint]<ts.Oversold)
				// buySignal = stochRSI[ts.DataPoint] > ts.Overbought && smoothKRSI[ts.DataPoint-ts.StochRSIPeriod-ts.SmoothK] > ts.Overbought)
				// sellSignal = stochRSI[ts.DataPoint] < ts.Oversold && smoothKRSI[ts.DataPoint-ts.StochRSIPeriod-ts.SmoothK] < ts.Oversold)
			}			
			if strings.Contains(ts.Strategy, "EMA") && ts.DataPoint > 0 {
				count++
				buySignal = shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint]
				sellSignal = shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint]
			}
			if  strings.Contains(ts.Strategy, "RSI") && ts.DataPoint > 0  && ts.DataPoint <= len(rsi) {
				count++
				buySignal = rsi[ts.DataPoint-1] < 30
				sellSignal = rsi[ts.DataPoint-1] > 70
			}
			if strings.Contains(ts.Strategy, "MACD") && ts.DataPoint > 0{
				count++
				buySignal = macdLine[ts.DataPoint-1] <= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] > signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] > 0
				sellSignal = macdLine[ts.DataPoint-1] >= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] < signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] < 0
			}
			if strings.Contains(ts.Strategy, "Bollinger") && ts.DataPoint > 0  && ts.DataPoint < len(upperBand){
				count++
				buySignal = buySignal || ts.ClosingPrices[len(ts.ClosingPrices)-1] < lowerBand[ts.DataPoint]
				sellSignal = sellSignal || ts.ClosingPrices[len(ts.ClosingPrices)-1] > upperBand[ts.DataPoint]
			} 
		}

		if count == 0 {
			buySignal, sellSignal = false, false
		}
		_ = timeStamps
		// PlotEMA(timeStamps, stochRSI, smoothKRSI)
		PlotRSI(stochRSI[290:315], smoothKRSI[290:315])
		
		// fmt.Printf("EMA: %v, RSI: %v, MACD: %v, Bollinger: %v buySignal: %v, sellSignal: %v", ts.Strategy.UseEMA, ts.Strategy.UseRSI, ts.Strategy.UseMACD, ts.Strategy.UseBollinger, buySignal, sellSignal)
	}

	return buySignal, sellSignal
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
func (ts *TradingSystem) EntryRule() bool {
	// Combine the results of technical and fundamental analysis to decide entry conditions.
	technicalBuy, _ := ts.TechnicalAnalysis()
	// fundamentalBuy, fundamentalSell := ts.FundamentalAnalysis()
	return technicalBuy
}

// ExitRule defines the exit conditions for a trade.
func (ts *TradingSystem) ExitRule() bool {
	// Combine the results of technical and fundamental analysis to decide exit conditions.
	_, technicalSell := ts.TechnicalAnalysis()
	// fundamentalBuy, fundamentalSell := ts.FundamentalAnalysis()
	return technicalSell
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) UpdateHistoricalData() error {
	var err error
	binanceExch, err := binanceapi.NewExchServices(config.NewExchangesConfig())
	if err != nil {
		log.Fatal(err)
	}
	ts.HistoricalData, err = binanceExch.FetchHistoricalCandlesticks(binanceExch.Symbol, binanceExch.CandleInterval, binanceExch.CandleStartTime, binanceExch.CandleEndTime)

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
func (ts *TradingSystem) CalculateMACD(LongPeriod, ShortPeriod int) (macdLine, signalLine, macdHistogram []float64) {
	longEMA, shortEMA, _, err := ts.CandleExponentialMovingAverage(LongPeriod, ShortPeriod)
	if err != nil {
		log.Fatalf("Error: in CalclateMACD while tring to get EMA")
	}
	// Calculate MACD line
	macdLine = make([]float64, len(ts.ClosingPrices))
	for i := range ts.ClosingPrices {
		macdLine[i] = shortEMA[i] - longEMA[i]
	}

	// Calculate signal line using the MACD line
	signalLine = CalculateExponentialMovingAverage(macdLine, ts.MacdSignalPeriod)

	// Calculate MACD Histogram
	macdHistogram = make([]float64, len(ts.ClosingPrices))
	for i := range ts.ClosingPrices {
		macdHistogram[i] = macdLine[i] - signalLine[i]
	}

	return macdLine, signalLine, macdHistogram
}

//CandleExponentialMovingAverage calculates EMA from condles
func (ts *TradingSystem) CandleExponentialMovingAverage(LongPeriod, ShortPeriod int) (longEMA, shortEMA []float64, timestamps []int64, err error) {
	var ema55, ema15 *ema.Ema
	if LongPeriod == 0 || ShortPeriod == 0{
		ema55 = ema.NewEma(alphaFromN(ts.LongPeriod))
		ema15 = ema.NewEma(alphaFromN(ts.ShortPeriod))
	}else{	
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
