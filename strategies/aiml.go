
package strategies

import (

)

func (ts *TradingSystem) EvaluateEMAParameters(data []float64, emaPeriods []int) (EMAProfit, MACDProfit, RSIProfit float64) {
	if len(emaPeriods) != 2 {
		panic("Expecting exactly two EMA periods for crossover strategy")
	}

	// Calculate trading performance for the given set of EMA periods
	longEMA, shortEMA, _, _ := ts.CandleExponentialMovingAverage(emaPeriods[1], emaPeriods[0])

	// Calculate trading performance for the given set of MACD EMA periods on a fixed signal
	MACDLine, SignalLine, MacdHistogram  := ts.CalculateMACD(emaPeriods[1], emaPeriods[0])

	// Calculate Relative Strength Index (RSI) using historical data.
	rsi := CalculateRSI(ts.ClosingPrices, ts.RsiPeriod)

	inPosition := false
	entryPrice := 0.0

	//EMA
	for j := 1; j < len(data); j++ {
		if shortEMA[j-1] < longEMA[j-1] && shortEMA[j] >= longEMA[j] && !inPosition {
			// Buy signal
			entryPrice = data[j]
			inPosition = true
		} else if shortEMA[j-1] > longEMA[j-1] && shortEMA[j] <= longEMA[j] && inPosition {
			// Sell signal
			EMAProfit += data[j] - entryPrice
			inPosition = false
		}
	}

	//MACD
	for j := 1; j < len(data); j++ {
		if MACDLine[j-1] <= SignalLine[j-1] && MACDLine[j] > SignalLine[j] && MacdHistogram[j] > 0 && !inPosition{
			// Buy signal
			entryPrice = data[j]
			inPosition = true
		} else if MACDLine[j-1] >= SignalLine[j-1] && MACDLine[j] < SignalLine[j] && MacdHistogram[j] < 0 && inPosition{
			// Sell signal
			MACDProfit += data[j] - entryPrice
			inPosition = false
		}
	}

	//RSI
	for j := 1; j < len(rsi); j++ {
		if (rsi[j-1] < 30) && !inPosition{
			// Buy signal
			entryPrice = data[j]
			inPosition = true
		} else if (rsi[j-1] > 70) && inPosition{
			// Sell signal
			RSIProfit += data[j] - entryPrice
			inPosition = false
		}
	}

	//Bollinger

	return EMAProfit, MACDProfit, RSIProfit
}