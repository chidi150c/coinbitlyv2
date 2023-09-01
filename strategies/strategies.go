package strategies

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"sync"

	"coinbitly.com/binanceapi"
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
	Mu sync.Mutex // Mutex for protecting concurrent writes
	// HistoricalData  []model.Candle
	Symbol                   string
	ClosingPrices            []float64
	Container1               []float64
	Container2               []float64
	Timestamps               []int64
	Signals                  []string
	APIServices              model.APIServices
	TransactionCost          float64
	Slippage                 float64
	InitialCapital           float64
	PositionSize             float64
	EntryPrice               []float64
	InTrade                  bool
	QuoteBalance             float64
	BaseBalance              float64
	RiskCost                 float64
	DataPoint                int
	CurrentPrice             float64
	EntryQuantity            []float64
	Scalping                 string
	StrategyCombLogic        string
	EntryCostLoss            []float64
	TradeCount               int
	EnableStoploss           bool
	StopLossTrigered         bool
	StopLossRecover          []float64
	RiskFactor               float64
	MaxDataSize              int
	Log                      *log.Logger
	ShutDownCh               chan string
	EpochTime                time.Duration
	CSVWriter                *csv.Writer
	RiskProfitLossPercentage float64
	ChartChan                chan model.ChartData
	BaseCurrency             string //in Binance is called BaseAsset
	QuoteCurrency            string //in Binance is called QuoteAsset
	MiniQty                  float64
}

// NewTradingSystem(): This function initializes the TradingSystem and fetches
// historical data from the exchange using the GetCandlesFromExch() function.
// It sets various strategy parameters like short and long EMA periods, RSI period,
// MACD signal period, Bollinger Bands period, and more.
func NewTradingSystem(BaseCurrency string, liveTrading bool, loadFrom string) (*TradingSystem, error) {
	// Initialize the trading system.
	ts := &TradingSystem{}
	ts.RiskFactor = 2.0         // Define 1% slippage
	ts.TransactionCost = 0.0009 // Define 0.1% transaction cost
	ts.Slippage = 0.0009
	ts.InitialCapital = 100.0             //Initial Capital for simulation on backtesting
	ts.RiskProfitLossPercentage = 0.00025 //percentage of Initial Capital used to represent Target profit or stoplooss amount
	ts.QuoteBalance = ts.InitialCapital   //continer to hold the balance
	ts.Scalping = ""                      //"UseTA"
	ts.StrategyCombLogic = "OR"
	ts.EnableStoploss = true
	ts.StopLossRecover = append(ts.StopLossRecover, math.MaxFloat64) //
	ts.MaxDataSize = 500
	ts.ShutDownCh = make(chan string)
	ts.EpochTime = time.Second * 30
	ts.ChartChan = make(chan model.ChartData, 1)
	ts.BaseCurrency = BaseCurrency
	if liveTrading {
		err := ts.LiveUpdate(loadFrom)
		if err != nil {
			return &TradingSystem{}, err
		}
	} else {
		err := ts.UpdateHistoricalData(loadFrom)
		if err != nil {
			return &TradingSystem{}, err
		}
	}
	return ts, nil
}
func (ts *TradingSystem) NewAppData() *model.AppData {
	md := &model.AppData{}
	md.DataPoint = 0
	md.Strategy = "EMA"
	md.ShortPeriod = 10 // Define moving average short period for the strategy.
	md.LongPeriod = 30  // Define moving average long period for the strategy.
	md.ShortEMA = 0.0
	md.LongEMA = 0.0
	md.TargetProfit = ts.InitialCapital * ts.RiskProfitLossPercentage
	md.TargetStopLoss = ts.InitialCapital * ts.RiskProfitLossPercentage
	md.RiskPositionPercentage = 0.1 // Define risk management parameter 5% balance
	md.TotalProfitLoss = 0.0
	return md
}
func (ts *TradingSystem) TickerQueueAdjustment() {
	if len(ts.ClosingPrices) > ts.MaxDataSize {
		ts.ClosingPrices = ts.ClosingPrices[1:] // Remove the oldest element
		ts.Timestamps = ts.Timestamps[1:]       // Remove the oldest element
		ts.Signals = ts.Signals[1:]             // Remove the oldest element
		ts.DataPoint--
	}
}

// LiveTrade(): This function performs live trading process using live
// data. It gets live ticker prices from exchange, checks for entry and exit
// conditions, and executes trades accordingly. It also tracks trading performance
// and updates the current balance after each trade.
func (ts *TradingSystem) LiveTrade(loadFrom string) {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl, syscall.SIGINT)
	md := ts.NewAppData()
	fmt.Println("App started. Press Ctrl+C to exit.")
	go ts.ShutDown(md, sigchnl)
	ts.BaseBalance = 0.0
	ts.QuoteBalance = ts.InitialCapital
	// Initialize variables for tracking trading performance.
	ts.TradeCount = 0
	var err error
	for {
		ts.CurrentPrice, err = ts.APIServices.FetchTicker(ts.Symbol)
		if err != nil {
			log.Fatal("Error Reporting Live Trade: ", err)
		}
		ts.DataPoint++
		md.DataPoint++
		ts.ClosingPrices = append(ts.ClosingPrices, ts.CurrentPrice)
		ts.Timestamps = append(ts.Timestamps, time.Now().Unix())

		//Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading
		//Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading
		md.TotalProfitLoss = ts.Trading(md, loadFrom)

		ts.TickerQueueAdjustment() //At this point you have all three(ts.ClosingPrices, ts.Timestamps and ts.Signals) assigned

		err = ts.Reporting(md, "Live Trading")
		if err != nil {
			fmt.Println("Error Reporting Live Trade: ", err)
			return
		}

		time.Sleep(ts.EpochTime)
		// err = ts.APIServices.WriteTickerToDB(ts.ClosingPrices[ts.DataPoint], ts.Timestamps[ts.DataPoint])
		// if (err != nil) && (!strings.Contains(fmt.Sprintf("%v", err), "Skipping write")) {
		// 	log.Fatalf("Error: writing to influxDB: %v", err)
		// }
	}
}

