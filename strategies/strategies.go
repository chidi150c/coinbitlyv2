package strategies

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"gonum.org/v1/gonum/stat"

	"coinbitly.com/binanceapi"
	"coinbitly.com/config"
	"coinbitly.com/hitbtcapi"
	"coinbitly.com/model"
	"github.com/apourchet/investment/lib/ema"
	"github.com/pkg/errors"
)

const mainValue = 80.5

// TradingSystem struct: The TradingSystem struct represents the main trading
// system and holds various parameters and fields related to the strategy,
// trading state, and performance.
type TradingSystem struct {
	ID uint
	Mu sync.Mutex // Mutex for protecting concurrent writes
	// HistoricalData  []model.Candle
	Symbol                   string
	ClosingPrices            []float64
	Container1               []float64
	Container2               []float64
	Timestamps               []int64
	Signals                  []string
	APIServices              model.APIServices
	NextInvestBuYPrice       []float64
	NextProfitSeLLPrice      []float64
	CommissionPercentage     float64
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
	EntryCostLoss            []float64
	TradeCount               int
	TradingLevel             int
	ClosedWinTrades          int
	EnableStoploss           bool
	StopLossTrigered         bool
	StopLossRecover          []float64 //Redundant
	RiskFactor               float64
	MaxDataSize              int
	Log                      *log.Logger
	ShutDownCh               chan string
	EpochTime                time.Duration
	CSVWriter                *csv.Writer
	RiskProfitLossPercentage float64
	BaseCurrency             string //in Binance is called BaseAsset
	QuoteCurrency            string //in Binance is called QuoteAsset
	MiniQty                  float64
	MaxQty                   float64
	MinNotional              float64
	StepSize                 float64
	TargetStopLoss           float64
	TargetProfit             float64
	TotalProfitLoss          float64
	RiskPositionPercentage   float64
	ShortPeriod              int
	LongPeriod               int
	DBServices               model.DBServices
	RDBServices              *RDBServices
	DBStoreTicker            *time.Ticker
	TSDataChan               chan []byte
	ADataChan                chan []byte
	Zoom                     int
	StartTime                time.Time
	LowestPrice              float64
	HighestPrice             float64
	Index                    int
	TLevelValue              int
	TLevelAdjust             bool
	// FreeFall                 bool
	UpgdChan    chan bool
	SupIndex    int
	SupQuantity float64
	CrossUPTime time.Time
}

