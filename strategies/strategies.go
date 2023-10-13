package strategies

import (
	"encoding/csv"
	"encoding/json"
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
	StopLossRecover          []float64
	RiskFactor               float64
	MaxDataSize              int
	Log                      *log.Logger
	ShutDownCh               chan string
	EpochTime                time.Duration
	CSVWriter                *csv.Writer
	RiskProfitLossPercentage float64
	StoreAppDataChan         chan string
	BaseCurrency             string //in Binance is called BaseAsset
	QuoteCurrency            string //in Binance is called QuoteAsset
	MiniQty                  float64
	MaxQty                   float64
	MinNotional              float64
	StepSize                 float64
	DBServices               model.DBServices
	RDBServices              *RDBServices
	DBStoreTicker            *time.Ticker
	TSDataChan               chan []byte
	ADataChan                chan []byte
	MDChan                   chan *model.AppData
	Zoom                     int
}

// NewTradingSystem(): This function initializes the TradingSystem and fetches
// historical data from the exchange using the GetCandlesFromExch() function.
// It sets various strategy parameters like short and long EMA periods, RSI period,
// MACD signal period, Bollinger Bands period, and more.
func NewTradingSystem(BaseCurrency string, liveTrading bool, loadExchFrom, loadDBFrom string) (*TradingSystem, error) {
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
			ts.RiskFactor = 2.0
			ts.CommissionPercentage = 0.00075
			ts.RiskProfitLossPercentage = 0.0008
			ts.EnableStoploss = true
			ts.StopLossRecover = append(ts.StopLossRecover, math.MaxFloat64)
			ts.MaxDataSize = 500
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
			ts.RiskFactor = 2.0
			ts.CommissionPercentage = 0.00075
			ts.RiskProfitLossPercentage = 0.0008
			ts.EnableStoploss = true
			ts.StopLossRecover = append(ts.StopLossRecover, math.MaxFloat64)
			ts.MaxDataSize = 500
			ts.BaseCurrency = BaseCurrency
		} else {
			loadDataFrom = "DataBase"
			// ts.InitialCapital = 54.038193 + 26.47
			// ts.RiskProfitLossPercentage = 0.0008

			// mdTargetProfit := (54.038193 + 26.47) * 0.0008
			// mdTargetStopLoss := (54.038193 + 26.47) * 0.0008
			// if ts.TradingLevel >= 2{
			// 	mdTargetProfit = mdTargetProfit + ((mdTargetProfit * float64(ts.TradingLevel))/8.0)			
			// }
			// ts.EntryPrice = append(ts.EntryPrice, 206.2)
			// ts.EntryQuantity = append(ts.EntryQuantity, 0.126)
			// ts.EntryCostLoss = append(ts.EntryCostLoss, (0.00075 * 206.2 * 0.126))
			// nextProfitSeLLPrice := ((mdTargetProfit + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / 0.126) + ts.EntryPrice[len(ts.EntryPrice)-1]
			// nextInvBuYPrice := (-(mdTargetStopLoss + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / 0.126) + ts.EntryPrice[len(ts.EntryPrice)-1]
			// commissionAtProfitSeLLPrice := nextProfitSeLLPrice * 0.126 * ts.CommissionPercentage
			// commissionAtInvBuYPrice := nextInvBuYPrice * 0.126 * ts.CommissionPercentage
			// ts.NextProfitSeLLPrice = append(ts.NextProfitSeLLPrice, nextProfitSeLLPrice+commissionAtProfitSeLLPrice)
			// ts.NextInvestBuYPrice = append(ts.NextInvestBuYPrice, nextInvBuYPrice-commissionAtInvBuYPrice)
			// ts.TradingLevel = len(ts.EntryPrice)
			// ts.InTrade = true
			// ts.Signals = append(ts.Signals, "Buy")
		}
	}
	fmt.Println("TS = ", ts)
	if liveTrading {
		err = ts.LiveUpdate(loadExchFrom, loadDBFrom, loadDataFrom)
		if err != nil {
			return &TradingSystem{}, err
		}
	} else {
		err = ts.UpdateHistoricalData(loadExchFrom, loadDBFrom)
		if err != nil {
			return &TradingSystem{}, err
		}
	}
	ts.RDBServices = rDBServices
	if !strings.Contains(loadExchFrom, "Testnet") {
		ts.Zoom = 100
	}else{
		ts.Zoom = 499			
	}
	ts.ShutDownCh = make(chan string)
	ts.EpochTime = time.Second * 10
	ts.StoreAppDataChan = make(chan string, 1)
	ts.DBStoreTicker = time.NewTicker(ts.EpochTime)
	ts.TSDataChan = make(chan []byte)
	ts.ADataChan = make(chan []byte)
	ts.MDChan = make(chan *model.AppData)
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
			}
			if len(ts.EntryPrice) >= 1 {
				trade.EntryCostLoss = ts.EntryCostLoss
				trade.EntryQuantity = ts.EntryQuantity
				trade.EntryPrice = ts.EntryPrice
				trade.NextInvestBuYPrice = ts.NextInvestBuYPrice
				trade.NextProfitSeLLPrice = ts.NextProfitSeLLPrice
			}

			// Serialize the DBAppData object to JSON
			appDataJSON, err := json.Marshal(trade)
			if err != nil {
				log.Printf("Error marshaling TradingSystem to JSON: %v", err)
			} else {
				select {
				case ts.TSDataChan <- appDataJSON:
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
				case <-time.After(time.Second * 900):
					if err = ts.RDBServices.UpdateDBTradingSystem(ts); err != nil {
						log.Printf("Creating TradingSystem happening now!!! where %v", err)
						ts.ID, err = ts.RDBServices.CreateDBTradingSystem(ts)
						if err != nil {
							log.Printf("Error Creating TradingSystem: %v", err)
						}
					}
					ts.StoreAppDataChan <- ""
					time.Sleep(time.Second * 10)
				}
			}
		}()
	}
	return ts, nil
}
func (ts *TradingSystem) NewAppData(loadExchFrom string) *model.AppData {
	// Initialize the App Data
	// loadDataFrom := ""
	var (
		md  *model.AppData
		err error
	)
	if loadExchFrom == "BinanceTestnet" {
		md, err = &model.AppData{}, fmt.Errorf(("Testnet Error simulation"))
		if err != nil {
			fmt.Println("MD = ", md)
			log.Printf("\n%v: But going ahead to initialize empty AppData struct\n", err)
			md = &model.AppData{}
			md.DataPoint = 0
			md.Strategy = "EMA"
			md.ShortPeriod = 15 //10 Define moving average short period for the strategy.
			md.LongPeriod = 55  //30 Define moving average long period for the strategy.
			md.ShortEMA = 0.0
			md.LongEMA = 0.0
			md.TargetProfit = ts.InitialCapital * ts.RiskProfitLossPercentage
			md.TargetStopLoss = ts.InitialCapital * ts.RiskProfitLossPercentage
			md.RiskPositionPercentage = 0.25 // Define risk management parameter 5% balance
			md.TotalProfitLoss = 0.0
		}
	} else {
		rDBServices := NewRDBServices(loadExchFrom)
		md, err = rDBServices.ReadDBAppData(0)
		if err != nil {
			fmt.Println("MD = ", md)
			log.Printf("\n%v: But going ahead to initialize empty AppData struct\n", err)
			md = &model.AppData{}
			md.DataPoint = 0
			md.Strategy = "EMA"
			md.ShortPeriod = 15 // Define moving average short period for the strategy.
			md.LongPeriod = 55  // Define moving average long period for the strategy.
			md.ShortEMA = 0.0
			md.LongEMA = 0.0
			md.TargetProfit = ts.InitialCapital * ts.RiskProfitLossPercentage
			md.TargetStopLoss = ts.InitialCapital * ts.RiskProfitLossPercentage
			md.RiskPositionPercentage = 0.25 // Define risk management parameter 5% balance
			md.TotalProfitLoss = 0.0
		}else{
			// md.ShortPeriod = 15 //10 Define moving average short period for the strategy.
			// md.LongPeriod = 55  //30 Define moving average long period for the strategy.
			md.TargetProfit = (54.038193 + 26.47) * 0.0008
			md.TargetStopLoss = (54.038193 + 26.47) * 0.0008
		}
	}
	fmt.Println("MD = ", md)
	go func() {
		for {
			select {
			case ts.MDChan <- md:
			}
		}
	}()
	//Sending md to react Frontend UI
	go func() {
		log.Printf("Ready to send AppDatat to Frontend UI")
		for { // Serialize the DBAppData object to JSON for the UI frontend
			appDataJSON, err := json.Marshal(md)
			if err != nil {
				log.Printf("Error2 marshaling DBAppData to JSON: %v", err)
				// panic(fmt.Sprintf("Do not panic just look up the trail path: AppData at this panic = %v", md))
			} else {
				select {
				case ts.ADataChan <- appDataJSON:
				}
			}
			time.Sleep(time.Second * 1)
		}
	}()
	//Updating md to Database
	go func() {
		var err error
		log.Printf("Ready to send AppDatat to Database")
		for {
			select {
			case <-ts.StoreAppDataChan:
				if err = ts.RDBServices.UpdateDBAppData(md); err != nil {
					log.Printf("Creating AppData happening now!!!")
					md.ID, err = ts.RDBServices.CreateDBAppData(md)
					if err != nil {
						log.Printf("Error Storing AppData: %v", err)
					}
				}
			}
		}
	}()
	return md
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) UpdateHistoricalData(loadExchFrom, loadDBFrom string) error {
	var (
		err             error
		exch            model.APIServices
		exchConfigParam *config.ExchConfig
		dBConfigParam   *config.DBConfig
		ok              bool
		DB              model.DBServices
	)

	if exchConfigParam, ok = config.NewExchangeConfigs()[loadExchFrom]; !ok {
		fmt.Println("Error Internal: Exch name not", loadExchFrom)
		return errors.New("Exch name not " + loadExchFrom)
	}
	// Check if REQUIRED "InfluxDB" exists in the map
	if dBConfigParam, ok = config.NewDataBaseConfigs()[loadDBFrom]; !ok {
		fmt.Println("Error Internal: DB Manager Required")
		return errors.New("Required DB services not contracted")
	}
	switch loadDBFrom {
	case "InfluxDB":
		//Initiallize InfluDB config Parameters and instanciate its struct
		DB, err = influxdb.NewAPIServices(dBConfigParam)
		if err != nil {
			log.Fatalf("Error getting Required services from InfluxDB: %v", err)
		}
	default:
		return errors.Errorf("Error updating historical data from %s and %s: invalid loadExchFrom tags \"%s\" and \"%s\" ", loadExchFrom, loadDBFrom, loadExchFrom, loadDBFrom)
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
		return errors.Errorf("Error updating historical data from %s and %s invalid loadFrom tags \"%s\" and \"%s\" ", loadExchFrom, loadDBFrom, loadExchFrom, loadDBFrom)
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
		// err := ts.DBServices.WriteCandleToDB(candle.Close, candle.Timestamp)
		// if (err != nil) && (!strings.Contains(fmt.Sprintf("%v", err), "Skipping write")) {
		// 	log.Fatalf("Error: writing to influxDB: %v", err)
		// }
	}
	return nil
}