// Backtest(): This function simulates the backtesting process using historical
// price data. It iterates through the closing prices, checks for entry and exit
// conditions, and executes trades accordingly. It also tracks trading performance
// and updates the current balance after each trade.
func (ts *TradingSystem) Backtest(loadFrom string) {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)
	//Count,Strategy,ShortPeriod,LongPeriod,ShortMACDPeriod,LongMACDPeriod,
	//SignalMACDPeriod,RSIPeriod,StochRSIPeriod,SmoothK,SmoothD,RSIOverbought,RSIOversold,
	//StRSIOverbought,StRSIOversold,BollingerPeriod,BollingerNumStdDev,TargetProfit,
	//TargetStopLoss,RiskPositionPercentage
	backT := []*model.AppData{ //StochRSI
		{0, "EMA", 20, 55, 0.0, 0.0, 0.5, 3.0, 0.25, 0.0},
	}

	fmt.Println("App started. Press Ctrl+C to exit.")
	for i, md := range backT {
		md = ts.NewAppData()
		ts.BaseBalance = 0.0
		ts.QuoteBalance = ts.InitialCapital
		// Initialize variables for tracking trading performance.
		ts.TradeCount = 0
		// Simulate the backtesting process using historical price data.
		for ts.DataPoint, ts.CurrentPrice = range ts.ClosingPrices {
			_ = ts.Trading(md, loadFrom)
		}
		err := ts.Reporting(md, "Backtesting")
		if err != nil {
			fmt.Println("Error Reporting BackTest Trading: ", err)
			return
		}
		fmt.Println()
		fmt.Println("Press Ctrl+C to continue...")
		<-sigchnl
		if i == len(backT)-1 {
			fmt.Println("End of Backtesting: Press Ctrl+C again to shutdown...")
			sigchnl = make(chan os.Signal, 1)
			signal.Notify(sigchnl)
			ts.ShutDown(md, sigchnl)
		}
	}
}

func (ts *TradingSystem) Trading(md *model.AppData, loadFrom string) (totalProfitLoss float64) {
	// Execute the trade if entry conditions are met.
	if (!ts.InTrade) && (ts.EntryRule(md)) && (ts.CurrentPrice <= ts.StopLossRecover[len(ts.StopLossRecover)-1]) {
		// Record entry price for calculating profit/loss and stoploss later.
		ts.EntryPrice = append(ts.EntryPrice, ts.CurrentPrice)
		// Execute the buy order using the ExecuteStrategy function.
		resp, err := ts.ExecuteStrategy(md, "Buy")
		if err != nil {
			// fmt.Println("Error executing buy order:", err)
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
		} else if strings.Contains(resp, "BUY") {
			// Record Signal for plotting graph later
			ts.Signals = append(ts.Signals, "Buy")
			ts.Log.Println(resp)
			// Mark that we are in a trade.
			ts.InTrade = true
			ts.TradeCount++
		} else {
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Positio
		}
		// Close the trade if exit conditions are met.
	} else if ts.InTrade || (ts.StopLossTrigered && (ts.CurrentPrice > ts.StopLossRecover[len(ts.StopLossRecover)-1])) {
		if ts.ExitRule(md) {
			// Execute the sell order using the ExecuteStrategy function.
			resp, err := ts.ExecuteStrategy(md, "Sell")
			if err != nil {
				// fmt.Println("Error:", err, " at:", ts.CurrentPrice, ", TargetStopLoss:", md.TargetStopLoss)
				ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
			} else if strings.Contains(resp, "SELL") {
				// Record Signal for plotting graph later.
				ts.Signals = append(ts.Signals, "Sell")
				ts.Log.Println(resp)
				if len(ts.StopLossRecover) > 1 {
					ts.EntryPrice = ts.EntryPrice[:len(ts.EntryPrice)-1]
					ts.EntryCostLoss = ts.EntryCostLoss[:len(ts.EntryCostLoss)-1]
					ts.EntryQuantity = ts.EntryQuantity[:len(ts.EntryQuantity)-1]
					ts.StopLossRecover = ts.StopLossRecover[:len(ts.StopLossRecover)-1]
					md.RiskPositionPercentage /= ts.RiskFactor //reduce riskPosition by a factor
					ts.InTrade = true
					ts.StopLossTrigered = true
				} else if len(ts.StopLossRecover) == 1 {
					// Mark that we are no longer in a trade.
					ts.InTrade = false
					ts.StopLossTrigered = false
					ts.StopLossRecover[len(ts.StopLossRecover)-1] = math.MaxFloat64
					ts.EntryPrice = []float64{}
					ts.EntryCostLoss = []float64{}
					ts.EntryQuantity = []float64{}
				}
				// ts.RiskStopLossPercentage /= ts.RiskFactor

				ts.TradeCount++
			} else {
				ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
			}
		} else {
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Positio
		}
	} else {
		ts.EntryRule(md)
		ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
	}
	return md.TotalProfitLoss
}