// NewTradingSystem(): This function initializes the TradingSystem and fetches
// historical data from the exchange using the GetCandlesFromExch() function.
// It sets various strategy parameters like short and long EMA periods, RSI period,
// MACD signal period, Bollinger Bands period, and more.
func NewTradingSystem(BaseCurrency string, liveTrading bool, loadExchFrom string) (*TradingSystem, error) {
	// Initialize the trading system.
	var (
		ts          *TradingSystem
		err         error
		rDBServices *RDBServices
	)
	loadDataFrom := ""
	rDBServices = NewRDBServices(loadExchFrom)
	if loadExchFrom == "BinanceTestnet" {
		ts, err = &TradingSystem{}, fmt.Errorf(("Testnet Error simulation"))
		if err != nil {
			fmt.Println("TS = ", ts)
			log.Printf("\n%v: But going ahead to initialize empty TS struct\n", err)
			ts = &TradingSystem{}
			ts.RiskFactor = float64(ts.Index)
			ts.CommissionPercentage = 0.00075
			ts.RiskProfitLossPercentage = 0.001
			ts.TargetProfit = mainValue * ts.RiskProfitLossPercentage
			ts.TargetStopLoss = mainValue * ts.RiskProfitLossPercentage
			ts.EnableStoploss = true
			ts.BaseCurrency = BaseCurrency
			ts.QuoteBalance = 100.0
		} else {
			loadDataFrom = "DataBase"
		}
	} else {
		ts, err = rDBServices.ReadDBTradingSystem(0)
		if err != nil {
			fmt.Println("TS = ", ts)
			log.Printf("\n%v: But going ahead to initialize empty TS struct\n", err)
			ts = &TradingSystem{}
			ts.RiskFactor = float64(ts.Index)
			ts.CommissionPercentage = 0.00075
			ts.RiskProfitLossPercentage = 0.001
			ts.TargetProfit = mainValue * ts.RiskProfitLossPercentage
			ts.TargetStopLoss = mainValue * ts.RiskProfitLossPercentage
			ts.EnableStoploss = true
			ts.BaseCurrency = BaseCurrency
		} else {
			loadDataFrom = "DataBase"
			ts.InitialCapital = 54.038193 + 26.47 + 54.2 + 86.5 + 100.0 + 16.6 + 58.0 + 56.72 + 18.0
		}
	}
	if len(ts.StopLossRecover) == 0 {
		ts.StopLossRecover = append(ts.StopLossRecover, float64(ts.ID))
	}
	fmt.Println("TS = ", ts)
	if liveTrading {
		err = ts.LiveUpdate(loadExchFrom, loadDataFrom)
		if err != nil {
			return &TradingSystem{}, err
		}
	} else {
		err = ts.UpdateHistoricalData(loadExchFrom)
		if err != nil {
			return &TradingSystem{}, err
		}
	}
	ts.RDBServices = rDBServices
	if !strings.Contains(loadExchFrom, "Testnet") {
		ts.Zoom = 100
	} else {
		ts.Zoom = 499
	}
	ts.MaxDataSize = 500
	ts.ShutDownCh = make(chan string)
	ts.EpochTime = time.Second * 60
	ts.StartTime = time.Now()
	ts.LowestPrice = math.MaxFloat64
	ts.DBStoreTicker = time.NewTicker(ts.EpochTime)
	ts.TSDataChan = make(chan []byte)
	ts.ADataChan = make(chan []byte)
	ts.UpgdChan = make(chan bool)
	//Sending TS to Frontend UI
	go func() { //Goroutine to feed the front end React UI
		log.Printf("Ready to send TS to Frontend UI")
		for {
			if len(ts.Signals) < 1 {
				time.Sleep(ts.EpochTime)
				continue
			}
			trade := model.TradingSystemData{
				Symbol:                   ts.Symbol,
				ClosingPrices:            ts.ClosingPrices,
				Timestamps:               ts.Timestamps,
				Signals:                  ts.Signals,
				CommissionPercentage:     ts.CommissionPercentage,
				InitialCapital:           ts.InitialCapital,
				PositionSize:             ts.PositionSize,
				InTrade:                  ts.InTrade,
				QuoteBalance:             ts.QuoteBalance,
				BaseBalance:              ts.BaseBalance,
				RiskCost:                 ts.RiskCost,
				DataPoint:                ts.DataPoint,
				CurrentPrice:             ts.CurrentPrice,
				TradeCount:               ts.TradeCount,
				TradingLevel:             ts.TradingLevel,
				ClosedWinTrades:          ts.ClosedWinTrades,
				EnableStoploss:           ts.EnableStoploss,
				StopLossTrigered:         ts.StopLossTrigered,
				StopLossRecover:          ts.StopLossRecover,
				RiskFactor:               ts.RiskFactor,
				MaxDataSize:              ts.MaxDataSize,
				RiskProfitLossPercentage: ts.RiskProfitLossPercentage,
				BaseCurrency:             ts.BaseCurrency,
				QuoteCurrency:            ts.QuoteCurrency,
				MiniQty:                  ts.MiniQty,
				MaxQty:                   ts.MaxQty,
				MinNotional:              ts.MinNotional,
				StepSize:                 ts.StepSize,
				TargetStopLoss:           ts.TargetStopLoss,
				TargetProfit:             ts.TargetProfit,
				TotalProfitLoss:          ts.TotalProfitLoss,
				RiskPositionPercentage:   ts.RiskPositionPercentage,
				ShortPeriod:              ts.ShortPeriod,
				LongPeriod:               ts.LongPeriod,
			}
			if len(ts.EntryPrice) > 0 {
				trade.EntryCostLoss = ts.EntryCostLoss
				trade.EntryQuantity = ts.EntryQuantity
				trade.EntryPrice = ts.EntryPrice
				trade.NextInvestBuYPrice = ts.NextInvestBuYPrice
				trade.NextProfitSeLLPrice = ts.NextProfitSeLLPrice
			}

			// Serialize the TradeSys object to JSON
			TradeSysJSON, err := json.Marshal(trade)
			if err != nil {
				log.Printf("Error marshaling TradingSystem to JSON: %v", err)
			} else {
				select {
				case ts.TSDataChan <- TradeSysJSON:
					// log.Printf("Sent TS to Frontend UI DataPoint: %d", ts.DataPoint)
				}
			}
			time.Sleep(time.Millisecond * 300)
		}
	}()
	if !strings.Contains(loadExchFrom, "Remote") {
		//Updating TS to the Database
		go func() { //Goroutine to store TS into Database
			log.Printf("Ready to send TS to Database")
			for {
				select {
				case ts.UpgdChan <- true:
					if err = ts.RDBServices.UpdateDBTradingSystem(ts); err != nil {
						panic(fmt.Sprintf("Error Upgrade Updating TradingSystem: with id: %d %v", ts.ID, err))
					}
					ts.Log.Printf("Updating TradingSystem happening now!!! where ts.ID: %d error: %v ", ts.ID, err)
					ts.UpgdChan <- true
					ts.UpgdChan <- true
				case <-time.After(time.Second * 900):
					if err = ts.RDBServices.UpdateDBTradingSystem(ts); err != nil {
						ts.ID, err = ts.RDBServices.CreateDBTradingSystem(ts)
						if err != nil {
							log.Printf("Error Creating TradingSystem: %v", err)
						}
					}
					ts.Log.Printf("Creating TradingSystem happening with Id: %d and Stage: %d where %v", ts.ID, len(ts.StopLossRecover), err)
					time.Sleep(time.Second * 10)
				}
			}
		}()
	}
	return ts, nil
}
func (ts *TradingSystem) NewDataPoint(loadExchFrom string) *model.DataPoint {
	// Initialize the App Data
	// loadDataFrom := ""
	var (
		dp  *model.DataPoint
		err error
	)
	if loadExchFrom == "BinanceTestnet" {
		dp, err = &model.DataPoint{}, fmt.Errorf(("Testnet Error simulation"))
		if err != nil {
			fmt.Println("dp = ", dp)
			log.Printf("\n%v: But going ahead to initialize empty DataPoint struct\n", err)
			dp = &model.DataPoint{}
			ts.TargetProfit = mainValue * ts.RiskProfitLossPercentage
			ts.TargetStopLoss = mainValue * ts.RiskProfitLossPercentage
			dp.ProfitLoss = 0.0
		}
	} else {
		dp = &model.DataPoint{}
		ts.TargetProfit = mainValue * ts.RiskProfitLossPercentage
		ts.TargetStopLoss = mainValue * ts.RiskProfitLossPercentage
	}
	fmt.Println("DataPoint = ", dp)
	return &model.DataPoint{}
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) UpdateHistoricalData(loadExchFrom string) error {
	var (
		err             error
		exch            model.APIServices
		exchConfigParam *config.ExchConfig
		ok              bool
		DB              model.DBServices
	)

	if exchConfigParam, ok = config.NewExchangeConfigs()[loadExchFrom]; !ok {
		fmt.Println("Error Internal: Exch name not", loadExchFrom)
		return errors.New("Exch name not " + loadExchFrom)
	}

	switch loadExchFrom {
	case "HitBTC":
		exch, err = hitbtcapi.NewAPIServices(exchConfigParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from HitBTC: %v", err)
		}
	case "Binance":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnet":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnetWithDB":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnetWithDBRemote":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	default:
		return errors.Errorf("Error updating historical data from %s invalid loadFrom tags \"%s\" ", loadExchFrom, loadExchFrom)
	}
	ts.InitialCapital = exchConfigParam.InitialCapital
	ts.DBServices = DB
	ts.APIServices = exch
	HistoricalData, err := ts.APIServices.FetchCandles(exchConfigParam.Symbol, exchConfigParam.CandleInterval, exchConfigParam.CandleStartTime, exchConfigParam.CandleEndTime)
	ts.BaseCurrency = exchConfigParam.BaseCurrency
	ts.QuoteCurrency = exchConfigParam.QuoteCurrency
	ts.Symbol = exchConfigParam.Symbol
	ts.MiniQty, ts.MaxQty, ts.StepSize, ts.MinNotional, err = exch.FetchExchangeEntities(ts.Symbol)
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
	}
	return nil
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) LiveUpdate(loadExchFrom, LoadDataFrom string) error {
	var (
		err             error
		exch            model.APIServices
		exchConfigParam *config.ExchConfig
		ok              bool
		DB              model.DBServices
	)

	if exchConfigParam, ok = config.NewExchangeConfigs()[loadExchFrom]; !ok {
		fmt.Println("Error Internal: Exch name not", loadExchFrom)
		return errors.New("Exch name not " + loadExchFrom)
	}

	switch loadExchFrom {
	case "HitBTC":
		exch, err = hitbtcapi.NewAPIServices(exchConfigParam)
		if err != nil {
			log.Fatalf("Error getting new exchange services from HitBTC: %v", err)
		}
	case "Binance":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnet":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnetWithDB":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	case "BinanceTestnetWithDBRemote":
		exch, err = binanceapi.NewAPIServices(exchConfigParam, loadExchFrom)
		if err != nil {
			log.Fatalf("Error getting new exchange services from Binance: %v", err)
		}
	default:
		return errors.Errorf("Error updating Live data from %s invalid loadFrom tags \"%s\" ", loadExchFrom, loadExchFrom)
	}
	ts.Symbol = exchConfigParam.Symbol
	ts.DBServices = DB
	ts.APIServices = exch
	ts.CurrentPrice, err = exch.FetchTicker(ts.Symbol)
	ts.MiniQty, ts.MaxQty, ts.StepSize, ts.MinNotional, err = exch.FetchExchangeEntities(ts.Symbol)
	if err != nil {
		return err
	}
	if LoadDataFrom != "DataBase" {
		ts.InitialCapital = exchConfigParam.InitialCapital
		ts.BaseCurrency = exchConfigParam.BaseCurrency
		ts.QuoteCurrency = exchConfigParam.QuoteCurrency
	}
	if loadExchFrom == "BinanceTestnet" {
		ts.QuoteBalance = 100.0
	} else if !strings.Contains(loadExchFrom, "Testnet") {
		go func() { //goroutine to get balances
			log.Printf("First Balance Update Occuring Now!!!")
			ts.QuoteBalance, ts.BaseBalance, err = ts.APIServices.GetQuoteAndBaseBalances(ts.Symbol)
			if err != nil {
				log.Fatalf("GetQuoteAndBaseBalances Error: %v", err)
			}
			log.Printf("Quote Balance for %s: %.8f\n", ts.Symbol, ts.QuoteBalance)
			log.Printf("Base Balance for %s: %.8f\n", ts.Symbol, ts.BaseBalance)
			for {
				select {
				case <-time.After(time.Minute * 30):
					log.Printf("Balance Update Occuring Now!!!")
					ts.Log.Printf("Balance Update Occuring Now!!!\n")
					ts.QuoteBalance, ts.BaseBalance, err = ts.APIServices.GetQuoteAndBaseBalances(ts.Symbol)
					if err != nil {
						log.Printf("GetQuoteAndBaseBalances Error: %v", err)
						ts.Log.Printf("GetQuoteAndBaseBalances Error: %v", err)
					}
				}
			}
		}()
	}
	// Mining data for historical analysis
	ts.DataPoint = len(ts.ClosingPrices) - 1
	return nil
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
func (ts *TradingSystem) LiveTrade(loadExchFrom string) {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl, syscall.SIGINT)
	dataPoint := ts.NewDataPoint(loadExchFrom)
	fmt.Println("App started. Press Ctrl+C to exit.")
	go ts.ShutDown(sigchnl)
	// Initialize variables for tracking trading performance.
	var err error
	log.Println("Live Trading sarting now!!!")
	for {
		ts.CurrentPrice, err = ts.APIServices.FetchTicker(ts.Symbol)
		if err != nil {
			log.Printf("Error Reporting Live Trade: %v", err)
			time.Sleep(ts.EpochTime)
			continue
		}
		ts.DataPoint++
		ts.ClosingPrices = append(ts.ClosingPrices, ts.CurrentPrice)
		ts.Timestamps = append(ts.Timestamps, time.Now().Unix())
		dataPoint.Date = time.Now()
		//Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading
		//Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading
		ts.Trading(dataPoint, loadExchFrom)

		// err = ts.DataPointtoCSV(dataPoint)
		// if err != nil {
		// 	fmt.Printf("Error creating Writing to CSV File %v", err)
		// }

		ts.TickerQueueAdjustment() //At this point you have all three(ts.ClosingPrices, ts.Timestamps and ts.Signals) assigned

		err = ts.Reporting("Live Trading")
		if err != nil {
			fmt.Println("Error Reporting Live Trade: ", err)
			return
		}
		if (len(ts.EntryPrice) > 0) && ((ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1] > ts.CurrentPrice) || (ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1] < ts.CurrentPrice)) {
			if !ts.InTrade {
				ts.LowestPrice = math.MaxFloat64
				ts.StartTime = time.Now()
			} else if ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1] < ts.CurrentPrice {
				ts.StartTime = time.Now()
				ts.HighestPrice = 0.0
			}
		}
		
		if (len(ts.EntryPrice) > 0) && (ts.LowestPrice > ts.CurrentPrice) {
			ts.LowestPrice = ts.CurrentPrice
		}
		if (len(ts.EntryPrice) > 0) && (ts.HighestPrice < ts.CurrentPrice) {
			ts.HighestPrice = ts.CurrentPrice
		}
		if ts.InTrade && (len(ts.EntryPrice) > 0) {
			//NextSell Re-Adjustment
			nextProfitSeLLPrice := ((ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / ts.EntryQuantity[len(ts.EntryQuantity)-1]) + ts.EntryPrice[len(ts.EntryPrice)-1]
			commissionAtProfitSeLLPrice := nextProfitSeLLPrice * ts.EntryQuantity[len(ts.EntryQuantity)-1] * ts.CommissionPercentage
			if ((nextProfitSeLLPrice + commissionAtProfitSeLLPrice) < ts.HighestPrice) && (time.Since(ts.StartTime) > elapseTimeSeLL(ts.TradingLevel)) {
				before := ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1]
				ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1] = ts.HighestPrice
				ts.Log.Printf("NextProfitSeLLPrice Re-adjusted!!! from Before: %.8f to Now: %.8f", before, ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1])
				ts.StartTime = time.Now()
				ts.HighestPrice = 0.0
			}
		} else if len(ts.EntryPrice) > 0 {
			//NextBuy Re-Adjustment
			nextInvBuYPrice := (-(ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / ts.EntryQuantity[len(ts.EntryQuantity)-1]) + ts.EntryPrice[len(ts.EntryPrice)-1]
			if time.Since(ts.StartTime) > elapseTime(ts.TradingLevel) {
				if len(ts.EntryPrice) < 5 {
					before := ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1]
					ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1] = ts.LowestPrice
					ts.Log.Printf("NextInvestBuYPrice Re-adjusted!!! from Before: %.8f to Now: %.8f", before, ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1])
					ts.StartTime = time.Now()
					ts.LowestPrice = math.MaxFloat64
				} else if (nextInvBuYPrice) > ts.LowestPrice {
					before := ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1]
					ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1] = ts.LowestPrice
					ts.Log.Printf("NextInvestBuYPrice Re-adjusted!!! from Before: %.8f to Now: %.8f", before, ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1])
					ts.StartTime = time.Now()
					ts.LowestPrice = math.MaxFloat64
				}
			}
		}
		if !ts.InTrade {
			ts.RiskPositionPercentage = ts.LowestPrice // Define risk management parameter 5% balance
		} else {
			ts.RiskPositionPercentage = ts.HighestPrice // Define risk management parameter 5% balance
		}
		if time.Since(dataPoint.Date) > time.Minute{
			ts.Log.Printf("Time inconsistency as interval %.2fsecs is significantly larger than 60secs", time.Since(dataPoint.Date).Seconds())
		}
		time.Sleep(ts.EpochTime-(time.Since(dataPoint.Date)))
	}
}
func deleteElement(slice []float64, index int) []float64 {
	// Check if the index is valid
	if index < 0 || index >= len(slice) {
		fmt.Println("Index out of range")
		return slice
	}

	// Create a new slice that excludes the element at the specified index
	return append(slice[:index], slice[index+1:]...)
}
func elapseTime(level int) time.Duration {
	switch level {
	case 0:
		return time.Minute * 20
	case 1:
		return time.Minute * 30
	case 2:
		return time.Minute * 40
	case 3:
		return time.Minute * 60
	case 4:
		return time.Minute * 60 * 2
	case 5:
		return time.Minute * 60 * 3
	case 6:
		return time.Minute * 60 * 4
	default:
		return time.Minute * 60 * 5
	}
}
func elapseTimeSeLL(level int) time.Duration {
	return time.Minute * 20
	// switch level {
	// case 0:
	// 	return time.Minute * 20
	// case 1:
	// 	return time.Minute * 25
	// case 2:
	// 	return time.Minute * 30
	// case 3:
	// 	return time.Minute * 38
	// case 4:
	// 	return time.Minute * 46
	// case 5:
	// 	return time.Minute * 52
	// case 6:
	// 	return time.Minute * 58
	// case 7:
	// 	return time.Minute * 64
	// case 8:
	// 	return time.Minute * 70
	// case 9:
	// 	return time.Minute * 76
	// default:
	// 	return time.Minute * 60 * 2
	// }
}

