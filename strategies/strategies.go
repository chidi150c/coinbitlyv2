package strategies

import (
	"fmt"
	"log"
	"math"
	"time"

	"coinbitly.com/config"
	"coinbitly.com/hitbtcapi"
	"coinbitly.com/model"
	"github.com/apourchet/investment/lib/ema"
)

type techAnalysis struct{
	UseEMA bool
	UseRSI bool
	UseMACD bool
	UseBollinger bool
}

// TradingSystem represents the main trading system.
type TradingSystem struct {
	// Add any necessary fields and indicators here.
	Strategy techAnalysis
	HistoricalData []model.Candlestick
	ClosingPrices []float64
	Timestamps[]int64
	Signals []string
	ShortPeriod int
	LongPeriod int
	RsiPeriod int
	MacdSignalPeriod int
	BollingerPeriod int
	BollingerNumStdDev float64
	TransactionCost float64
	Slippage float64
	// Fields to track trade state and performance.
	InitialCapital float64
	PositionSize float64
	RiskPercentage float64
	TotalProfitLoss float64
	EntryPrice    float64
	InTrade       bool
	CurrentBalance float64
	StopLossPrice float64
	RiskAmount float64
	DataPoint int
	CurrentPrice float64
}

// TechnicalAnalysis performs technical analysis and generates trading signals.
func (ts *TradingSystem) TechnicalAnalysis() (buySignal, sellSignal bool) {
	// Calculate moving averages (MA) using historical data.
	shortEMA := CalculateExponentialMovingAverage(ts.ClosingPrices, ts.ShortPeriod)
	longEMA := CalculateExponentialMovingAverage(ts.ClosingPrices, ts.LongPeriod)

	// Calculate Relative Strength Index (RSI) using historical data.
	rsi := CalculateRSI(ts.ClosingPrices, ts.RsiPeriod)

	// Calculate MACD and Signal Line using historical data.
	macdLine, signalLine, macdHistogram := ts.CalculateMACD()

	// Calculate Bollinger Bands using historical data.
	_, upperBand, lowerBand := CalculateBollingerBands(ts.ClosingPrices, ts.BollingerPeriod, ts.BollingerNumStdDev)

	// Determine the buy and sell signals based on the moving averages, RSI, MACD line, and Bollinger Bands.
	if len(shortEMA) > 1 && len(longEMA) > 1 && len(rsi) > 0 && len(macdLine) > 1 && len(signalLine) > 1 && len(upperBand) > 0 && len(lowerBand) > 0 {
		// Get the latest moving average values, RSI, MACD line, and Bollinger Bands.
		previousshortEMA := shortEMA[len(shortEMA)-2]
		previouslongEMA := longEMA[len(longEMA)-2]
		currentshortEMA := shortEMA[len(shortEMA)-1]
		currentlongEMA := longEMA[len(longEMA)-1]
		currentRSI := rsi[len(rsi)-1]
		previousMACDLine := macdLine[len(macdLine)-2]
		currentMACDLine := macdLine[len(macdLine)-1]
		previousSignalLine := signalLine[len(signalLine)-2]
		currentSignalLine := signalLine[len(signalLine)-1]
		currentUpperBand := upperBand[len(upperBand)-1]
		currentLowerBand := lowerBand[len(lowerBand)-1]
		currentMacdHistogram := macdHistogram[len(macdHistogram)-1]
		
		// Buy Signal:
		// Short-term MA crosses above long-term MA,
		// RSI is below 30 (oversold),
		// Price is below the lower Bollinger Band, and
		// MACD crosses above the Signal Line.
		// buySignal = previousshortEMA <= previouslongEMA && currentshortEMA > currentlongEMA && currentRSI < 30 && ts.ClosingPrices[len(ts.ClosingPrices)-1] < currentLowerBand && previousMACDLine <= previousSignalLine && currentMACDLine > currentSignalLine
				
		// Sell Signal:
		// Short-term MA crosses below long-term MA,
		// RSI is above 70 (overbought),
		// Price is above the upper Bollinger Band, and
		// MACD crosses below the Signal Line.
		if ts.Strategy.UseEMA {
			buySignal = previousshortEMA <= previouslongEMA && currentshortEMA > currentlongEMA 
			sellSignal = previousshortEMA >= previouslongEMA && currentshortEMA < currentlongEMA
		}
		if ts.Strategy.UseRSI {
			buySignal = currentRSI < 30 
			sellSignal = currentRSI > 70 
		}
		if ts.Strategy.UseMACD {
			buySignal = previousMACDLine <= previousSignalLine && currentMACDLine > currentSignalLine && currentMacdHistogram > 0
			sellSignal = previousMACDLine >= previousSignalLine && currentMACDLine < currentSignalLine && currentMacdHistogram < 0
		}
		if ts.Strategy.UseBollinger {
			buySignal = ts.ClosingPrices[len(ts.ClosingPrices)-1] < currentLowerBand 
			sellSignal = ts.ClosingPrices[len(ts.ClosingPrices)-1] > currentUpperBand
		}
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

// RiskManagement applies risk management rules to limit potential losses.
func (ts *TradingSystem) RiskManagement() bool{
	// Calculate stop-loss price based on the fixed percentage of risk per trade.
	ts.StopLossPrice = ts.EntryPrice * (1.0 - ts.RiskPercentage)

	// Calculate position size based on the fixed percentage of risk per trade.
	ts.RiskAmount = ts.CurrentBalance * ts.RiskPercentage
	ts.PositionSize = ts.RiskAmount / ts.StopLossPrice

	// Check if the current price breaches the stop-loss level
	if ts.InTrade && (ts.EntryPrice <= ts.StopLossPrice) {
		fmt.Printf("Stop-Loss Triggered at %v - SELL\n", ts.EntryPrice)

		// Record Signal for ploting graph later
		ts.Signals[ts.DataPoint] = "Sell" 

		// Record exit price for calculating profit/loss
		exitPrice := ts.CurrentPrice

		// Calculate profit/loss for the trade.
		tradeProfitLoss := CalculateProfitLoss(ts.EntryPrice, exitPrice, ts.PositionSize)
		ts.TotalProfitLoss += tradeProfitLoss

		// Update the current balance after each trade.
		ts.CurrentBalance += tradeProfitLoss
		return true
	}

	return false
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) UpdateHistoricalData() error {
	var err error
	ts.HistoricalData, err = GetCandlesFromExch()
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
func (ts *TradingSystem) CalculateMACD() ([]float64, []float64, []float64) {
	longEMA, shortEMA, _, err := ts.CandleExponentialMovingAverage()
	if err != nil {
		log.Fatalf("Error: in CalclateMACD while tring to get EMA")
	}
	// Calculate MACD line
	macdLine := make([]float64, len(ts.ClosingPrices))
	for i := range ts.ClosingPrices {
		macdLine[i] = shortEMA[i] - longEMA[i]
	}

	// Calculate signal line using the MACD line
	signalLine := CalculateExponentialMovingAverage(macdLine, ts.MacdSignalPeriod)

	// Calculate MACD Histogram
	macdHistogram := make([]float64, len(ts.ClosingPrices))
	for i := range ts.ClosingPrices {
		macdHistogram[i] = macdLine[i] - signalLine[i]
	}

	return macdLine, signalLine, macdHistogram
}

func (ts *TradingSystem) CandleExponentialMovingAverage() ([]float64, []float64, []int64, error) {
	ema55 := ema.NewEma(alphaFromN(ts.LongPeriod))
	ema15 := ema.NewEma(alphaFromN(ts.ShortPeriod))

	m55 := make([]float64, len(ts.HistoricalData))
	m15 := make([]float64, len(ts.HistoricalData))
	timestamps := make([]int64, len(ts.HistoricalData))
	for k, candle := range ts.HistoricalData {
		ema55.Step(candle.Close)
		ema15.Step(candle.Close)
		m55[k] = ema55.Compute()
		m15[k] = ema15.Compute()
		timestamps[k] = candle.Timestamp
	}
	return m55, m15, timestamps, nil
}
// Initialize initializes the TradingSystem and fetches historical data.
func (ts *TradingSystem) Initialize() error {
	ts.RsiPeriod = 14 // Define RSI period parameter.
	ts.MacdSignalPeriod = 9 // Define MACD period parameter.
	ts.BollingerPeriod = 20 // Define Bollinger Bands parameter.
	ts.BollingerNumStdDev = 2.0 // Define Bollinger Bands parameter.
	ts.TransactionCost = 0.001 // Define 0.1% transaction cost
	ts.Slippage = 0.01 // Define 1% slippage
	ts.RiskPercentage = 0.05 // Define risk management parameter 5% stop-loss
	ts.InitialCapital = 1000.0 //Initial Capital for simulation on backtesting
	ts.CurrentBalance = 0.0 //continer to hold the balance
	
	// Fetch historical data from the exchange
	err := ts.UpdateHistoricalData()
	if err != nil {
		return err
	}
	// Perform any other necessary initialization steps here
    fmt.Println(ts.ClosingPrices)
	ts.Signals = make([]string, len(ts.ClosingPrices)) // Holder of generated trading signals
	ts.Timestamps = make([]int64, len(ts.ClosingPrices)) // Holder of generated trading signals
	ts.Strategy.UseEMA = true

	return nil
}

// CalculateProfitLoss calculates the profit or loss from a trade.
func CalculateProfitLoss(entryPrice, exitPrice, positionSize float64) float64 {
	return (exitPrice - entryPrice) * positionSize
}

func Backtest(ts *TradingSystem) {
	// Initialize variables for tracking trading performance.
	tradeCount := 0
	riskAmount := 0.0
	exitPrice := 0.0
	ts.TotalProfitLoss = 0.0

	// Initialize the current balance with the initial capital.
	ts.CurrentBalance = ts.InitialCapital

	// Simulate the backtesting process using historical price data.
	for ts.DataPoint, ts.CurrentPrice = range ts.ClosingPrices {
		// Execute the trade if entry conditions are met.
		if !ts.InTrade && ts.EntryRule() {
			fmt.Printf("Backtesting: Trade executed at %v - BUY\n", ts.CurrentPrice)
			tradeCount++
			
			// Record entry price for calculating profit/loss later.
			ts.EntryPrice = ts.CurrentPrice
			
			// Calculate position size based on the fixed percentage of risk per trade.
			ts.RiskManagement() //make sure entry price is set before calling riskmanagement

			// Mark that we are in a trade.
			ts.InTrade = true

			// Record Signal for ploting graph later
			ts.Signals[ts.DataPoint] = "Buy" 
			
            // Calculate available CurrentBalance
            ts.CurrentBalance -= riskAmount

		// Close the trade if exit conditions are met.
		}else if ts.InTrade && ts.ExitRule() {
			fmt.Printf("Backtesting: Trade closed at %v - SELL\n", ts.CurrentPrice)
			tradeCount++

			if !ts.RiskManagement(){			
				// Record exit price for calculating profit/loss
				exitPrice = ts.CurrentPrice
				
				// Record Signal for ploting graph later
				ts.Signals[ts.DataPoint] = "Sell"

				// Calculate profit/loss for the trade.
				tradeProfitLoss := CalculateProfitLoss(ts.EntryPrice, exitPrice, ts.PositionSize)
				ts.TotalProfitLoss += tradeProfitLoss

				// Update the current balance after each trade.
				ts.CurrentBalance += tradeProfitLoss
			}

			// Mark that we are no longer in a trade.
			ts.InTrade = false

			// Implement risk management and update capital after each trade
			// For simplicity, we'll skip this step in this basic example

		}else{
			if !ts.RiskManagement(){
				ts.Signals[ts.DataPoint] = "Hold" // No Signal - Hold Position
			}
        }

		// Perform other necessary tasks, such as recording trading performance.
		// For real-world backtesting, you may need to track portfolio value, profit/loss, etc.
	}

	// Print the overall trading performance after backtesting.
	fmt.Printf("\nBacktesting Summary:\n")
	fmt.Printf("Total Trades: %d\n", tradeCount)
	fmt.Printf("Total Profit/Loss: %.2f\n", ts.TotalProfitLoss)
	fmt.Printf("Final Capital: %.2f\n", ts.CurrentBalance)
	
    err := CreateLineChartWithSignals(ts.Timestamps, ts.ClosingPrices, ts.Signals)
	if err != nil {
		fmt.Println("Error creating Line Chart with signals:", err)
		return
	}
}

func CalculateSimpleMovingAverage(data []float64, period int) []float64 {
	if period <= 0 || len(data) < period {
		return nil
	}

	sma := make([]float64, len(data)-period+1)
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
func alphaFromN(N int) float64 {
	var n = float64(N)
	return 2. / (n + 1.)
}

// CalculateExponentialMovingAverage calculates the Exponential Moving Average (EMA) for the given data and period.
func CalculateExponentialMovingAverage(data []float64, period int) []float64 {
	ema := make([]float64, len(data))
	smoothingFactor := 2.0 / (float64(period) + 1.0)

	// Calculate the initial SMA as the sum of the first 'period' data points divided by 'period'.
	sma := 0.0
	for i := 0; i < period; i++ {
		sma += data[i]
	}
	ema[period-1] = sma / float64(period)

	// Calculate the EMA for the remaining data points.
	for i := period; i < len(data); i++ {
		ema[i] = (data[i] - ema[i-1]) * smoothingFactor + ema[i-1]
	}

	return ema
}

// CalculateRSI calculates the Relative Strength Index (RSI) for the given data and period.
func CalculateRSI(data []float64, period int) []float64 {
	rsiValues := make([]float64, 0, len(data))

	if len(data) < period {
		return rsiValues
	}

	// Calculate the first average gain and loss for the initial period.
	var avgGain, avgLoss float64
	for i := 1; i < period; i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			avgGain += change
		} else {
			avgLoss += -change
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI values for the remaining data.
	for i := period; i < len(data); i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			avgGain = (avgGain*(float64(period-1)) + change) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*(float64(period-1)) + (-change)) / float64(period)
		}

		rs := avgGain / avgLoss
		rsi := 100.0 - (100.0 / (1.0 + rs))
		rsiValues = append(rsiValues, rsi)
	}

	return rsiValues
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

// Function to generate trading signals using Moving Average Strategy
func GenerateSignals(closingPrices []float64, shortPeriod, longPeriod int) []string {
    // Calculate short-term and long-term EMAs using the functions we defined earlier
    shortEMA := CalculateExponentialMovingAverage(closingPrices, shortPeriod)
    longEMA := CalculateExponentialMovingAverage(closingPrices, longPeriod)

    signals := make([]string, len(closingPrices))
    for i := 1; i < len(closingPrices); i++ {
        if shortEMA[i] > longEMA[i] && shortEMA[i-1] <= longEMA[i-1] {
            signals[i] = "Buy" // Golden Cross - Buy Signal
        } else if shortEMA[i] < longEMA[i] && shortEMA[i-1] >= longEMA[i-1] {
            signals[i] = "Sell" // Death Cross - Sell Signal
        } else {
            signals[i] = "Hold" // No Signal - Hold Position
        }
    }
	fmt.Println("longEMA =", longEMA[len(closingPrices)-1], "shortEMA =", shortEMA[len(closingPrices)-1])
	fmt.Println()
	fmt.Println("signals =", signals, " = ", len(signals), "counts")
	fmt.Println()
    return signals
}

func GetCandlesFromExch() ([]model.Candlestick, error){
    exchConfig := config.NewExchangesConfig()
    hitbtcExch, err := hitbtcapi.NewExchServices(exchConfig)
    if err != nil{
        log.Fatal(err)
    }

	startTime := time.Now().Add(-5 * 24 * time.Hour).Unix() // 3 days ago
	endTime := time.Now().Unix()

    return hitbtcExch.FetchHistoricalCandlesticks(hitbtcExch.Exch.Symbol, hitbtcExch.Exch.Interval, startTime, endTime)
}