// ExecuteStrategy executes the trade based on the provided trade action and current price.
// The tradeAction parameter should be either "Buy" or "Sell".
func (ts *TradingSystem) ExecuteStrategy(md *model.AppData, tradeAction string) (string, error) {
	// Calculate position size based on the fixed percentage of risk per trade.
	resp := ts.RiskManagement(md) // make sure entry price is set before calling risk management
	if strings.Contains(resp, "Marked!!!") {
		return resp, nil
	}
	switch tradeAction {
	case "Buy":
		if ts.InTrade {
			return "", fmt.Errorf("cannot execute a buy order while already in a trade")
		}
		localProfitLoss := 0.0
		// Update the current balance after the trade (considering transaction fees and slippage).
		transactionCost := ts.TransactionCost * ts.CurrentPrice * ts.PositionSize
		slippageCost := ts.Slippage * ts.CurrentPrice * ts.PositionSize
		localProfitLoss -= transactionCost + slippageCost
		//check if there is enough Quote (USDT) for the buy transaction
		tsPositionSize := ts.PositionSize
		if ts.QuoteBalance < (ts.PositionSize*ts.CurrentPrice)+transactionCost+slippageCost {
			ts.Log.Printf("Unable to buY!!!! Insufficient QuoteBalance: %.8f (ts.PositionSize %.8f* ts.CurrentPrice %.8f) = %.8f + transactionCost %.8f+ slippageCost %.8f", ts.QuoteBalance, ts.PositionSize, ts.CurrentPrice, ts.PositionSize*ts.CurrentPrice, transactionCost, slippageCost)
			tsPositionSize = (ts.QuoteBalance - (transactionCost + slippageCost)) / ts.CurrentPrice
			if tsPositionSize < ts.MiniQty {
				return "", fmt.Errorf("cannot execute a buy order due to insufficient QuoteBalance: %.8f (ts.PositionSize %.8f* ts.CurrentPrice %.8f) = %.8f + transactionCost %.8f+ slippageCost %.8f", ts.QuoteBalance, ts.PositionSize, ts.CurrentPrice, ts.PositionSize*ts.CurrentPrice, transactionCost, slippageCost)
			}
		}
		// Record entry price for calculating profit/loss later.
		ts.EntryQuantity = append(ts.EntryQuantity, tsPositionSize)

		entryOrderID, err := ts.APIServices.PlaceLimitBuyOrder(ts.Symbol, ts.EntryPrice[len(ts.EntryPrice)-1], ts.EntryQuantity[len(ts.EntryQuantity)-1])
		if err != nil {
			fmt.Println("Error placing entry order:", err)
			ts.Log.Println("Error placing entry order:", err)
			return "", fmt.Errorf("Error placing entry order: %v", err)
		}

		fmt.Printf("Entry order placed with ID: %d\n", entryOrderID)
		// Update the profit, quote and base balances after the trade.
		ts.QuoteBalance -= (ts.PositionSize * ts.CurrentPrice) + transactionCost + slippageCost
		ts.BaseBalance += ts.PositionSize
		md.TotalProfitLoss += localProfitLoss
		ts.EntryCostLoss = append(ts.EntryCostLoss, localProfitLoss)
		// Mark that we are in a trade.
		ts.InTrade = true

		nextsEllPt := (md.TargetProfit / ts.PositionSize) + ts.EntryPrice[len(ts.EntryPrice)-1]
		nextLossPt := (-md.TargetStopLoss / ts.PositionSize) + ts.EntryPrice[len(ts.EntryPrice)-1]
		resp := fmt.Sprintf("- BUY at EntryPrice[%d]: %.8f, EntryQuant[%d]: %.8f, QBal: %.8f, BBal: %.8f, GlobalP&L %.2f LocalP&L: %.8f PosPcent: %.8f tsDataPt: %d mdDataPt: %d nextsEllPt %.8f nextLossPt %.8f TargProfit %.8f TargLoss %.8f\n",
			len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, localProfitLoss, md.RiskPositionPercentage, ts.DataPoint, md.DataPoint, nextsEllPt, nextLossPt, md.TargetProfit, -md.TargetStopLoss)
		return resp, nil
	case "Sell":
		if !ts.InTrade {
			return "", fmt.Errorf("cannot execute a sell order without an existing trade")
		}

		// Calculate profit/loss for the trade.
		exitPrice := ts.CurrentPrice
		tsEntryQuantity := ts.EntryQuantity[len(ts.EntryQuantity)-1]
		localProfitLoss := CalculateProfitLoss(ts.EntryPrice[len(ts.EntryPrice)-1], exitPrice, tsEntryQuantity)
		transactionCost := ts.TransactionCost * exitPrice * tsEntryQuantity
		slippageCost := ts.Slippage * exitPrice * tsEntryQuantity
		localProfitLoss -= transactionCost + slippageCost + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]

		// (localProfitLoss < transactionCost+slippageCost+md.TargetProfit)
		if (ts.BaseBalance < tsEntryQuantity) || (localProfitLoss < md.TargetProfit) {
			if ts.BaseBalance < tsEntryQuantity {
				return "", fmt.Errorf("cannot execute a sell order insufficient BaseBalance: %.8f needed up to: %.8f", ts.BaseBalance, tsEntryQuantity)
			}
			ts.Log.Println("TA Signalled: SELL at currentPrice:", exitPrice, "localProfitLoss:", localProfitLoss, "expecting Profit:", md.TargetProfit)
			return "", fmt.Errorf("cannot execute a sell order this trade localProfitLoss: %.8f, target profit: %.8f", localProfitLoss, md.TargetProfit)
		}
		exitOrderID, err := ts.APIServices.PlaceLimitSellOrder(ts.Symbol, exitPrice, tsEntryQuantity)
		if err != nil {
			fmt.Println("Error placing exit order:", err)
			ts.Log.Println("Error placing exit order:", err)
			return "", fmt.Errorf("Error placing exit order: %v", err)
		}
		fmt.Printf("Exit order placed with ID: %d\n", exitOrderID)
		// Mark that we are no longer in a trade.
		ts.InTrade = false

		// Update the totalP&L, quote and base balances after the trade.
		ts.QuoteBalance += (tsEntryQuantity * exitPrice) - transactionCost - slippageCost
		ts.BaseBalance -= tsEntryQuantity
		md.TotalProfitLoss += localProfitLoss

		resp := fmt.Sprintf("- SELL at ExitPrice: %.8f, EntryPrice[%d]: %.8f, EntryQuant[%d]: %.8f, QBal: %.8f, BBal: %.8f, GlobalP&L %.2f LocalP&L: %.8f PosPcent: %.8f tsDataPt: %d mdDataPt: %d \n",
			exitPrice, len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, localProfitLoss, md.RiskPositionPercentage, ts.DataPoint, md.DataPoint)
		return resp, nil
	default:
		return "", fmt.Errorf("invalid trade action: %s", tradeAction)
	}
}