// Backtest(): This function simulates the backtesting process using historical
// price data. It iterates through the closing prices, checks for entry and exit
// conditions, and executes trades accordingly. It also tracks trading performance
// and updates the current balance after each trade.
func (ts *TradingSystem) Backtest(loadExchFrom string) {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)

	fmt.Println("App started. Press Ctrl+C to exit.")
	var err error
	dataPoint := ts.NewDataPoint(loadExchFrom)
	// Simulate the backtesting process using historical price data.
	for ts.DataPoint, ts.CurrentPrice = range ts.ClosingPrices {
		select {
		case <-sigchnl:
			fmt.Println("ShutDown Requested: Press Ctrl+C again to shutdown...")
			sigchnl := make(chan os.Signal, 1)
			signal.Notify(sigchnl)
			ts.ShutDown(sigchnl)
		default:
		}

		ts.Trading(dataPoint, loadExchFrom)

		// err = ts.DataPointtoCSV(dataPoint)
		// if err != nil {
		// 	fmt.Printf("Error creating Writing to CSV File %v", err)
		// }
		// time.Sleep(ts.EpochTime)
	}
	err = ts.Reporting("Backtesting")
	if err != nil {
		fmt.Println("Error Reporting BackTest Trading: ", err)
		return
	}

	fmt.Println("Press Ctrl+C to continue...")
	<-sigchnl
	fmt.Println("End of Backtesting: Press Ctrl+C again to shutdown...")
	sigchnl = make(chan os.Signal, 1)
	signal.Notify(sigchnl)
	ts.ShutDown(sigchnl)

	if err != nil {
		fmt.Println("Backtest: ", err)
	}

}

func (ts *TradingSystem) Trading(dp *model.DataPoint, loadExchFrom string) {
	// Execute the trade if entry conditions are met.
	v := 0.0
	passed := false
	targetCrossed := false
	if len(ts.NextInvestBuYPrice) > 0 {
		if ts.CurrentPrice <= ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1] {
			targetCrossed = true
		}
	} else {
		targetCrossed = true
	}

	///////////////////////////////////////////////////////////////
	strategy := ts.StrategyProcessing(dp)
	////////////////////////////////////////////////////////////////

	if (strategy == "buy") && targetCrossed {
		// Execute the buy order using the ExecuteStrategy function.
		resp, err := ts.ExecuteStrategy(dp, "Buy")
		if err != nil {
			// fmt.Println("Error executing buy order:", err)
			dp.Label2 = 0
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
		} else if strings.Contains(resp, "BUY") {
			// Record Signal for plotting graph later
			dp.Label2 = 1
			ts.Signals = append(ts.Signals, "Buy")
			ts.Log.Println(resp)
			ts.TradeCount++
		} else {
			dp.Label2 = 0
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Positio
		}
		// Close the trade if exit conditions are met.
		passed = true
	} 
	targetCrossed = false
	for ts.Index, v = range ts.NextProfitSeLLPrice {
		if ts.CurrentPrice > v {
			targetCrossed = true
			break
		}
	}
	ts.RiskFactor = float64(ts.Index)
	if (strategy == "sell") && targetCrossed{
		// Execute the sell order using the ExecuteStrategy function.
		resp, err := ts.ExecuteStrategy(dp, "Sell")
		if err != nil {
			dp.Label2 = 0
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
		} else if strings.Contains(resp, "SELL") {
			// Record Signal for plotting graph later.
			dp.Label2 = -1
			ts.Signals = append(ts.Signals, "Sell")
			ts.Log.Println(resp)
			ts.TradeCount++
		} else {
			dp.Label2 = 0
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
		}
		passed = true
	}
	if (strategy == "hold") || (!passed){                   //To takecare of EMA Calculation and Grphing
		dp.Label2 = 0
		ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
	}
	dp.ProfitLoss = ts.TotalProfitLoss
	dp.ProfitLoss = ts.TotalProfitLoss
	dp.CurrentPrice = ts.CurrentPrice
}