// UpdateClosingPrices fetches historical data from the exchange and updates the ClosingPrices field in TradingSystem.
func (ts *TradingSystem) LiveUpdate(loadExchFrom, loadDBFrom, LoadDataFrom string) error {
	var (
		err             error
		exch            model.APIServices
		exchConfigParam *config.ExchConfig
		dBConfigParam   *config.DBConfig
		ok              bool
		DB              model.DBServices
	)

	if exchConfigParam, ok = config.NewExchangeConfigs()[loadExchFrom]; !ok {
		fmt.Println("Error Internal: Exch name not", loadExchFrom)
		return errors.New("Exch name not " + loadExchFrom)
	}
	// Check if REQUIRED "InfluxDB" exists in the map
	if dBConfigParam, ok = config.NewDataBaseConfigs()["InfluxDB"]; !ok {
		fmt.Println("Error Internal: Exch Required InfluxDB")
		return errors.New("Required InfluxDB services not contracted")
	}
	switch loadDBFrom {
	case "InfluxDB":
		//Initiallize InfluDB config Parameters and instanciate its struct
		DB, err = influxdb.NewAPIServices(dBConfigParam)
		if err != nil {
			log.Fatalf("Error getting Required services from InfluxDB: %v", err)
		}
	default:
		return errors.Errorf("Error updating Live data from %s and %s invalid loadFrom tags \"%s\" and \"%s\" ", loadExchFrom, loadDBFrom, loadExchFrom, loadDBFrom)
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
		return errors.Errorf("Error updating Live data from %s and %s invalid loadFrom tags \"%s\" and \"%s\" ", loadExchFrom, loadDBFrom, loadExchFrom, loadDBFrom)
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
	md := ts.NewAppData(loadExchFrom)
	fmt.Println("App started. Press Ctrl+C to exit.")
	go ts.ShutDown(md, sigchnl)
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
		md.DataPoint++
		ts.ClosingPrices = append(ts.ClosingPrices, ts.CurrentPrice)
		ts.Timestamps = append(ts.Timestamps, time.Now().Unix())
		//Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading
		//Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading Trading
		ts.Trading(md, loadExchFrom)

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
func (ts *TradingSystem) Backtest(loadExchFrom string) {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)
	//Count,Strategy,ShortPeriod,LongPeriod,ShortMACDPeriod,LongMACDPeriod,
	//SignalMACDPeriod,RSIPeriod,StochRSIPeriod,SmoothK,SmoothD,RSIOverbought,RSIOversold,
	//StRSIOverbought,StRSIOversold,BollingerPeriod,BollingerNumStdDev,TargetProfit,
	//TargetStopLoss,RiskPositionPercentage
	backT := []*model.AppData{ //StochRSI
		{0, 0, "EMA", 20, 55, 0.0, 0.0, 0.5, 3.0, 0.25, 0.0},
	}

	fmt.Println("App started. Press Ctrl+C to exit.")
	var err error
	for i, md := range backT {
		md = ts.NewAppData(loadExchFrom)
		// Simulate the backtesting process using historical price data.
		for ts.DataPoint, ts.CurrentPrice = range ts.ClosingPrices {
			select {
			case <-sigchnl:
				fmt.Println("ShutDown Requested: Press Ctrl+C again to shutdown...")
				sigchnl := make(chan os.Signal, 1)
				signal.Notify(sigchnl)
				ts.ShutDown(md, sigchnl)
			default:
			}

			ts.Trading(md, loadExchFrom)

			// time.Sleep(ts.EpochTime)
		}
		err = ts.Reporting(md, "Backtesting")
		if err != nil {
			fmt.Println("Error Reporting BackTest Trading: ", err)
			return
		}

		fmt.Println("Press Ctrl+C to continue...", md.ID)
		<-sigchnl
		if i == len(backT)-1 {
			fmt.Println("End of Backtesting: Press Ctrl+C again to shutdown...")
			sigchnl = make(chan os.Signal, 1)
			signal.Notify(sigchnl)
			ts.ShutDown(md, sigchnl)
		}

		if err != nil {
			fmt.Println(err)
		}
	}
}

func (ts *TradingSystem) Trading(md *model.AppData, loadExchFrom string) {
	// Execute the trade if entry conditions are met.
	if (!ts.InTrade) && (ts.EntryRule(md)) && (ts.CurrentPrice <= ts.StopLossRecover[len(ts.StopLossRecover)-1]) {
		// Execute the buy order using the ExecuteStrategy function.
		resp, err := ts.ExecuteStrategy(md, "Buy")
		if err != nil {
			// fmt.Println("Error executing buy order:", err)
			ts.Signals = append(ts.Signals, "Hold") // No Signal - Hold Position
		} else if strings.Contains(resp, "BUY") {
			// Record Signal for plotting graph later
			ts.Signals = append(ts.Signals, "Buy")
			ts.Log.Println(resp)
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
}

// ExecuteStrategy executes the trade based on the provided trade action and current price.
// The tradeAction parameter should be either "Buy" or "Sell".
func (ts *TradingSystem) ExecuteStrategy(md *model.AppData, tradeAction string) (string, error) {
	var (
		orderResp model.Response
		err       error
	)
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
		// adjustedPrice := math.Floor(price/lotSizeStep) * lotSizeStep
		quantity := math.Floor(ts.PositionSize/ts.MiniQty) * ts.MiniQty
		// Calculate the total cost of the trade
		totalCost := quantity * ts.CurrentPrice
		//check if there is enough Quote (USDT) for the buy transaction
		if ts.QuoteBalance < totalCost {
			ts.Log.Printf("Unable to buY!!!! Insufficient QuoteBalance: %.8f (ts.PositionSize %.8f* ts.CurrentPrice %.8f) = %.8f", ts.QuoteBalance, ts.PositionSize, ts.CurrentPrice, ts.PositionSize*ts.CurrentPrice)
			quantity = ts.QuoteBalance / ts.CurrentPrice
			quantity = math.Floor(quantity/ts.MiniQty) * ts.MiniQty
			if quantity < ts.MiniQty {
				return "", fmt.Errorf("cannot execute a buy order due to insufficient QuoteBalance: %.8f (ts.PositionSize %.8f* ts.CurrentPrice %.8f) = %.8f", ts.QuoteBalance, ts.PositionSize, ts.CurrentPrice, ts.PositionSize*ts.CurrentPrice)
			}
		}
		// Check if the total cost meets the minNotional requirement
		if totalCost < ts.MinNotional {
			ts.Log.Printf("Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, ts.CurrentPrice, totalCost, ts.MinNotional)
			return "", fmt.Errorf("Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, ts.CurrentPrice, totalCost, ts.MinNotional)
		}
		//Placing a Buy order
		orderResp, err = ts.APIServices.PlaceLimitOrder(ts.Symbol, "BUY", ts.CurrentPrice, quantity)
		if err != nil {
			fmt.Println("Error placing entry order:", err)
			ts.Log.Println("Error placing entry order:", err)
			return "", fmt.Errorf("Error placing entry order: %v", err)
		}
		//The cummulativeQuoteQty represents the total quote asset quantity (e.g., USDT) transacted. So, in the calculation for the average price:
		averagePrice := orderResp.CumulativeQuoteQty / orderResp.ExecutedQty
		//For a buy order:
		//the commission from Binance being in BaseCurrency was deducted from the executedQty
		totalCost = (orderResp.ExecutedQty * averagePrice)
		ts.Log.Printf("Entry order placed with ID: %d Commission: %.8f CumulativeQuoteQty: %.8f ExecutedPrice: %.8f ExecutedQty: %.8f Status: %s\n", orderResp.OrderID, orderResp.Commission, orderResp.CumulativeQuoteQty,
			orderResp.ExecutedPrice, orderResp.ExecutedQty, orderResp.Status)
		// Update the profit, quote and base balances after the trade.
		ts.QuoteBalance -= totalCost
		ts.BaseBalance += orderResp.ExecutedQty
		md.TotalProfitLoss -= (orderResp.Commission * averagePrice)
		mdTargetProfit := md.TargetProfit 
		if ts.TradingLevel >= 2{
			mdTargetProfit = md.TargetProfit + ((md.TargetProfit * float64(ts.TradingLevel))/8.0)			
		}
		//Record entry entities for calculating profit/loss and stoploss later.
		ts.EntryPrice = append(ts.EntryPrice, ts.CurrentPrice)
		ts.EntryQuantity = append(ts.EntryQuantity, orderResp.ExecutedQty)
		ts.EntryCostLoss = append(ts.EntryCostLoss, (orderResp.Commission * averagePrice))
		nextProfitSeLLPrice := ((mdTargetProfit + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / orderResp.ExecutedQty) + ts.EntryPrice[len(ts.EntryPrice)-1]
		nextInvBuYPrice := (-(md.TargetStopLoss + ts.EntryCostLoss[len(ts.EntryCostLoss)-1]) / orderResp.ExecutedQty) + ts.EntryPrice[len(ts.EntryPrice)-1]
		commissionAtProfitSeLLPrice := nextProfitSeLLPrice * orderResp.ExecutedQty * ts.CommissionPercentage
		commissionAtInvBuYPrice := nextInvBuYPrice * orderResp.ExecutedQty * ts.CommissionPercentage
		ts.NextProfitSeLLPrice = append(ts.NextProfitSeLLPrice, nextProfitSeLLPrice+commissionAtProfitSeLLPrice)
		ts.NextInvestBuYPrice = append(ts.NextInvestBuYPrice, nextInvBuYPrice-commissionAtInvBuYPrice)

		ts.TradingLevel = len(ts.EntryPrice)

		// Mark that we are in a trade.
		ts.InTrade = true

		resp := fmt.Sprintf("- BUY at EntryPrice[%d]: %.8f, EntryQuantity[%d]: %.8f, QBal: %.8f, BBal: %.8f, BuyCommission: %.8f PosPcent: %.8f, \nGlobalP&L %.2f, nextProfitSeLLPrice %.8f nextInvBuYPrice %.8f TargProfit %.8f TargLoss %.8f, tsDataPt: %d, mdDataPt: %d \n",
			len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, orderResp.ExecutedQty, ts.QuoteBalance, ts.BaseBalance, orderResp.Commission, md.RiskPositionPercentage, md.TotalProfitLoss, nextProfitSeLLPrice, nextInvBuYPrice, md.TargetProfit, -md.TargetStopLoss, ts.DataPoint, md.DataPoint)
		return resp, nil
	case "Sell":
		if !ts.InTrade {
			return "", fmt.Errorf("cannot execute a sell order without an existing trade")
		}

		// Calculate profit/loss for the trade.
		exitPrice := ts.CurrentPrice
		// adjustedPrice := math.Floor(price/lotSizeStep) * lotSizeStep
		quantity := math.Floor(ts.EntryQuantity[len(ts.EntryQuantity)-1]/ts.MiniQty) * ts.MiniQty

		// (localProfitLoss < transactionCost+slippageCost+md.TargetProfit)
		if (ts.BaseBalance < quantity) || (exitPrice < ts.NextProfitSeLLPrice[len(ts.NextProfitSeLLPrice)-1]) {
			if ts.BaseBalance < quantity {
				quantity = math.Floor(ts.BaseBalance/ts.MiniQty) * ts.MiniQty
				if quantity < ts.MiniQty {
					return "", fmt.Errorf("cannot execute a sell order due to insufficient BaseBalance: %.8f miniQuantity required: %.8f", ts.QuoteBalance, ts.MiniQty)
				} else {
					return "", fmt.Errorf("cannot execute a sell order insufficient BaseBalance: %.8f needed up to: %.8f", ts.BaseBalance, quantity)
				}
			} else {
				i := len(ts.NextProfitSeLLPrice) - 1
				ts.Log.Printf("TA Signalled: SeLL, currentPrice: %.8f, NextProfitSeLLPrice[%d]: %.8f, Target Profit: %.8f", exitPrice, i, ts.NextProfitSeLLPrice[i], md.TargetProfit)
				return "", fmt.Errorf("cannot execute a sell order at: %.8f, expecting NextProfitSeLLPrice[%d]: %.8f, target profit: %.8f", exitPrice, i, ts.NextProfitSeLLPrice[i], md.TargetProfit)
			}
		}
		// Calculate the total cost of the trade
		totalCost := quantity * exitPrice
		// Check if the total cost meets the minNotional requirement
		if totalCost < ts.MinNotional {
			ts.Log.Printf("Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, exitPrice, totalCost, ts.MinNotional)
			return "", fmt.Errorf("Not placing trade for %s: Quantity=%.4f, Price=%.2f, Total=%.2f does not meet MinNotional=%.2f\n", ts.Symbol, quantity, exitPrice, totalCost, ts.MinNotional)
		}

		//Placing Order
		orderResp, err := ts.APIServices.PlaceLimitOrder(ts.Symbol, "SELL", exitPrice, quantity)
		if err != nil {
			fmt.Println("Error placing exit order:", err)
			ts.Log.Println("Error placing exit order:", err)
			return "", fmt.Errorf("Error placing exit order: %v", err)
		}
		averagePrice := exitPrice
		if orderResp.ExecutedPrice > exitPrice{
			averagePrice = orderResp.ExecutedPrice
		}
		if orderResp.ExecutedQty <= 0.0{
			orderResp.ExecutedQty = quantity
		}
		//For a sell order:
		//You subtract the commission from the total cost because you are receiving less due to the commission fee. So, the formula would
		//be something like: totalCost = (executedQty * averagePrice) - totalCommission.
		//Confirm Binance or the Exch is deliverying commission for Sell in USDT
		totalCost = (orderResp.ExecutedQty * averagePrice)

		ts.Log.Printf("Exit order placed with ID: %d Commission: %.8f CumulativeQuoteQty: %.8f ExecutedPrice: %.8f ExecutedQty: %.8f Status: %s\n", orderResp.OrderID, orderResp.Commission, orderResp.CumulativeQuoteQty,
			orderResp.ExecutedPrice, orderResp.ExecutedQty, orderResp.Status)
		// Update the totalP&L, quote and base balances after the trade.
		ts.QuoteBalance += totalCost
		ts.BaseBalance -= orderResp.ExecutedQty
		localProfitLoss := CalculateProfitLoss(ts.EntryPrice[len(ts.EntryPrice)-1], averagePrice, orderResp.ExecutedQty)
		// ts.Log.Printf("Profit Before Global: %v, Local: %v\n",md.TotalProfitLoss, localProfitLoss)
		md.TotalProfitLoss += localProfitLoss
		// ts.Log.Printf("Profit After Global: %v, Local: %v\n",md.TotalProfitLoss, localProfitLoss)
		if localProfitLoss > 0 {
			ts.ClosedWinTrades += 2
		}
		// Mark that we are no longer in a trade.
		ts.InTrade = false

		resp := fmt.Sprintf("- SELL at ExitPrice: %.8f, EntryPrice[%d]: %.8f, EntryQuantity[%d]: %.8f, QBal: %.8f, BBal: %.8f, \nGlobalP&L: %.2f SellCommission: %.8f PosPcent: %.8f tsDataPt: %d mdDataPt: %d \n",
			exitPrice, len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, orderResp.Commission, md.RiskPositionPercentage, ts.DataPoint, md.DataPoint)

		if len(ts.EntryPrice) > 1 {
			ts.EntryPrice = ts.EntryPrice[:len(ts.EntryPrice)-1]
			ts.EntryCostLoss = ts.EntryCostLoss[:len(ts.EntryCostLoss)-1]
			ts.EntryQuantity = ts.EntryQuantity[:len(ts.EntryQuantity)-1]
			ts.StopLossRecover = ts.StopLossRecover[:len(ts.StopLossRecover)-1]
			ts.NextProfitSeLLPrice = ts.NextProfitSeLLPrice[:len(ts.NextProfitSeLLPrice)-1]
			ts.NextInvestBuYPrice = ts.NextInvestBuYPrice[:len(ts.NextInvestBuYPrice)-1]
			md.RiskPositionPercentage /= ts.RiskFactor //reduce riskPosition by a factor
			ts.InTrade = true
			ts.StopLossTrigered = true
			ts.TradingLevel = len(ts.EntryPrice)
		} else if len(ts.EntryPrice) <= 1 {
			// Mark that we are no longer in a trade.
			ts.InTrade = false
			ts.StopLossTrigered = false
			ts.StopLossRecover = []float64{math.MaxFloat64}
			ts.EntryPrice = []float64{}
			ts.EntryCostLoss = []float64{}
			ts.EntryQuantity = []float64{}
			ts.NextProfitSeLLPrice = []float64{}
			ts.NextInvestBuYPrice = []float64{}
			ts.TradingLevel = 0
		}

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

	ts.RiskCost = math.Floor((ts.MinNotional+1.0)/ts.StepSize) * ts.StepSize
 
	switch ts.TradingLevel {
	case 0:
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 1:
		ts.RiskCost += 5.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 2:
		ts.RiskCost += 10.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 3:
		ts.RiskCost += 15.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 4:
		ts.RiskCost += 20.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 5:
		ts.RiskCost += 25.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 6:
		ts.RiskCost += 30.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 7:
		ts.RiskCost += 35.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 8:
		ts.RiskCost += 40.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	case 9:
		ts.RiskCost += 45.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	default:
		ts.RiskCost = 100.0
		ts.PositionSize = ts.RiskCost / ts.CurrentPrice
	}

	// if this function was call for Buy Return to the caller at this point
	if !ts.InTrade {
		return "false"
	}
	//if you got here then this function was call for sell in which there is an entryPrice to be exited

	// Calculate profit/loss for the trade.
	exitPrice := ts.CurrentPrice
	// Check if the current price breaches the stop-loss level to trigger a buy signal.
	i := len(ts.NextInvestBuYPrice) - 1
	if (exitPrice < ts.NextInvestBuYPrice[i]) && ts.EnableStoploss { //&& (ts.BaseBalance >= quantity)
		// I commented it out to make stoploss just a marker
		// // Update the quote and base balances after the trade.
		// ts.QuoteBalance -= (quantity * exitPrice) + transactionCost + slippageCost
		// ts.BaseBalance += quantity
		// md.TotalProfitLoss += localProfitLoss

		// Mark that stoploss is triggered.
		ts.InTrade = false
		ts.StopLossTrigered = true
		md.RiskPositionPercentage *= ts.RiskFactor
		ts.TradingLevel = len(ts.EntryPrice)
		ts.StopLossRecover = append(ts.StopLossRecover, ts.CurrentPrice) //* (1.0 - ts.RiskStopLossPercentage)

		ts.Log.Printf("Stoploss Marked!!! For L%d demarc: at CurrentPrice: %.8f, of EntryPrice[%d]: %.8f, NextInvestBuYPrice[%d]: %.8f Target StopLoss: %.8f",
			len(ts.StopLossRecover), exitPrice, len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], i, ts.NextInvestBuYPrice[i], -md.TargetStopLoss)
		// I commented it out to make stoploss just a marker
		resp := fmt.Sprintf("Stoploss Marked!!! For L%d demarc: at CurrentPrice: %.8f NextInvestBuYPrice[%d]: %.8f Target StopLoss: %.8f", len(ts.StopLossRecover), exitPrice, i, ts.NextInvestBuYPrice[i], -md.TargetStopLoss)
		// 	ts.CurrentPrice, quantity, ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, localProfitLoss, md.RiskPositionPercentage, ts.DataPoint, md.DataPoint)
		return resp
	} else {
		ts.Log.Printf("Stoploss NOT Marked at CurrentPrice: %.8f, of EntryPrice[%d]: %.8f, NextInvestBuYPrice[%d]: %.8f Target StopLoss: %.8f",
			exitPrice, len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], i, ts.NextInvestBuYPrice[i], -md.TargetStopLoss)
	}
	return "false"
}

// TechnicalAnalysis(): This function performs technical analysis using the
// calculated moving averages (short and long EMA), RSI, MACD line, and Bollinger
// Bands. It determines the buy and sell signals based on various strategy rules.
func (ts *TradingSystem) TechnicalAnalysis(md *model.AppData, Action string) (buySignal, sellSignal bool) {
	// Calculate moving averages (MA) using historical data.
	ch := make(chan string)
	var(
		err1, err2 error
		shortEMA, longEMA []float64
	) 
	go func (ch chan string)  {
		longEMA, err2 = CandleExponentialMovingAverageV2(CandleExponentialMovingAverageV1(ts.ClosingPrices, 6), md.LongPeriod)
		ch <- ""
	}(ch)
	shortEMA, err1 = CandleExponentialMovingAverageV2(CandleExponentialMovingAverageV1(ts.ClosingPrices, 8), md.ShortPeriod)
	<-ch

	// longEMA, period4EMA, err := CandleExponentialMovingAverage(ts.ClosingPrices, md.LongPeriod, 4)
	// shortEMA, err := CandleExponentialMovingAverageV1(period4EMA, md.ShortPeriod)
	// period3EMA, err := CandleExponentialMovingAverageV1(ts.ClosingPrices, 3)
	// longEMA, shortEMA, err := CandleExponentialMovingAverage(period3EMA, md.LongPeriod, md.ShortPeriod)
	if (err1 != nil) || (err2 != nil) {
		// log.Printf("Error: in TechnicalAnalysis Unable to get EMA: %v", err)
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

			if buySignal && (Action == "Entry") && (len(ts.NextInvestBuYPrice) >= 1) {
				i := len(ts.NextInvestBuYPrice) - 1
				ts.Log.Printf("TA Signalled: BuY: %v at currentPrice: %.8f, will BuY below NextInvestBuYPrice[%d]: %.8f, and StopLossRecover[%d]: %.8f, Target Stoploss: %.8f", buySignal, ts.CurrentPrice, i, ts.NextInvestBuYPrice[i], i, ts.StopLossRecover[i], -md.TargetStopLoss)
			} else if sellSignal && (Action == "Exit") && (len(ts.NextProfitSeLLPrice) >= 1) {
				i := len(ts.NextProfitSeLLPrice) - 1
				ts.Log.Printf("TA Signalled: SeLL, currentPrice: %.8f, will SeLL above NextProfitSeLLPrice[%d]: %.8f, and StopLossRecover[%d]: %.8f, Target Profit: %.8f", ts.CurrentPrice, i, ts.NextProfitSeLLPrice[i], i, ts.StopLossRecover[i], md.TargetProfit)
			} else if buySignal && (Action == "Entry") {
				ts.Log.Printf("TA Signalled: BuY, at currentPrice: %.8f", ts.CurrentPrice)
			} else if sellSignal && (Action == "Exit") {
				ts.Log.Printf("TA Signalled: SeLL, currentPrice: %.8f", ts.CurrentPrice)
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

func (ts *TradingSystem) Reporting(md *model.AppData, from string) error {
	var err error

	if (len(ts.Container1) > 0) && (len(ts.Container1) <= md.LongPeriod) {
		err = ts.CreateLineChartWithSignals(md, ts.Timestamps, ts.ClosingPrices, ts.Signals, "")
	} else if len(ts.Container1) <= ts.Zoom {
		err = ts.CreateLineChartWithSignalsV3(md, ts.Timestamps, ts.ClosingPrices, ts.Container1, ts.Container2, ts.Signals, "")
	} else {
		b := len(ts.ClosingPrices) - 1
		a := len(ts.ClosingPrices) - ts.Zoom
		err = ts.CreateLineChartWithSignalsV3(md, ts.Timestamps[a:b], ts.ClosingPrices[a:b], ts.Container1[a:b], ts.Container2[a:b], ts.Signals[a:b], "")
	}
	if err != nil {
		return fmt.Errorf("Error creating Line Chart with signals: %v", err)
	}
	if from == "Live Trading" {
		// err = ts.AppDatatoCSV(md)
		// if err != nil {
		// 	return fmt.Errorf("Error creating Writing to CSV File %v", err)
		// }
	}
	return nil
}

func (ts *TradingSystem) ShutDown(md *model.AppData, sigchnl chan os.Signal) {
	for {
		select {
		case sig := <-sigchnl:
			if sig == syscall.SIGINT {
				log.Printf("Received signal: %v. Exiting...\n", sig)
				//Check if there is still asset remainning and sell off
				if ts.BaseBalance > 0.0 {
					// After sell off Update the quote and base balances after the trade.

					// Calculate profit/loss for the trade.
					exitPrice := ts.CurrentPrice
					localProfitLoss := CalculateProfitLoss(ts.EntryPrice[len(ts.EntryPrice)-1], exitPrice, ts.BaseBalance)
					transactionCost := ts.CommissionPercentage * exitPrice * ts.BaseBalance

					// Store profit/loss for the trade.

					localProfitLoss -= transactionCost
					// md.TotalProfitLoss += localProfitLoss

					ts.Mu.Lock()
					ts.QuoteBalance += (ts.BaseBalance * exitPrice) - transactionCost
					ts.BaseBalance -= ts.BaseBalance
					ts.Signals = append(ts.Signals, "Sell")
					ts.InTrade = false
					ts.TradeCount++
					ts.Mu.Unlock()
					log.Printf("- SELL-OFF at EntryPrice[%d]: %.8f, EntryQuantity[%d]: %.8f, QBal: %.8f, BBal: %.8f, GlobalP&L %.2f LocalP&L: %.8f PosPcent: %.8f tsDataPt: %d mdDataPt: %d \n",
						len(ts.EntryPrice)-1, ts.EntryPrice[len(ts.EntryPrice)-1], len(ts.EntryQuantity)-1, ts.EntryQuantity[len(ts.EntryQuantity)-1], ts.QuoteBalance, ts.BaseBalance, md.TotalProfitLoss, localProfitLoss, md.RiskPositionPercentage, ts.DataPoint, md.DataPoint)
				}
				// Print the overall trading performance after backtesting.
				log.Printf("Summary: Strategy: %s ", md.Strategy)
				log.Printf("Total Trades: %d, out of %d trials ", ts.TradeCount, len(ts.Signals))
				log.Printf("Total Profit/Loss: %.2f, ", md.TotalProfitLoss)
				log.Printf("Final Capital: %.2f, ", ts.QuoteBalance)
				log.Printf("Final Asset: %.8f tsDataPoint: %d, mdDataPoint: %d\n\n", ts.BaseBalance, ts.DataPoint, md.DataPoint)
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

	longEMA, shortEMA, err := CandleExponentialMovingAverage(closingPrices, LongPeriod, ShortPeriod)
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
func CandleExponentialMovingAverage(closingPrices []float64, LongPeriod, ShortPeriod int) (longEMA, shortEMA []float64, err error) {
	if LongPeriod <= 0 || len(closingPrices) < LongPeriod || closingPrices == nil{
		return nil, nil, fmt.Errorf("Error Calculating Candle EMA: not enoguh data for period %v", LongPeriod)
	}
	var ema55, ema15 *ema.Ema
	ema55 = ema.NewEma(alphaFromN(LongPeriod))
	ema15 = ema.NewEma(alphaFromN(ShortPeriod))

	longEMA = make([]float64, len(closingPrices))
	shortEMA = make([]float64, len(closingPrices))
	for k, closePrice := range closingPrices {
		ema55.Step(closePrice)
		ema15.Step(closePrice)
		longEMA[k] = ema55.Compute()
		shortEMA[k] = ema15.Compute()
	}
	return longEMA, shortEMA, nil
}

//CandleExponentialMovingAverage calculates EMA from condles
func CandleExponentialMovingAverageV2(closingPrices []float64, ShortPeriod int) (shortEMA []float64, err error) {
	if ShortPeriod <= 0 || len(closingPrices) < ShortPeriod || closingPrices == nil{
		return nil, fmt.Errorf("Error Calculating Candle EMA: not enoguh data for period %v", ShortPeriod)
	}
	var ema15 *ema.Ema
	ema15 = ema.NewEma(alphaFromN(ShortPeriod))

	shortEMA = make([]float64, len(closingPrices))
	for k, closePrice := range closingPrices {
		ema15.Step(closePrice)
		shortEMA[k] = ema15.Compute()
	}
	return shortEMA, nil
}
//CandleExponentialMovingAverage calculates EMA from condles
func CandleExponentialMovingAverageV1(closingPrices []float64, ShortPeriod int) (shortEMA []float64) {
	if ShortPeriod <= 0 || len(closingPrices) < ShortPeriod || closingPrices == nil{
		return nil
	}
	var ema15 *ema.Ema
	ema15 = ema.NewEma(alphaFromN(ShortPeriod))

	shortEMA = make([]float64, len(closingPrices))
	for k, closePrice := range closingPrices {
		ema15.Step(closePrice)
		shortEMA[k] = ema15.Compute()
	}
	return shortEMA
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