// RiskManagement applies risk management rules to limit potential losses.
// It calculates the stop-loss price based on the fixed percentage of risk per trade and the position size.
// If the current price breaches the stop-loss level, it triggers a sell signal and exits the trade.
func (ts *TradingSystem) RiskManagement(md *model.AppData) string {
	// Calculate position size based on the fixed percentage of risk per trade.
	ts.RiskCost = ts.QuoteBalance * md.RiskPositionPercentage
	ts.PositionSize = ts.RiskCost / ts.CurrentPrice

	// if this function was call for Buy Return to the caller at this point
	if !ts.InTrade {
		return "false"
	}
	//if you got here then this function was call for sell

	// Calculate profit/loss for the trade.
	exitPrice := ts.CurrentPrice
	tsEntryQuantity := ts.EntryQuantity[len(ts.EntryQuantity)-1]
	localProfitLoss := CalculateProfitLoss(ts.EntryPrice[len(ts.EntryPrice)-1], exitPrice, tsEntryQuantity)
	transactionCost := ts.TransactionCost * exitPrice * tsEntryQuantity
	slippageCost := ts.Slippage * exitPrice * tsEntryQuantity
	localProfitLoss -= transactionCost + slippageCost + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]

	// Check if the current price breaches the stop-loss level to trigger a buy signal.
	if (localProfitLoss <= -md.TargetStopLoss) && ts.EnableStoploss { //&& (ts.BaseBalance >= tsEntryQuantity)

		// I commented it out to make stoploss just a marker
		// // Update the quote and base balances after the trade.
		// ts.QuoteBalance -= (tsEntryQuantity * exitPrice) + transactionCost + slippageCost
		// ts.BaseBalance += tsEntryQuantity
		// md.TotalProfitLoss += localProfitLoss

		// Mark that stoploss is triggered.
		ts.InTrade = false
		ts.StopLossTrigered = true
		md.RiskPositionPercentage *= ts.RiskFactor                       //increase riskPosition by a factor
		ts.StopLossRecover = append(ts.StopLossRecover, ts.CurrentPrice) //* (1.0 - ts.RiskStopLossPercentage)
		ts.Log.Println("Stoploss Marked!!! at this Price:", exitPrice, "based on EntryPrice[", len(ts.EntryPrice)-1,
			"] = ", ts.EntryPrice[len(ts.EntryPrice)-1], "currentLoss:", localProfitLoss, "expecting Loss:", -md.TargetStopLoss)
		// I commented it out to make stoploss just a marker
		resp := fmt.Sprintf("Stoploss Marked!!! at this Price: %.8f currentLoss: %.8f expecting Loss: %.8f", exitPrice, localProfitLoss, -md.TargetStopLoss)
		// 	ts.CurrentPrice, tsEntryQuantity, ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, localProfitLoss, md.RiskPositionPercentage, ts.DataPoint, md.DataPoint)
		return resp
	} else {
		ts.Log.Println("Stoploss not Marked yet at this Price:", exitPrice, "currentLoss:", localProfitLoss, "expecting Loss:", -md.TargetStopLoss)
	}
	return "false"
}