// ExecuteStrategy executes the trade based on the provided trade action and current price.
// The tradeAction parameter should be either "Buy" or "Sell".
func (ts *TradingSystem) ExecuteStrategy(dp *model.DataPoint, tradeAction string) (string, error) {
	var (
		orderResp model.Response
		err       error
	)
	switch tradeAction {
	case "Buy":
		ts.RiskManagement(dp)
		// adjustedPrice := math.Floor(price/lotSizeStep) * lotSizeStep
		quantity := math.Floor(ts.PositionSize/ts.MiniQty) * ts.MiniQty
		// Calculate the total cost of the trade
		totalCost := quantity * ts.CurrentPrice
		//check if there is enough Quote (USDT) for the buy transaction
		if ts.QuoteBalance < totalCost {
			//if we have hit bottom earlier and expecting to sell but rather hit buy target
			ts.Log.Printf("Unable to buY!!!! Insufficient QuoteBalance: %.8f (ts.PositionSize %.8f* ts.CurrentPrice %.8f) = %.8f", ts.QuoteBalance, ts.PositionSize, ts.CurrentPrice, ts.PositionSize*ts.CurrentPrice)
			return "", fmt.Errorf("cannot execute a buy order due to insufficient QuoteBalance: %.8f (ts.PositionSize %.8f* ts.CurrentPrice %.8f) = %.8f", ts.QuoteBalance, ts.PositionSize, ts.CurrentPrice, ts.PositionSize*ts.CurrentPrice)
			// }
		}
		// Check if the total cost meets the minNotional requirement
		if totalCost < ts.MinNotional {
			ts.Log.Printf("Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, ts.CurrentPrice, totalCost, ts.MinNotional)
			return "", fmt.Errorf("Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, ts.CurrentPrice, totalCost, ts.MinNotional)
		}

		//Placing a Buy order
		//////////////////////////////////////////////////////////////////
		//////////////////////////////////////////////////////////////////
		orderResp, err = ts.APIServices.PlaceLimitOrder(ts.Symbol, "BUY", ts.CurrentPrice, quantity)
		if err != nil {
			fmt.Println("Error placing entry order:", err)
			ts.Log.Println("Error placing entry order:", err)
			return "", fmt.Errorf("Error placing entry order: %v", err)
		}
		if (orderResp.ExecutedQty != quantity) && (orderResp.ExecutedQty > 0.0) {
			quantity = orderResp.ExecutedQty
		}
		//the commission from Binance being in BaseCurrency was deducted from the executedQty
		totalCost = (quantity * ts.CurrentPrice)
		//Commission: 0.00000000 CumulativeQuoteQty: 0.00000000 ExecutedPrice: 29840.00000000 ExecutedQty: 0.00000000 Status:
		ts.Log.Printf("Entry order placed with ID: %d Commission: %.8f CumulativeQuoteQty: %.8f ExecutedPrice: %.8f ExecutedQty: %.8f Status: %s\n", orderResp.OrderID, orderResp.Commission, orderResp.CumulativeQuoteQty,
			orderResp.ExecutedPrice, quantity, orderResp.Status)
		// Update the profit, quote and base balances after the trade.
		ts.QuoteBalance -= totalCost
		ts.BaseBalance += quantity
		ts.TotalProfitLoss -= (ts.CommissionPercentage * quantity * ts.CurrentPrice)

		//Record entry entities for calculating profit/loss and stoploss later.
		ts.EntryPrice = append(ts.EntryPrice, ts.CurrentPrice)
		ts.EntryQuantity = append(ts.EntryQuantity, quantity)
		ts.EntryCostLoss = append(ts.EntryCostLoss, (ts.CommissionPercentage * quantity * ts.CurrentPrice))
		nextProfitSeLLPrice := ((ts.TargetProfit + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / quantity) + ts.EntryPrice[len(ts.EntryPrice)-1]
		nextInvBuYPrice := (-(ts.TargetStopLoss + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / quantity) + ts.EntryPrice[len(ts.EntryPrice)-1]
		commissionAtProfitSeLLPrice := nextProfitSeLLPrice * quantity * ts.CommissionPercentage
		commissionAtInvBuYPrice := nextInvBuYPrice * quantity * ts.CommissionPercentage
		ts.NextProfitSeLLPrice = append(ts.NextProfitSeLLPrice, nextProfitSeLLPrice+commissionAtProfitSeLLPrice)
		if ts.TLevelAdjust {
			ts.Log.Printf("Replenished bursted:[%d] !!! \n", ts.TLevelValue)
			ts.TLevelAdjust = false
		}
		ts.NextInvestBuYPrice = append(ts.NextInvestBuYPrice, nextInvBuYPrice-commissionAtInvBuYPrice)
		ts.TradingLevel = len(ts.EntryPrice)
		ts.LowestPrice = math.MaxFloat64
		ts.StartTime = time.Now()
		resp := fmt.Sprintf("- BUY at EntryPrice[%d]: %.8f, EntryQuantity[%d]: %.8f, QBal: %.8f, BBal: %.8f, EntryCostLoss[%d]: %.8f PosPcent: %.8f, \nGlobalP&L %.2f, nextProfitSeLLPrice[%d]: %.8f nextInvBuYPrice[%d]: %.8f TargProfit %.8f TargLoss %.8f, tsDataPt: %d\n",
			len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, len(ts.EntryCostLoss)-1, ts.EntryCostLoss[len(ts.EntryCostLoss)-1], ts.RiskPositionPercentage, ts.TotalProfitLoss, len(ts.NextProfitSeLLPrice)-1, ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1], len(ts.NextInvestBuYPrice)-1, ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1], ts.TargetProfit, -ts.TargetStopLoss, ts.DataPoint)

		<-ts.UpgdChan
		<-ts.UpgdChan
		if ((!ts.InTrade) && (ts.StopLossTrigered)) || (ts.TradingLevel >= 5) {
			//upgrade to next stage
			ts.ID, err = ts.RDBServices.CreateDBTradingSystem(ts)
			<-ts.UpgdChan
			if err != nil {
				panic(fmt.Sprintf("Error Creating TradingSystem: %v", err))
			} else {
				log.Printf("Upgrade done with New Ts ID: %d\n", ts.ID)
				ts.EntryPrice = []float64{}
				ts.EntryQuantity = []float64{}
				ts.EntryCostLoss = []float64{}
				ts.NextProfitSeLLPrice = []float64{}
				ts.NextInvestBuYPrice = []float64{}
				ts.StopLossRecover = append(ts.StopLossRecover, float64(ts.ID))
				ts.TradingLevel = len(ts.EntryPrice)
				ts.LowestPrice = math.MaxFloat64
				ts.StartTime = time.Now()
			}
		} else {
			<-ts.UpgdChan
		}

		ts.RiskManagement(dp)
		// adjustedPrice := math.Floor(price/lotSizeStep) * lotSizeStep
		quantity = math.Floor(ts.PositionSize/ts.MiniQty) * ts.MiniQty
		// Calculate the total cost of the trade
		totalCost = quantity * ts.CurrentPrice
		i := len(ts.NextInvestBuYPrice)
		before := 0.0
		if i > 1 {
			before = ts.NextInvestBuYPrice[i-2]
		}
		//Check whether to move to next stage due to insufficiency in next buy
		if ts.QuoteBalance < totalCost { //level
			if (!ts.InTrade) && (ts.StopLossTrigered) {
				ts.InTrade = false
				ts.StopLossTrigered = false
			} else {
				ts.InTrade = true
				ts.StopLossTrigered = true
				ts.HighestPrice = ts.CurrentPrice
				//so that is goes down the full next buy original target without adjustment
				for k, _ := range ts.NextInvestBuYPrice {
					if k >= 1 {
						ts.NextInvestBuYPrice[k-1] = ts.NextInvestBuYPrice[i-1]
					}
				}
				ts.Log.Printf("NextSell Re-Adjusting Switched ON for %s: as Balance = %.8f is Less than NextRiskCost = %.8f and NextInvestBuYPrice[%d] updated with [%d] from %.8f to %.8f \n", ts.Symbol, ts.QuoteBalance, totalCost, i-2, i-1, before, ts.NextInvestBuYPrice[i-1])
				resp = fmt.Sprintf("- BUY at EntryPrice[%d]: %.8f, EntryQuantity[%d]: %.8f, QBal: %.8f, BBal: %.8f, EntryCostLoss[%d]: %.8f PosPcent: %.8f, \nGlobalP&L %.2f, nextProfitSeLLPrice[%d]: %.8f nextInvBuYPrice[%d]: %.8f TargProfit %.8f TargLoss %.8f, tsDataPt: %d, mdDataPt: %d \n",
					len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, len(ts.EntryCostLoss)-1, ts.EntryCostLoss[len(ts.EntryCostLoss)-1], ts.RiskPositionPercentage, ts.TotalProfitLoss, len(ts.NextProfitSeLLPrice)-1, ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1], len(ts.NextInvestBuYPrice)-1, ts.NextInvestBuYPrice[len(ts.NextInvestBuYPrice)-1], ts.TargetProfit, -ts.TargetStopLoss, ts.DataPoint, ts.DataPoint)
			}
		} else {
			if !ts.InTrade {
				ts.StopLossTrigered = false
			}
			ts.HighestPrice = 0.0
			ts.InTrade = false
		}
		return resp, nil
	case "Sell":
		ts.Log.Printf("Trying to SeLL now, currentPrice: %.8f, Target Profit: %.8f", ts.CurrentPrice, ts.TargetProfit)
		suplemented := false
		floatHd := 0.0
		Asset := (ts.BaseBalance * ts.CurrentPrice) + ts.QuoteBalance
		qpcent := (ts.QuoteBalance / Asset) * 100.0
		quantity := ts.EntryQuantity[ts.Index]
		//Deciding whether to execute a supplemental sell if quote percentage falls below the 20% threshold.
		if (len(ts.EntryPrice) >= 2) && (ts.InTrade && ts.StopLossTrigered) {
			localProfitLoss := CalculateProfitLoss(ts.EntryPrice[ts.Index], ts.CurrentPrice, quantity)
			v := 0.0
			ts.Log.Printf("Asset Calculated: %.8f QuotePercentage: %.8f Index [%d] Pre-LocalProfitLoss %.8f", Asset, qpcent, ts.Index, localProfitLoss)
			for ts.SupIndex, v = range ts.EntryQuantity {
				if ts.SupIndex != ts.Index {
					ts.SupQuantity = CalculateQuantity(ts.EntryPrice[ts.SupIndex], ts.CurrentPrice, -localProfitLoss*Qfactor(qpcent))
					ts.Log.Printf("SupIndex[%d] Calculated SupQuantity: %.8f Position Quantity: %.8f", ts.SupIndex, ts.SupQuantity, v)
					if v > ts.SupQuantity {
						qpcent = quantity + ts.SupQuantity
						floatHd = math.Floor(qpcent/ts.MiniQty) * ts.MiniQty
						if ts.SupQuantity > (qpcent - floatHd) {
							ts.SupQuantity -= qpcent - floatHd
							quantity += ts.SupQuantity
							suplemented = true
							ts.Log.Printf("Finally Suplemented SupIndex[%d] SupQuantity: %.8f ", ts.SupIndex, ts.SupQuantity)
							break
						}
					}
				}
			}
		}

		quantity = math.Floor(quantity/ts.MiniQty) * ts.MiniQty
		if ts.BaseBalance < quantity {
			ts.Log.Printf("But BaseBalance %.8f is < quantity %.8f", ts.BaseBalance, quantity)
			quantity = math.Floor(ts.BaseBalance/ts.MiniQty) * ts.MiniQty
			if quantity < ts.MiniQty {
				//Delete or Reset entry
				ts.DeleteOrResetEntry("floor BaseBalance", quantity, "MiniQty", ts.MiniQty)
				return "", fmt.Errorf("cannot execute a sell order due to insufficient BaseBalance: %.8f miniQuantity required: %.8f", ts.QuoteBalance, ts.MiniQty)
			} else {
				// return "", fmt.Errorf("cannot execute a sell order insufficient BaseBalance: %.8f needed up to: %.8f", ts.BaseBalance, quantity)
			}
		}
		// Calculate the total cost of the trade
		totalCost := quantity * ts.CurrentPrice
		// Check if the total cost meets the minNotional requirement
		if totalCost < ts.MinNotional {
			reqQuantity := ts.MinNotional / ts.CurrentPrice
			if len(ts.EntryPrice) > 1 {
				agg := ts.AggregateEntries()
				ts.Log.Printf(agg)
				return "", fmt.Errorf(agg)
			} else if ts.BaseBalance >= (reqQuantity + ts.MiniQty) {
				d := quantity
				quantity = math.Floor((reqQuantity+ts.MiniQty)/ts.MiniQty) * ts.MiniQty
				ts.Log.Printf("Suplemented Quantity Defficiency!!! from %.8f to %.8f", d, quantity)
			} else {
				//Delete or Reset entry
				ts.DeleteOrResetEntry("totalCost", totalCost, "MinNotional", ts.MinNotional)
				ts.Log.Printf("Less than MinNotional: Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, ts.CurrentPrice, totalCost, ts.MinNotional)

				return "", fmt.Errorf("Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, ts.CurrentPrice, totalCost, ts.MinNotional)
			}
		}

		//Placing Sell Order
		/////////////////////////////////////////////////////////////////////////////////////////////////////////////
		//////////////////////////////////////////////////////////////////////////////////////////////////
		orderResp, err := ts.APIServices.PlaceLimitOrder(ts.Symbol, "SELL", ts.CurrentPrice, quantity)
		if err != nil {
			fmt.Println("Error placing exit order:", err)
			ts.Log.Println("Error placing exit order:", err)
			return "", fmt.Errorf("Error placing exit order: %v", err)
		}
		if orderResp.ExecutedPrice != ts.CurrentPrice && (orderResp.ExecutedPrice > 0.0) {
			ts.CurrentPrice = orderResp.ExecutedPrice
		}
		if (orderResp.ExecutedQty != quantity) && (orderResp.ExecutedQty > 0.0) {
			quantity = orderResp.ExecutedQty
		}
		//For a sell order:
		//You subtract the commission from the total cost because you are receiving less due to the commission fee. So, the formula would
		//be something like: totalCost = (executedQty * ts.CurrentPrice) - totalCommission.
		//Confirm Binance or the Exch is deliverying commission for Sell in USDT
		totalCost = (quantity * ts.CurrentPrice)

		ts.Log.Printf("Exit order placed with ID: %d Commission: %.8f CumulativeQuoteQty: %.8f ExecutedPrice: %.8f ExecutedQty: %.8f Status: %s\n", orderResp.OrderID, orderResp.Commission, orderResp.CumulativeQuoteQty,
			orderResp.ExecutedPrice, quantity, orderResp.Status)
		// Update the totalP&L, quote and base balances after the trade.
		ts.QuoteBalance += totalCost
		ts.BaseBalance -= quantity
		localProfitLoss := 0.0
		if suplemented == true {
			pl := CalculateProfitLoss(ts.EntryPrice[ts.SupIndex], ts.CurrentPrice, ts.SupQuantity)
			quantity -= ts.SupQuantity
			ts.EntryQuantity[ts.SupIndex] -= ts.SupQuantity
			localProfitLoss = CalculateProfitLoss(ts.EntryPrice[ts.Index], ts.CurrentPrice, quantity) + pl
			ts.Log.Printf("STOPLOST!!! Suplemented with Entry [%d] for Quantity: %.8f to Remain %.8f for Asset Balance ratio 40:60", ts.SupIndex, ts.SupQuantity, ts.EntryQuantity[ts.SupIndex])
		} else {
			localProfitLoss = CalculateProfitLoss(ts.EntryPrice[ts.Index], ts.CurrentPrice, quantity)
		}
		// ts.Log.Printf("Profit Before Global: %v, Local: %v\n",ts.TotalProfitLoss, localProfitLoss)
		ts.TotalProfitLoss += localProfitLoss
		dp.ProfitLoss = ts.TotalProfitLoss
		// ts.Log.Printf("Profit After Global: %v, Local: %v\n",ts.TotalProfitLoss, localProfitLoss)
		if localProfitLoss >= 0 {
			ts.ClosedWinTrades += 2
		}
		if !ts.InTrade {
			ts.StopLossTrigered = false
		}
		ts.InTrade = false
		ts.StartTime = time.Now()
		ts.LowestPrice = math.MaxFloat64
		ts.HighestPrice = 0.0
		resp := fmt.Sprintf("- SELL at ts.CurrentPrice: %.8f, EntryPrice[%d]: %.8f, EntryQuantity[%d]: %.8f, SupposedIndex: [%d], QBal: %.8f, BBal: %.8f, \nGlobalP&L: %.2f localP&L: %.2f SellCommission: %.8f PosPcent: %.8f tsDataPt: %d \n",
			ts.CurrentPrice, ts.Index, ts.EntryPrice[ts.Index], ts.Index, ts.EntryQuantity[ts.Index], len(ts.EntryPrice)-1, ts.QuoteBalance, ts.BaseBalance, ts.TotalProfitLoss, localProfitLoss, orderResp.Commission, ts.RiskPositionPercentage, ts.DataPoint)
		if (len(ts.EntryPrice) - 1) > ts.Index {
			ts.Log.Printf("Bursted:[%d] !!! at [%d]\n", ts.Index, len(ts.EntryPrice)-1)
			ts.TLevelValue = ts.Index
			ts.TLevelAdjust = true
		}
		ts.EntryPrice = deleteElement(ts.EntryPrice, ts.Index)
		ts.EntryCostLoss = deleteElement(ts.EntryCostLoss, ts.Index)
		ts.EntryQuantity = deleteElement(ts.EntryQuantity, ts.Index)
		ts.NextProfitSeLLPrice = deleteElement(ts.NextProfitSeLLPrice, ts.Index)
		ts.NextInvestBuYPrice = deleteElement(ts.NextInvestBuYPrice, ts.Index)
		ts.TradingLevel = len(ts.EntryPrice)

		<-ts.UpgdChan
		<-ts.UpgdChan
		if len(ts.EntryPrice) == 0 && len(ts.StopLossRecover) > 1 {
			log.Printf("Down Staging: Deleting Current TS with ID: %d", ts.ID)
			ts.Log.Printf("Down Staging: Deleting Current TS with ID: %d", ts.ID)
			err = ts.RDBServices.DeleteDBTradingSystem(ts.ID)
			if err != nil {
				panic(fmt.Sprintf("Error Deleting TS during down staging ts.ID: %d", ts.ID))
			}
			tsn, err := ts.RDBServices.ReadDBTradingSystem(uint(ts.StopLossRecover[len(ts.StopLossRecover)-2]))
			if err != nil {
				panic(fmt.Sprintf("Error Reading next TS during down staging ts.ID: %d", ts.ID))
			}
			log.Printf("Upgrade done with New Ts ID: %d\n", ts.ID)
			ts.ID = tsn.ID
			ts.EntryPrice = tsn.EntryPrice
			ts.EntryCostLoss = tsn.EntryCostLoss
			ts.EntryQuantity = tsn.EntryQuantity
			ts.NextProfitSeLLPrice = tsn.NextProfitSeLLPrice
			ts.NextInvestBuYPrice = tsn.NextInvestBuYPrice
			ts.StopLossRecover = tsn.StopLossRecover
			ts.StopLossTrigered = tsn.StopLossTrigered
			ts.InTrade = tsn.InTrade
			<-ts.UpgdChan
			log.Printf("Down Staged Successfully!!! to ID: %d\n", ts.ID)
			ts.Log.Printf("Down Staged Successfully!!! to ID: %d\n", ts.ID)
		} else {
			<-ts.UpgdChan
		}
		return resp, nil
	default:
		return "", fmt.Errorf("invalid trade action: %s", tradeAction)
	}
}
func Qfactor(qpcent float64) float64 {
	if qpcent > 20.0 {
		return 2.0 / 6.0
	} else if qpcent > 15.0 {
		return 3.0 / 6.0
	} else if qpcent > 10.0 {
		return 4.0 / 6.0
	} else if qpcent > 5.0 {
		return 5.0 / 6.0
	} else {
		return 1.0
	}
}
func swapElements(slice []float64, index1, index2 int) []float64 {
	// Check if the indices are valid
	if index1 < 0 || index1 >= len(slice) || index2 < 0 || index2 >= len(slice) {
		fmt.Println("Invalid indices")
		return slice
	}

	// Swap the values at the specified indices
	slice[index1], slice[index2] = slice[index2], slice[index1]
	return slice
}
func (ts *TradingSystem) DeleteOrResetEntry(have string, quantity float64, expected string, target float64) {
	ts.Log.Printf("So Deleting and Resetting entry as %s %.8f is < %s %.8f", have, quantity, expected, target)
	if !ts.InTrade {
		ts.StopLossTrigered = false
	}
	ts.InTrade = false
	ts.StartTime = time.Now()
	ts.LowestPrice = math.MaxFloat64
	ts.HighestPrice = 0.0
	ts.EntryPrice = deleteElement(ts.EntryPrice, ts.Index)
	ts.EntryCostLoss = deleteElement(ts.EntryCostLoss, ts.Index)
	ts.EntryQuantity = deleteElement(ts.EntryQuantity, ts.Index)
	ts.NextProfitSeLLPrice = deleteElement(ts.NextProfitSeLLPrice, ts.Index)
	ts.NextInvestBuYPrice = deleteElement(ts.NextInvestBuYPrice, ts.Index)
	ts.TradingLevel = len(ts.EntryPrice)
}
func (ts *TradingSystem) AggregateEntries() string {
	AggResult := fmt.Sprintf("Aggregate Required !!! for [%d] of %.8f quantity", ts.Index, ts.EntryQuantity[ts.Index])
	if len(ts.EntryPrice) > 1 { //ts.Index
		for k, _ := range ts.EntryPrice {
			if k != ts.Index {
				ts.EntryQuantity[k] += ts.EntryQuantity[ts.Index]
				AggResult = fmt.Sprintf("Aggregated [%d] of X and [%d] of %.8f to be %.8f ", k, ts.Index, ts.EntryQuantity[ts.Index], ts.EntryQuantity[k])
				if !ts.InTrade {
					ts.StopLossTrigered = false
				}
				ts.InTrade = false
				ts.StartTime = time.Now()
				ts.LowestPrice = math.MaxFloat64
				ts.HighestPrice = 0.0
				ts.EntryPrice = deleteElement(ts.EntryPrice, ts.Index)
				ts.EntryCostLoss = deleteElement(ts.EntryCostLoss, ts.Index)
				ts.EntryQuantity = deleteElement(ts.EntryQuantity, ts.Index)
				ts.NextProfitSeLLPrice = deleteElement(ts.NextProfitSeLLPrice, ts.Index)
				ts.NextInvestBuYPrice = deleteElement(ts.NextInvestBuYPrice, ts.Index)
				ts.TradingLevel = len(ts.EntryPrice)
				break
			}
		}
	}
	return AggResult
}

// RiskManagement applies risk management rules to limit potential losses.
// It calculates the stop-loss price based on the fixed percentage of risk per trade and the position size.
// If the current price breaches the stop-loss level, it triggers a sell signal and exits the trade.
func (ts *TradingSystem) RiskManagement(dp *model.DataPoint) {
	// Calculate position size based on the fixed percentage of risk per trade.
	Asset := (ts.BaseBalance * ts.CurrentPrice) + ts.QuoteBalance
	num := (ts.MinNotional + 1.0) / ts.StepSize
	ts.RiskCost = math.Floor(num) * ts.StepSize
	if ts.InitialCapital < Asset {
		diff := Asset - ts.InitialCapital
		num += diff
	}
	ts.RiskCost = math.Floor(num) * ts.StepSize
	if !ts.TLevelAdjust {
		ts.TLevelValue = ts.TradingLevel
	}
	if (!ts.InTrade) && (ts.StopLossTrigered) {
		ts.TLevelValue = 0
	}
	switch ts.TLevelValue {
	// case 0:
	// 	ts.RiskCost += 35.0 + 10.0 + 5.0
	// 	ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	// 	ts.TargetProfit = mainValue * 0.00095
	// 	ts.TargetStopLoss = mainValue * 0.003
	case 0:
		ts.RiskCost += 40.0 + 12.5 + 7.5
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = mainValue * 0.001
		ts.TargetStopLoss = mainValue * 0.0035
	case 1:
		ts.RiskCost += 45.0 + 15. + 10.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = mainValue * 0.0015
		ts.TargetStopLoss = mainValue * 0.004
	case 2:
		ts.RiskCost += 50.0 + 17.5 + 12.5
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = mainValue * 0.002
		ts.TargetStopLoss = mainValue * 0.0045
	case 3:
		ts.RiskCost += 55.0 + 19.5 + 15.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = mainValue * 0.0025
		ts.TargetStopLoss = mainValue * 0.005
	case 4:
		ts.RiskCost += (60.0 + 22.0 + 17.5) * 2.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = (mainValue * 0.003) * 2.0
		ts.TargetStopLoss = mainValue * 0.0055
	case 5:
		ts.RiskCost += 65.0 + 24.5 + 19.5
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = mainValue * 0.0035
		ts.TargetStopLoss = mainValue * 0.006
	case 6:
		ts.RiskCost += 70.0 + 27.0 + 22.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = mainValue * 0.004
		ts.TargetStopLoss = mainValue * 0.0065
	default:
		ts.RiskCost += 75.0 + 29.5 + 24.5
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
		ts.TargetProfit = mainValue * 0.0045
		ts.TargetStopLoss = mainValue * 0.007
	}
}
func (ts *TradingSystem) AIAnalysis(dp *model.DataPoint) (buySignal, sellSignal bool) {
	// Create an instance of DataPointRequest and populate it with data
	if strings.Contains(ts.RDBServices.loadExchFrom, "TestnetWithOutAI"){
		ts.Log.Printf("ATAnalysis reached with: %v", dp)
		return true, true
	}
	// Convert dataPoint to JSON
	jsonData, err := json.Marshal(dp)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}

	// Send POST request
	var response *http.Response
	if strings.Contains(ts.RDBServices.loadExchFrom, "Testnet") {
		response, err = http.Post("http://localhost:5000/predict", "application/json", bytes.NewBuffer(jsonData))
	} else {
		response, err = http.Post("http://flask-app:5000/predict", "application/json", bytes.NewBuffer(jsonData))
	}
	if err != nil {
		ts.Log.Printf("Error occurred while sending POST request to AI Model!!!. Error: %s", err.Error())
		return true, true
	}
	defer response.Body.Close()
	// Read the response body
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ts.Log.Printf("Error occurred while reading the response body of AI Model!!!. Error: %s", err.Error())
		return true, true
	}

	
	// // Log raw response data for debugging
	// ts.Log.Printf("Raw response data: %s", string(responseData))

	// Unmarshal the response data into PredictionResponse struct
	var prediction model.PredictionResponse
	err = json.Unmarshal(responseData, &prediction)
	if err != nil {
		ts.Log.Printf("Error occurred during unmarshaling of AI Model!!!. Error: %s", err.Error())
		return true, true
	}

	// Use the prediction as needed
	ts.Log.Printf("Received prediction: %d ts.DataPoint %d ts.CurrentPrice %.8f", prediction.Prediction, ts.DataPoint, ts.CurrentPrice)

	if prediction.Prediction == -1 {
		dp.Label2 = -1
		return false, true
	} else if prediction.Prediction == 1 {
		dp.Label2 = 1
		return true, false
	}
	dp.Label2 = 0
	return false, false
}

// TechnicalAnalysis(): This function performs technical analysis using the
// calculated moving averages (short and long EMA), RSI, MACD line, and Bollinger
// Bands. It determines the buy and sell signals based on various strategy rules.
func (ts *TradingSystem) TechnicalAnalysis(dataPoint *model.DataPoint) (buySignal, sellSignal bool) {
	var HD []model.Candle
	ch0 := make(chan bool)
	go func (ch0 chan bool)  {
		var err error
		CandleStartTime := time.Now().Add(time.Duration(-1) * 20 * time.Minute).Unix()
		CandleEndTime :=   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix()
		CandleInterval :=  "1m"
		HD, err = ts.APIServices.FetchCandles(ts.Symbol, CandleInterval, CandleStartTime, CandleEndTime)
		if err != nil{
			log.Fatalf("Error Fetching Candle in TA goroutine: %v", err)
		}
	ch0 <- true
	}(ch0)	
	
	// Calculate moving averages (MA) using historical data.
	ch1 := make(chan bool)
	ch2 := make(chan bool)
	ch3 := make(chan bool)
	otherTAChan := make(chan bool)
	var (
		short4EMA, long8EMA   []float64
		short15EMA, long95EMA []float64
	)
	ema4 := CandleExponentialMovingAverageV1(ts.ClosingPrices, 3)
	if len(ema4) > 95 && ts.DataPoint >= 95 {
		go func(otherTAChan chan bool) {
			// Calculate Relative Strength Index (RSI) using historical data.
			RSI, _ := CalculateRSI(ema4, 14)
			// Calculate Stochastic RSI and SmoothK, SmoothD values
			dataPoint.StochRSI, dataPoint.SmoothKRSI, _ = StochasticRSIV2(RSI, 14, 3, 3)
			// Calculate MACD and Signal Line using historical data.
			dataPoint.MACDLine, dataPoint.MACDSigLine, dataPoint.MACDHist, _ = CalculateMACDV2(9, ema4, 26, 12)
			otherTAChan <- true
		}(otherTAChan)
		go func(ch1 chan bool) {
			//Calculate EMA 4 period
			short4EMA = CandleExponentialMovingAverageV1(ema4, 4)
			ch1 <- true
		}(ch1)
		go func(ch2 chan bool) {
			long8EMA = CandleExponentialMovingAverageV1(ema4, 8)
			ch2 <- true
		}(ch2)
		go func(ch3 chan bool) {
			long95EMA = CandleExponentialMovingAverageV1(ema4, 95)
			ch3 <- true
		}(ch3)
		short15EMA = CandleExponentialMovingAverageV1(ema4, 15)
		<-ch1
		<-ch2
		<-ch3

		window := 20
		// Determine the buy and sell signals based on the moving averages, RSI, MACD line, and Bollinger Bands.
	
		ts.ShortPeriod = 15
		ts.LongPeriod = 95
		ts.Container1 = short15EMA
		ts.Container2 = long95EMA
		//for price determination
		L8EMA3, L8EMA0 := long8EMA[ts.DataPoint-3], long8EMA[ts.DataPoint]
		S4EMA3, S4EMA0 := short4EMA[ts.DataPoint-3], short4EMA[ts.DataPoint]

		cch1 := make(chan bool)
		cch2 := make(chan bool)
		cch3 := make(chan bool)
		PriceDownGoingDown := false
		PriceDownGoingUp := false
		PriceUpGoingUp := false
		PriceUpGoingDown := false
		go func(cch1 chan bool) {
			PriceDownGoingDown = (L8EMA3 > S4EMA3) && (L8EMA0-S4EMA0 > L8EMA3-S4EMA3) && (L8EMA0 > S4EMA0)
			PriceDownGoingUp = (L8EMA0 > S4EMA0) && (L8EMA0-S4EMA0 < L8EMA3-S4EMA3) && (L8EMA3 > S4EMA3)
			cch1 <- true
		}(cch1)
		go func(cch2 chan bool) {
			PriceUpGoingUp = (S4EMA3 > L8EMA3) && (S4EMA0-L8EMA0 > S4EMA3-L8EMA3) && (S4EMA0 > L8EMA0)
			PriceUpGoingDown = (S4EMA0 > L8EMA0) && (S4EMA0-L8EMA0 < S4EMA3-L8EMA3) && (S4EMA3 > L8EMA3)
			cch2 <- true
		}(cch2)

		//for market determination
		L95EMA6, L95EMA3, L95EMA1, L95EMA0 := long95EMA[ts.DataPoint-6], long95EMA[ts.DataPoint-3], long95EMA[ts.DataPoint-1], long95EMA[ts.DataPoint]
		S15EMA6, S15EMA3, S15EMA1, S15EMA0 := short15EMA[ts.DataPoint-6], short15EMA[ts.DataPoint-3], short15EMA[ts.DataPoint-1], short15EMA[ts.DataPoint]

		MarketDownGoneDown := false
		MarketUpGoneUp := false
		CrossL95S15UP := false
		go func(cch3 chan bool) {
			MarketDownGoneDown = (L95EMA0 > S15EMA0) && (L95EMA6 > S15EMA6) && (L95EMA0-S15EMA0) < (L95EMA1-S15EMA1) && (L95EMA1-S15EMA1) > (L95EMA3-S15EMA3) && (L95EMA3-S15EMA3) > (L95EMA6-S15EMA6)
			MarketUpGoneUp = (S15EMA0 > L95EMA0) && (S15EMA6 > L95EMA6) && (S15EMA0-L95EMA0) < (S15EMA1-L95EMA1) && (S15EMA1-L95EMA1) > (S15EMA3-L95EMA3) && (S15EMA3-L95EMA3) > (S15EMA6-L95EMA6)
			cch3 <- true
		}(cch3)

		go func(ch3 chan bool) {
			// Cross-EMA Relationships
			CrossL95S15UP = (L95EMA6 > S15EMA6) && (L95EMA0 < S15EMA0)
			if CrossL95S15UP {
				ts.CrossUPTime = time.Now()
			}
			// CrossL95S15DN = (L95EMA6 < S15EMA6) &&  (L95EMA0 > S15EMA0)
			// if CrossL95S15DN{
			// 	CrossDNTime = time.Now()
			// }
			if time.Since(ts.CrossUPTime) > time.Minute*6.0 {
				ts.CrossUPTime = time.Time{}
			}
			// if time.Since(CrossDNTime) > time.Minute * 6.0{
			// 	CrossDNTime = time.Time{}
			// }
			ch3 <- true
		}(ch3)

		
		MarketDownGoingDown := false
		MarketDownGoingUp := false
		MarketUpGoingUp := false
		MarketUpGoingDown := false
		go func(ch1 chan bool) {
			MarketDownGoingDown = (L95EMA3 > S15EMA3) && (L95EMA0-S15EMA0 > L95EMA3-S15EMA3) && (L95EMA0 > S15EMA0)
			MarketDownGoingUp = (L95EMA0 > S15EMA0) && (L95EMA0-S15EMA0 < L95EMA3-S15EMA3) && (L95EMA3 > S15EMA3)
			ch1 <- true
		}(ch1)
		go func(ch2 chan bool) {
			MarketUpGoingUp = (S15EMA3 > L95EMA3) && (S15EMA0-L95EMA0 > S15EMA3-L95EMA3) && (S15EMA0 > L95EMA0)
			MarketUpGoingDown = (S15EMA0 > L95EMA0) && (S15EMA0-L95EMA0 < S15EMA3-L95EMA3) && (S15EMA3 > L95EMA3)
			ch2 <- true
		}(ch2)

		<-cch1
		<-cch2
		<-cch3
		<-ch3
		<-ch1
		<-ch2

		go func(ch1 chan bool) {
			// Moving Averages of Differences
			// Calculate Moving Average of Differential
			MA5Diff := calculateMovingAverage(rollingDiff(long95EMA[len(long95EMA)-window:], short15EMA[len(short15EMA)-window:]), 5)
			dataPoint.MA5DiffL95S15 = MA5Diff[len(MA5Diff)-1]
			dataPoint.MA5DiffL8S4 = MA5Diff[len(MA5Diff)-2]
			ch1 <- true
		}(ch1)
		go func(ch2 chan bool) {
			// Volatility Measures
			dataPoint.StdDevL95 = rollingStdDev(long95EMA[len(long95EMA)-window:], window)
			dataPoint.StdDevS15 = rollingStdDev(short15EMA[len(short15EMA)-window:], window)
			ch2 <- true
		}(ch2)
		go func(ch3 chan bool) {
			// Historical Trends
			dataPoint.RoCL8, _ = CalculateROC(ema4, 12)
			dataPoint.RoCS4, _ = CalculateROC(ema4, 12)
			<-ch0
			var (
				highs []float64
				lows []float64
				closes []float64 
				volumes []float64
			)
			for _, v := range HD{
				highs = append(highs, v.High)
				lows = append(lows, v.Low)
				closes = append(closes, v.Close)
				volumes = append(volumes, v.Volume)
			}
			//note length of closes must be greeater than 14 in this case is 20
			dataPoint.ATR = AverageTrueRange(highs, lows, closes, 14)
			dataPoint.OBV =	OnBalanceVolume(closes, volumes)
			ch3 <- true
		}(ch3)
		
		// Derived Metrics
		dataPoint.DiffL95S15 = L95EMA0 - S15EMA0
		dataPoint.DiffL8S4 = L8EMA0 - S4EMA0

		// Lagged Features (1-day lag)
		if ts.DataPoint > 0 {
			// Assume ema95 is your calculated 95-period EMA slice
			dataPoint.LaggedL95EMA = getLaggedValue(long95EMA, 15)
			dataPoint.LaggedS15EMA = getLaggedValue(short15EMA, 15)
		}
		<-ch1

	//BUY CONDITION
		if (ts.CrossUPTime != time.Time{}) && (ts.CurrentPrice < S15EMA0) {
			buySignal = true
			ts.Log.Printf("TAB1 Signalled: BuY: at currentPrice: %.8f, CrossL95S15UP %v Secs", ts.CurrentPrice, time.Since(ts.CrossUPTime).Seconds())
		} else if (MarketUpGoingUp && PriceDownGoingUp) || (MarketDownGoingUp && PriceUpGoingUp) {
			buySignal = true
			ts.Log.Printf("TAB3 Signalled: BuY: at currentPrice: %.8f, PriceDownGoingDown: %v, PriceDownGoingUp: %v, PriceUpGoingUp: %v, PriceUpGoingDown: %v, MarketDownGoingDown: %v, MarketDownGoingUp: %v, MarketUpGoingUp: %v, MarketUpGoingDown: %v", ts.CurrentPrice, PriceDownGoingDown, PriceDownGoingUp, PriceUpGoingUp, PriceUpGoingDown, MarketDownGoingDown, MarketDownGoingUp, MarketUpGoingUp, MarketUpGoingDown)
		} else if MarketDownGoneDown {
			buySignal = true
			ts.Log.Printf("TAB4 Signalled: BuY: at currentPrice: %.8f, MarketDownGoneDown: %v,  DiffL95S15: %.8f", ts.CurrentPrice, MarketDownGoneDown, math.Abs(dataPoint.DiffL95S15))
		} else if dataPoint.MA5DiffL95S15 > 0 && dataPoint.MA5DiffL8S4 <= 0 {
			buySignal = true
			ts.Log.Printf("TAB5 Signalled: BuY: at currentPrice: %.8f, (MA5DiffL95S15 %.8f > 0 && MA5DiffL8S4 %.8f <= 0) = %v ", ts.CurrentPrice, dataPoint.MA5DiffL95S15, dataPoint.MA5DiffL8S4, (dataPoint.MA5DiffL95S15 > 0 && dataPoint.MA5DiffL8S4 <= 0))
		
	//SELL CONDITION
		}else if (MarketUpGoingDown && PriceDownGoingDown) || (MarketDownGoingDown && PriceUpGoingDown) {
			sellSignal = true
			ts.Log.Printf("TAS3 Signalled: SeLL: at currentPrice: %.8f, PriceDownGoingDown: %v, PriceDownGoingUp: %v, PriceUpGoingUp: %v, PriceUpGoingDown: %v, MarketDownGoingDown: %v, MarketDownGoingUp: %v, MarketUpGoingUp: %v, MarketUpGoingDown: %v", ts.CurrentPrice, PriceDownGoingDown, PriceDownGoingUp, PriceUpGoingUp, PriceUpGoingDown, MarketDownGoingDown, MarketDownGoingUp, MarketUpGoingUp, MarketUpGoingDown)
		} else if MarketUpGoneUp {
			sellSignal = true
			ts.Log.Printf("TAS4 Signalled: SeLL: at currentPrice: %.8f, MarketUpGoneUp: %v, DiffL95S15: %.8f", ts.CurrentPrice, MarketUpGoneUp, math.Abs(dataPoint.DiffL95S15))
		} else if dataPoint.MA5DiffL95S15 < 0 && dataPoint.MA5DiffL8S4 >= 0 {
			sellSignal = true
			ts.Log.Printf("TAS5 Signalled: SeLL: at currentPrice: %.8f, (MA5DiffL95S15 %.8f < 0 && MA5DiffL8S4 %.8f >= 0) = %v", ts.CurrentPrice, dataPoint.MA5DiffL95S15, dataPoint.MA5DiffL8S4, (dataPoint.MA5DiffL95S15 < 0 && dataPoint.MA5DiffL8S4 >= 0))
		} else {
			ts.Log.Printf("TA0 Signalled: Missed: at currentPrice: %.8f, PriceDownGoingDown: %v, PriceDownGoingUp: %v, PriceUpGoingUp: %v, PriceUpGoingDown: %v, MarketDownGoingDown: %v, MarketDownGoingUp: %v, MarketUpGoingUp: %v, MarketUpGoingDown: %v", ts.CurrentPrice, PriceDownGoingDown, PriceDownGoingUp, PriceUpGoingUp, PriceUpGoingDown, MarketDownGoingDown, MarketDownGoingUp, MarketUpGoingUp, MarketUpGoingDown)
		}
		<-ch2
		<-ch3
		<-otherTAChan
		ts.Log.Printf("Legnth of Historical Data is = %d", len(HD))
	}
	return buySignal, sellSignal
}
// Function to calculate moving average
func calculateMovingAverage(data []float64, period int) []float64 {
    ma := make([]float64, len(data))
    var sum float64

    for i := 0; i < period; i++ {
        sum += data[i]
        ma[i] = sum / float64(i+1)
    }

    for i := period; i < len(data); i++ {
        sum += data[i] - data[i-period]
        ma[i] = sum / float64(period)
    }

    return ma
}
		
		// Incorporate in trading logic
		// if currentEMA15 > laggedEMA95 {
		// 	// Possible buy signal
		// } else if currentEMA15 < laggedEMA95 {
		// 	// Possible sell signal
		// }

		//Stochastic RSI	
		// buySignal = (stochRSI[ts.DataPoint] > ts.StRSIOverbought && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] > ts.StRSIOverbought)
		// sellSignal = (stochRSI[ts.DataPoint] < md.StRSIOversold && smoothKRSI[ts.DataPoint-md.StochRSIPeriod-md.SmoothK] < md.StRSIOversold)
		//EMA
		// buySignal = (shortEMA[ts.DataPoint-1] < longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] >= longEMA[ts.DataPoint])
		// sellSignal = (shortEMA[ts.DataPoint-1] >= longEMA[ts.DataPoint-1] && shortEMA[ts.DataPoint] < longEMA[ts.DataPoint])
		//RSI
		// buySignal = (rsi[ts.DataPoint-1] < md.RSIOversold)
		// sellSignal = (rsi[ts.DataPoint-1] > md.RSIOverbought)
		//MACD
		// buySignal = (macdLine[ts.DataPoint-1] <= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] > signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] > 0)
		// sellSignal = (macdLine[ts.DataPoint-1] >= signalLine[ts.DataPoint] && macdLine[ts.DataPoint] < signalLine[ts.DataPoint] && macdHistogram[ts.DataPoint] < 0)
		
func rollingDiff(L, S []float64) (diff []float64) {
	for k, v := range L {
		diff = append(diff, v-S[k])
	}
	return diff
}

// Function to calculate rolling mean with handling for NaN
func rollingMean(data []float64, window int) float64 {
	if len(data) < window {
		// Handle the case where there are not enough data points
		return 0.0 // or any default value you prefer
	}

	// Take the last 'window' elements from the slice
	data = data[len(data)-window:]

	return stat.Mean(data, nil)
}

func getLaggedValue(ema []float64, lag int) float64 {
    if len(ema) > lag {
        // Return the value 'lag' periods ago
        return ema[len(ema)-1-lag]
    }
    return 0.0 // or handle this case as needed
}

// Function to calculate rolling standard deviation with handling for NaN
func rollingStdDev(data []float64, window int) float64 {
	if len(data) < window {
		// Handle the case where there are not enough data points
		return 0.0 // or any default value you prefer
	}

	// Take the last 'window' elements from the slice
	data = data[len(data)-window:]

	return stat.StdDev(data, nil)
}

// CalculatePercentileRank calculates the percentile rank of a value in a dataset.
func CalculatePercentileRank(value float64, data []float64) float64 {
	count := 0
	for _, v := range data {
		if v <= value {
			count++
		}
	}
	return float64(count) / float64(len(data))
}

// CalculateROC calculates the Rate of Change for a given slice of prices and a period.
// data is a slice of float64 representing the prices, and period is the number of intervals for the ROC.
func CalculateROC(data []float64, period int) (float64, error) {
	if period <= 0 {
		return 0, fmt.Errorf("period must be greater than 0")
	}
	if len(data) < period {
		return 0, fmt.Errorf("data slice is too short for the given period")
	}

	// Get the most recent price and the price 'period' intervals ago.
	mostRecentPrice := data[len(data)-1]
	previousPrice := data[len(data)-1-period]

	// Avoid division by zero.
	if previousPrice == 0 {
		return 0, fmt.Errorf("previous price is zero, cannot calculate ROC")
	}

	// Calculate ROC.
	roc := (mostRecentPrice - previousPrice) / previousPrice * 100
	return roc, nil
}
// SimpleMovingAverage calculates the simple moving average for a slice of data
func SimpleMovingAverage(data []float64, period int) float64 {
    if len(data) < period {
        return 0 // or handle error as per your need
    }

    sum := 0.0
    for _, value := range data[len(data)-period:] {
        sum += value
    }

    return sum / float64(period)
}
func OnBalanceVolume(closes, volumes []float64) int64 {
    obv := int64(0)

    for i := 1; i < len(closes); i++ {
        if closes[i] > closes[i-1] {
            obv += int64(volumes[i])
        } else if closes[i] < closes[i-1] {
            obv -= int64(volumes[i])
        }
    }
    return obv
}

func AverageTrueRange(highs, lows, closes []float64, period int) float64 {
    trs := make([]float64, len(closes)-1)
    for i := 1; i < len(closes); i++ {
        tr1 := highs[i] - lows[i]
        tr2 := math.Abs(highs[i] - closes[i-1])
        tr3 := math.Abs(lows[i] - closes[i-1])
        trs[i-1] = math.Max(math.Max(tr1, tr2), tr3)
    }
    return SimpleMovingAverage(trs, period)
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
func (ts *TradingSystem) StrategyProcessing(dataPoint *model.DataPoint) string {
	// Combine the results of technical and fundamental analysis to decide entry conditions.
	technicalBuy, technicalSell := ts.TechnicalAnalysis(dataPoint)
	// AIBuy, AISell := ts.AIAnalysis(dataPoint)
	// _, _ = AIBuy, AISell
	if technicalBuy{
		dataPoint.Label = 1
		return "buy"
	}else if technicalSell{
		dataPoint.Label = -1
		return "sell"
	}else {
		dataPoint.Label = 0
		return "hold"
	}
}

func (ts *TradingSystem) Reporting(from string) error {
	var err error

	if (len(ts.Container1) > 0) && (len(ts.Container1) <= ts.LongPeriod) {
		err = ts.CreateLineChartWithSignals(ts.Timestamps, ts.ClosingPrices, ts.Signals, "")
	} else if len(ts.Container1) <= ts.Zoom {
		err = ts.CreateLineChartWithSignalsV3(ts.Timestamps, ts.ClosingPrices, ts.Container1, ts.Container2, ts.Signals, "")
	} else {
		if len(ts.Signals) < len(ts.ClosingPrices) {
			log.Printf("Lengths of ts.Timestamps %d, ts.Container1 %d, ts.Container2 %d, ts.Signals %d, ts.ClosingPrices %d\n", len(ts.Timestamps), len(ts.Container1), len(ts.Container2), len(ts.Signals), len(ts.ClosingPrices))
			ts.Log.Printf("Lengths of ts.Timestamps %d, ts.Container1 %d, ts.Container2 %d, ts.Signals %d, ts.ClosingPrices %d\n", len(ts.Timestamps), len(ts.Container1), len(ts.Container2), len(ts.Signals), len(ts.ClosingPrices))
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
		}
		b := len(ts.ClosingPrices) - 1
		a := len(ts.ClosingPrices) - ts.Zoom
		err = ts.CreateLineChartWithSignalsV3(ts.Timestamps[a:b], ts.ClosingPrices[a:b], ts.Container1[a:b], ts.Container2[a:b], ts.Signals[a:b], "")
	}
	if err != nil {
		return fmt.Errorf("Error creating Line Chart with signals: %v", err)
	}
	return nil
}

func (ts *TradingSystem) ShutDown(sigchnl chan os.Signal) {
	for {
		select {
		case sig := <-sigchnl:
			if sig == syscall.SIGINT {
				log.Printf("Received signal: %v. Exiting...\n", sig)
				//Check if there is still asset remainning and sell off
				if ts.BaseBalance > 0.0 {
					// After sell off Update the quote and base balances after the trade.

					// Calculate profit/loss for the trade.
					localProfitLoss := CalculateProfitLoss(ts.EntryPrice[len(ts.EntryPrice)-1], ts.CurrentPrice, ts.BaseBalance)
					transactionCost := ts.CommissionPercentage * ts.CurrentPrice * ts.BaseBalance

					// Store profit/loss for the trade.

					localProfitLoss -= transactionCost
					// ts.TotalProfitLoss += localProfitLoss

					ts.Mu.Lock()
					ts.QuoteBalance += (ts.BaseBalance * ts.CurrentPrice) - transactionCost
					ts.BaseBalance -= ts.BaseBalance
					ts.Signals = append(ts.Signals, "Sell")
					ts.TradeCount++
					ts.Mu.Unlock()
					log.Printf("- SELL-OFF at EntryPrice[%d]: %.8f, EntryQuantity[%d]: %.8f, QBal: %.8f, BBal: %.8f, GlobalP&L %.2f LocalP&L: %.8f PosPcent: %.8f tsDataPt: %d \n",
						len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, ts.TotalProfitLoss, localProfitLoss, ts.RiskPositionPercentage, ts.DataPoint)
				}
				// Print the overall trading performance after backtesting.
				log.Printf("Total Trades: %d, out of %d trials ", ts.TradeCount, len(ts.Signals))
				log.Printf("Total Profit/Loss: %.2f, ", ts.TotalProfitLoss)
				log.Printf("Final Capital: %.2f, ", ts.QuoteBalance)
				log.Printf("Final Asset: %.8f tsDataPoint: %d\n\n", ts.BaseBalance, ts.DataPoint)
				ts.DBStoreTicker.Stop()
				// if err := ts.APIServices.CloseDB(); err != nil{
				// 	log.Printf("Error while closing the DataBase: %v", err)
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

	long8EMA, short4EMA, err := CandleExponentialMovingAverage(closingPrices, LongPeriod, ShortPeriod)
	if err != nil {
		log.Fatalf("Error: in CalclateMACD while tring to get EMA")
	}
	// Calculate MACD line
	macdLine = make([]float64, len(closingPrices))
	for i := range closingPrices {
		macdLine[i] = short4EMA[i] - long8EMA[i]
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
// CalculateMACD calculates the Moving Average Convergence Divergence (MACD) and MACD Histogram for the given data and periods.
func CalculateMACDV2(SignalMACDPeriod int, closingPrices []float64, LongPeriod, ShortPeriod int) (MACDLine, SignalLine, MACDHistogram float64, err error) {
	if LongPeriod <= 0 || len(closingPrices) < LongPeriod {
		err = fmt.Errorf("Error Calculating EMA: not enoguh data for period %v at total datapoint: %d", LongPeriod, len(closingPrices)-1)
		return 0.0, 0.0, 0.0, err
	}
	var macdLine, signalLine, macdHistogram []float64
	long8EMA, short4EMA, err := CandleExponentialMovingAverage(closingPrices, LongPeriod, ShortPeriod)
	if err != nil {
		log.Fatalf("Error: in CalclateMACD while tring to get EMA")
	}
	// Calculate MACD line
	macdLine = make([]float64, len(closingPrices))
	for i := range closingPrices {
		macdLine[i] = short4EMA[i] - long8EMA[i]
	}

	// Calculate signal line using the MACD line
	signalLine, err = CalculateExponentialMovingAverage(macdLine, SignalMACDPeriod)

	// Calculate MACD Histogram
	macdHistogram = make([]float64, len(closingPrices))
	for i := range closingPrices {
		macdHistogram[i] = macdLine[i] - signalLine[i]
	}

	return macdLine[len(macdLine)-1], signalLine[len(signalLine)-1], macdHistogram[len(macdHistogram)-1], err
}
//CandleExponentialMovingAverage calculates EMA from condles
func CandleExponentialMovingAverage(closingPrices []float64, LongPeriod, ShortPeriod int) (long8EMA, short4EMA []float64, err error) {
	if LongPeriod <= 0 || len(closingPrices) < LongPeriod || closingPrices == nil {
		return nil, nil, fmt.Errorf("Error Calculating Candle EMA: not enoguh data for period %v", LongPeriod)
	}
	var ema95, ema15 *ema.Ema
	ema95 = ema.NewEma(alphaFromN(LongPeriod))
	ema15 = ema.NewEma(alphaFromN(ShortPeriod))

	long8EMA = make([]float64, len(closingPrices))
	short4EMA = make([]float64, len(closingPrices))
	for k, closePrice := range closingPrices {
		ema95.Step(closePrice)
		ema15.Step(closePrice)
		long8EMA[k] = ema95.Compute()
		short4EMA[k] = ema15.Compute()
	}
	return long8EMA, short4EMA, nil
}

//CandleExponentialMovingAverage calculates EMA from condles
func CandleExponentialMovingAverageV2(closingPrices []float64, ShortPeriod int) (short4EMA []float64, err error) {
	if ShortPeriod <= 0 || len(closingPrices) < ShortPeriod || closingPrices == nil {
		return nil, fmt.Errorf("Error Calculating Candle EMA: not enoguh data for period %v", ShortPeriod)
	}
	var ema15 *ema.Ema
	ema15 = ema.NewEma(alphaFromN(ShortPeriod))

	short4EMA = make([]float64, len(closingPrices))
	for k, closePrice := range closingPrices {
		ema15.Step(closePrice)
		short4EMA[k] = ema15.Compute()
	}
	return short4EMA, nil
}

//CandleExponentialMovingAverage calculates EMA from condles
func CandleExponentialMovingAverageV1(closingPrices []float64, ShortPeriod int) (short4EMA []float64) {
	if ShortPeriod <= 0 || len(closingPrices) < ShortPeriod || closingPrices == nil {
		return nil
	}
	var ema15 *ema.Ema
	ema15 = ema.NewEma(alphaFromN(ShortPeriod))

	short4EMA = make([]float64, len(closingPrices))
	for k, closePrice := range closingPrices {
		ema15.Step(closePrice)
		short4EMA[k] = ema15.Compute()
	}
	return short4EMA
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
func CalculateQuantity(entryPrice, exitPrice, profitLoss float64) float64 {
	return profitLoss / (exitPrice - entryPrice)
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
// StochasticRSI calculates the Stochastic RSI for a given RSI data series, period, and SmoothK, SmoothD.
func StochasticRSIV2(rsi []float64, period, smoothK, smoothD int) (float64, float64, error) {
	if period <= 0 || len(rsi) < period+1 {
		return 0.0, 0.0, fmt.Errorf("Error Calculating StochRSI: not enoguh data for period %v", period)
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
		return 0.0, 0.0, fmt.Errorf("Error Calculating EMA in Stochastic RSI: %v", err)
	}
	emaSmoothD, err := CalculateExponentialMovingAverage(emaSmoothK, smoothD)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("Error Calculating EMA in Stochastic RSI: %v", err)
	}

	return emaSmoothK[len(emaSmoothK)-1], emaSmoothD[len(emaSmoothD)-1], nil
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