// TechnicalAnalysis(): This function performs technical analysis using the
// calculated moving averages (short and long EMA), RSI, MACD line, and Bollinger
// Bands. It determines the buy and sell signals based on various strategy rules.
func (ts *TradingSystem) TechnicalAnalysis(md *model.AppData, Action string) (buySignal, sellSignal bool) {
	// Calculate moving averages (MA) using historical data.
	longEMA, shortEMA, _, err := CandleExponentialMovingAverage(ts.ClosingPrices, ts.Timestamps, md.LongPeriod, md.ShortPeriod)
	if err != nil {
		log.Printf("Error: in TechnicalAnalysis Unable to get EMA: %v", err)
		md.LongEMA, md.ShortEMA = 0.0, 0.0
	} else {
		md.LongEMA, md.ShortEMA = longEMA[ts.DataPoint], shortEMA[ts.DataPoint]
	}
	// Determine the buy and sell signals based on the moving averages, RSI, MACD line, and Bollinger Bands.
	if len(shortEMA) > 7 && len(longEMA) > 7 && ts.DataPoint >= 7 {
		if strings.Contains(md.Strategy, "EMA") && ts.DataPoint > 1 {
			ts.Container1 = shortEMA
			ts.Container2 = longEMA

			// buySignal = (shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint])
			// sellSignal = (shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint])

			buySignal = longEMA[ts.DataPoint-6] > shortEMA[ts.DataPoint-7] &&
				(longEMA[ts.DataPoint-6]-shortEMA[ts.DataPoint-6] >= longEMA[ts.DataPoint-7]-shortEMA[ts.DataPoint-7]) &&
				(longEMA[ts.DataPoint-5]-shortEMA[ts.DataPoint-5] >= longEMA[ts.DataPoint-6]-shortEMA[ts.DataPoint-6]) &&
				(longEMA[ts.DataPoint-4]-shortEMA[ts.DataPoint-4] >= longEMA[ts.DataPoint-5]-shortEMA[ts.DataPoint-5]) &&
				(longEMA[ts.DataPoint-3]-shortEMA[ts.DataPoint-3] >= longEMA[ts.DataPoint-4]-shortEMA[ts.DataPoint-4]) &&
				(longEMA[ts.DataPoint-2]-shortEMA[ts.DataPoint-2] >= longEMA[ts.DataPoint-3]-shortEMA[ts.DataPoint-3]) &&
				(longEMA[ts.DataPoint-1]-shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-2]-shortEMA[ts.DataPoint-2]) &&
				(longEMA[ts.DataPoint]-shortEMA[ts.DataPoint] < longEMA[ts.DataPoint-1]-shortEMA[ts.DataPoint-1])

			sellSignal = shortEMA[ts.DataPoint-6] > longEMA[ts.DataPoint-7] &&
				(shortEMA[ts.DataPoint-6]-longEMA[ts.DataPoint-6] >= shortEMA[ts.DataPoint-7]-longEMA[ts.DataPoint-7]) &&
				(shortEMA[ts.DataPoint-5]-longEMA[ts.DataPoint-5] >= shortEMA[ts.DataPoint-6]-longEMA[ts.DataPoint-6]) &&
				(shortEMA[ts.DataPoint-4]-longEMA[ts.DataPoint-4] >= shortEMA[ts.DataPoint-5]-longEMA[ts.DataPoint-5]) &&
				(shortEMA[ts.DataPoint-3]-longEMA[ts.DataPoint-3] >= shortEMA[ts.DataPoint-4]-longEMA[ts.DataPoint-4]) &&
				(shortEMA[ts.DataPoint-2]-longEMA[ts.DataPoint-2] >= shortEMA[ts.DataPoint-3]-longEMA[ts.DataPoint-3]) &&
				(shortEMA[ts.DataPoint-1]-longEMA[ts.DataPoint-1] >= shortEMA[ts.DataPoint-2]-longEMA[ts.DataPoint-2]) &&
				(shortEMA[ts.DataPoint]-longEMA[ts.DataPoint] < shortEMA[ts.DataPoint-1]-longEMA[ts.DataPoint-1])

			if buySignal && (Action == "Entry") {
				ts.Log.Println("TA Signalled: Buy:", buySignal, "at currentPrice:", ts.CurrentPrice)
			} else if sellSignal && (Action == "Exit") {
				ts.Log.Println("TA Signalled: Sell:", sellSignal, "at currentPrice:", ts.CurrentPrice)
			}

		}
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
func (ts *TradingSystem) EntryRule(md *model.AppData) bool {
	// Combine the results of technical and fundamental analysis to decide entry conditions.
	technicalBuy, _ := ts.TechnicalAnalysis(md, "Entry")
	// fundamentalBuy, fundamentalSell := ts.FundamentalAnalysis()
	return technicalBuy
}

// ExitRule defines the exit conditions for a trade.
func (ts *TradingSystem) ExitRule(md *model.AppData) bool {
	// Combine the results of technical and fundamental analysis to decide exit conditions.
	_, technicalSell := ts.TechnicalAnalysis(md, "Exit")
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
		infDB       *influxdb.CandleServices
	)

	if configParam, ok = config.NewExchangesConfig()[loadFrom]; ok {
		// Check if REQUIRED "InfluxDB" exists in the map
		if influxDBParam, ok := config.NewExchangesConfig()["InfluxDB"]; ok {
			//Initiallize InfluDB config Parameters and instanciate its struct
			infDB, err = influxdb.NewAPIServices(influxDBParam)
			if err != nil {
				log.Fatalf("Error getting Required services from InfluxDB: %v", err)
			}
		} else {
			fmt.Println("Error Internal: Exch Required InfluxDB")
			return errors.New("Required InfluxDB services not contracted")
		}
	} else {
		fmt.Println("Error Internal: Exch name not", loadFrom)
		return errors.New("Exch name not " + loadFrom)
	}

	switch loadFrom {
	case "HitBTC":
		exch, err = hitbtcapi.NewAPIServices(infDB, configParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from HitBTC: %v", err)
		}
	case "Binance":
		exch, err = binanceapi.NewAPIServices(infDB, configParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnet":
		exch, err = binanceapi.NewAPIServices(infDB, configParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "InfluxDB":
		exch = infDB
	default:
		return errors.Errorf("Error updating historical data from %s invalid loadFrom tag \"%s\" ", loadFrom, loadFrom)
	}

	HistoricalData, err := exch.FetchCandles(configParam.Symbol, configParam.CandleInterval, configParam.CandleStartTime, configParam.CandleEndTime)
	ts.APIServices = exch
	ts.BaseCurrency = configParam.BaseCurrency
	ts.QuoteCurrency = configParam.QuoteCurrency
	ts.Symbol = configParam.Symbol
	ts.MiniQty, err = exch.FetchMiniQuantity(ts.Symbol)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	// Extract the closing prices from the candles data
	for _, candle := range HistoricalData {
		ts.ClosingPrices = append(ts.ClosingPrices, candle.Close)
		ts.Timestamps = append(ts.Timestamps, candle.Timestamp)
		if loadFrom != "InfluxDB" {
			// err := ts.APIServices.WriteCandleToDB(candle.Close, candle.Timestamp)
			// if (err != nil) && (!strings.Contains(fmt.Sprintf("%v", err), "Skipping write")) {
			// 	log.Fatalf("Error: writing to influxDB: %v", err)
			// }
		}
	}
	return nil
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) LiveUpdate(loadFrom string) error {
	var (
		err         error
		exch        model.APIServices
		configParam *config.ExchConfig
		ok          bool
		infDB       *influxdb.CandleServices
	)
	
	if configParam, ok = config.NewExchangesConfig()[loadFrom]; ok {
		// Check if REQUIRED "InfluxDB" exists in the map
		if influxDBParam, ok := config.NewExchangesConfig()["InfluxDB"]; ok {
			//Initiallize InfluDB config Parameters and instanciate its struct
			infDB, err = influxdb.NewAPIServices(influxDBParam)
			if err != nil {
				log.Fatalf("Error getting Required services from InfluxDB: %v", err)
			}
		} else {
			fmt.Println("Error Internal: Exch Required InfluxDB")
			return errors.New("Required InfluxDB services not contracted")
		}
	} else {
		fmt.Println("Error Internal: Exch name not", loadFrom)
		return errors.New("Exch name not " + loadFrom)
	}

	switch loadFrom {
	case "HitBTC":
		exch, err = hitbtcapi.NewAPIServices(infDB, configParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from HitBTC: %v", err)
		}
	case "Binance":
		exch, err = binanceapi.NewAPIServices(infDB, configParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnet":
		exch, err = binanceapi.NewAPIServices(infDB, configParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "InfluxDB":
		exch = infDB
	default:
		return errors.Errorf("Error updating historical data from %s invalid loadFrom tag \"%s\" ", loadFrom, loadFrom)
	}
	ts.APIServices = exch
	ts.BaseCurrency = configParam.BaseCurrency
	ts.QuoteCurrency = configParam.QuoteCurrency
	ts.Symbol = configParam.Symbol
	ts.CurrentPrice, err = exch.FetchTicker(ts.Symbol)
	if err != nil {
		return err
	}
	ts.MiniQty, err = exch.FetchMiniQuantity(ts.Symbol)
	if err != nil {
		return err
	}
	// Mining data for historical analysis
	ts.DataPoint = 0
	ts.ClosingPrices = append(ts.ClosingPrices, ts.CurrentPrice)
	ts.Timestamps = append(ts.Timestamps, time.Now().Unix())
	ts.Signals = append(ts.Signals, "Hold")
	// err = ts.APIServices.WriteTickerToDB(ts.ClosingPrices[ts.DataPoint], ts.Timestamps[ts.DataPoint])
	// if (err != nil) && (!strings.Contains(fmt.Sprintf("%v", err), "Skipping write")) {
	// 	log.Fatalf("Error: writing to influxDB: %v", err)
	// }
	return nil
}

func (ts *TradingSystem) Reporting(md *model.AppData, from string) error {
	var err error

	if len(ts.Container1) > 0 && (!strings.Contains(md.Strategy, "Bollinger")) && (!strings.Contains(md.Strategy, "EMA")) && (!strings.Contains(md.Strategy, "MACD")) {
		err = ts.CreateLineChartWithSignals(md, ts.Timestamps, ts.ClosingPrices, ts.Signals, "")
		err = ts.CreateLineChartWithSignals(md, ts.Timestamps, ts.Container1, ts.Signals, "StrategyOnly")
	} else if len(ts.Container1) > 0 && strings.Contains(md.Strategy, "Bollinger") {
		err = ts.CreateLineChartWithSignalsV3(md, ts.Timestamps, ts.ClosingPrices, ts.Container1, ts.Container2, ts.Signals, "")
	} else if len(ts.Container1) > 0 && strings.Contains(md.Strategy, "EMA") {
		err = ts.CreateLineChartWithSignalsV3(md, ts.Timestamps, ts.ClosingPrices, ts.Container1, ts.Container2, ts.Signals, "")
	} else {
		err = ts.CreateLineChartWithSignals(md, ts.Timestamps, ts.ClosingPrices, ts.Signals, "")
	}
	if err != nil {
		return fmt.Errorf("Error creating Line Chart with signals: %v", err)
	}
	if from == "Live Trading" {
		err = ts.AppDatatoCSV(md)
		if err != nil {
			return fmt.Errorf("Error creating Writing to CSV File %v", err)
		}
	}
	return nil
}

func (ts *TradingSystem) ShutDown(md *model.AppData, sigchnl chan os.Signal) {
	for {
		select {
		case sig := <-sigchnl:
			if sig == syscall.SIGINT {
				fmt.Printf("Received signal: %v. Exiting...\n", sig)
				//Check if there is still asset remainning and sell off
				if ts.BaseBalance > 0.0 {
					// After sell off Update the quote and base balances after the trade.

					// Calculate profit/loss for the trade.
					exitPrice := ts.CurrentPrice
					localProfitLoss := CalculateProfitLoss(ts.EntryPrice[len(ts.EntryPrice)-1], exitPrice, ts.BaseBalance)
					transactionCost := ts.TransactionCost * exitPrice * ts.BaseBalance
					slippageCost := ts.Slippage * exitPrice * ts.BaseBalance

					// Store profit/loss for the trade.

					localProfitLoss -= transactionCost + slippageCost
					// md.TotalProfitLoss += localProfitLoss

					ts.Mu.Lock()
					ts.QuoteBalance += (ts.BaseBalance * exitPrice) - transactionCost - slippageCost
					ts.BaseBalance -= ts.BaseBalance
					ts.Signals = append(ts.Signals, "Sell")
					ts.InTrade = false
					ts.TradeCount++
					ts.Mu.Unlock()
					fmt.Printf("- SELL-OFF at EntryPrice[%d]: %.8f, EntryQuant[%d]: %.8f, QBal: %.8f, BBal: %.8f, GlobalP&L %.2f LocalP&L: %.8f PosPcent: %.8f tsDataPt: %d mdDataPt: %d \n",
						len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, localProfitLoss, md.RiskPositionPercentage, ts.DataPoint, md.DataPoint)
				}
				// Print the overall trading performance after backtesting.
				fmt.Printf("Summary: Strategy: %s Combination: %s, ", md.Strategy, ts.StrategyCombLogic)
				fmt.Printf("Total Trades: %d, out of %d trials ", ts.TradeCount, len(ts.Signals))
				fmt.Printf("Total Profit/Loss: %.2f, ", md.TotalProfitLoss)
				fmt.Printf("Final Capital: %.2f, ", ts.QuoteBalance)
				fmt.Printf("Final Asset: %.8f tsDataPoint: %d, mdDataPoint: %d\n\n", ts.BaseBalance, ts.DataPoint, md.DataPoint)

				// if err := ts.APIServices.CloseDB(); err != nil{
				// 	fmt.Printf("Error while closing the DataBase: %v", err)
				// 	os.Exit(1)
				// }
				ts.ShutDownCh <- "Received termination signal. Shutting down..."
			}
		}
	}
}

// CalculateMACD calculates the Moving Average Convergence Divergence (MACD) and MACD Histogram for the given data and periods.
func CalculateMACD(SignalMACDPeriod int, closingPrices []float64, timeStamps []int64, LongPeriod, ShortPeriod int) (macdLine, signalLine, macdHistogram []float64, err error) {
	if LongPeriod <= 0 || len(closingPrices) < LongPeriod {
		err = fmt.Errorf("Error Calculating EMA: not enoguh data for period %v at total datapoint: %d", LongPeriod, len(closingPrices)-1)
		return nil, nil, nil, err
	}

	longEMA, shortEMA, _, err := CandleExponentialMovingAverage(closingPrices, timeStamps, LongPeriod, ShortPeriod)
	if err != nil {
		log.Fatalf("Error: in CalclateMACD while tring to get EMA")
	}
	// Calculate MACD line
	macdLine = make([]float64, len(closingPrices))
	for i := range closingPrices {
		macdLine[i] = shortEMA[i] - longEMA[i]
	}

	// Calculate signal line using the MACD line
	signalLine, err = CalculateExponentialMovingAverage(macdLine, SignalMACDPeriod)

	// Calculate MACD Histogram
	macdHistogram = make([]float64, len(closingPrices))
	for i := range closingPrices {
		macdHistogram[i] = macdLine[i] - signalLine[i]
	}

	return macdLine, signalLine, macdHistogram, err
}

//CandleExponentialMovingAverage calculates EMA from condles
func CandleExponentialMovingAverage(closingPrices []float64, timeStamps []int64, LongPeriod, ShortPeriod int) (longEMA, shortEMA []float64, timestamps []int64, err error) {
	if LongPeriod <= 0 || len(closingPrices) < LongPeriod {
		return nil, nil, nil, fmt.Errorf("Error Calculating Candle EMA: not enoguh data for period %v", LongPeriod)
	}
	var ema55, ema15 *ema.Ema
	ema55 = ema.NewEma(alphaFromN(LongPeriod))
	ema15 = ema.NewEma(alphaFromN(ShortPeriod))

	longEMA = make([]float64, len(closingPrices))
	shortEMA = make([]float64, len(closingPrices))
	timestamps = make([]int64, len(closingPrices))
	for k, closePrice := range closingPrices {
		ema55.Step(closePrice)
		ema15.Step(closePrice)
		longEMA[k] = ema55.Compute()
		shortEMA[k] = ema15.Compute()
		timestamps[k] = timeStamps[k]
	}
	return longEMA, shortEMA, timestamps, nil
}

// CalculateExponentialMovingAverage calculates the Exponential Moving Average (EMA) for the given data and period.
func CalculateExponentialMovingAverage(data []float64, period int) (ema []float64, err error) {
	if period <= 0 || len(data) < period {
		err = fmt.Errorf("Error Calculating EMA: not enoguh data for period %v", period)
		return nil, err
	}
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

	return ema, nil
}

// CalculateProfitLoss calculates the profit or loss from a trade.
func CalculateProfitLoss(entryPrice, exitPrice, positionSize float64) float64 {
	return (exitPrice - entryPrice) * positionSize
}

func CalculateSimpleMovingAverage(data []float64, period int) (sma []float64, err error) {
	if period <= 0 || len(data) < period {
		err = fmt.Errorf("Error Calculating SMA: not enoguh data for period %v", period)
		return nil, err
	}
	sma = make([]float64, len(data)-period+1)
	for i := 0; i <= len(data)-period; i++ {
		sum := 0.0
		for j := i; j < i+period; j++ {
			sum += data[j]
		}
		sma[i] = sum / float64(period)
	}

	return sma, nil
}

func CalculateStandardDeviation(data []float64, period int) []float64 {
	if period <= 0 || len(data) < period {
		return nil
	}

	stdDev := make([]float64, len(data)-period+1)
	for i := 0; i <= len(data)-period; i++ {
		sma, _ := CalculateSimpleMovingAverage(data[i:i+period], period)
		avg := sma[0]
		variance := 0.0
		for j := i; j < i+period; j++ {
			variance += math.Pow(data[j]-avg, 2)
		}
		stdDev[i] = math.Sqrt(variance / float64(period))
	}

	return stdDev
}

// StochasticRSI calculates the Stochastic RSI for a given RSI data series, period, and SmoothK, SmoothD.
func StochasticRSI(rsi []float64, period, smoothK, smoothD int) ([]float64, []float64, error) {
	if period <= 0 || len(rsi) < period+1 {
		return nil, nil, fmt.Errorf("Error Calculating StochRSI: not enoguh data for period %v", period)
	}
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
	emaSmoothK, err := CalculateExponentialMovingAverage(stochasticRSI, smoothK)
	if err != nil {
		return nil, nil, fmt.Errorf("Error Calculating EMA in Stochastic RSI: %v", err)
	}
	emaSmoothD, err := CalculateExponentialMovingAverage(emaSmoothK, smoothD)
	if err != nil {
		return nil, nil, fmt.Errorf("Error Calculating EMA in Stochastic RSI: %v", err)
	}

	return emaSmoothK, emaSmoothD, nil
}

// CalculateRSI calculates the Relative Strength Index for a given data series and period.
func CalculateRSI(data []float64, period int) ([]float64, error) {
	if period <= 0 || len(data) < period+1 {
		return nil, fmt.Errorf("Error Calculating RSI: not enoguh data for period %v", period)
	}
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

	return rsi, nil
}

// CalculateBollingerBands calculates the middle (SMA) and upper/lower Bollinger Bands for the given data and period.
func CalculateBollingerBands(data []float64, period int, numStdDev float64) ([]float64, []float64, []float64, error) {
	if period <= 0 || len(data) < period {
		return nil, nil, nil, fmt.Errorf("Error Calculating Bollinger Bands: not enoguh data for period %v", period)
	}
	sma, err := CalculateSimpleMovingAverage(data, period)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error Calculating SMA in Bollinger Bands: %v", err)
	}
	stdDev := CalculateStandardDeviation(data, period)

	upperBand := make([]float64, len(sma))
	lowerBand := make([]float64, len(sma))

	for i := range sma {
		upperBand[i] = sma[i] + numStdDev*stdDev[i]
		lowerBand[i] = sma[i] - numStdDev*stdDev[i]
	}

	return sma, upperBand, lowerBand, nil
}

func alphaFromN(N int) float64 {
	var n = float64(N)
	return 2. / (n + 1.)
}
