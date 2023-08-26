package binanceapi

import (
	"errors"
	"fmt"

	"coinbitly.com/config"
	"coinbitly.com/helper"
	"coinbitly.com/model"
)

const (
	exchname = "Binance"
)
type APIServices struct{
	*config.ExchConfig
	DataBase model.APIServices
}

func NewAPIServices(infDB model.APIServices, exchConfig *config.ExchConfig)(*APIServices, error){
	// Check if the environment variables are set
	if exchConfig.ApiKey == "" || exchConfig.SecretKey == "" {
		fmt.Println("Error: Binance API credentials not set.")
		return nil, errors.New("Error: Binance API credentials not set.")
	}
	return &APIServices{exchConfig, infDB}, nil
}
// FetchHistoricalCandlesticks fetches historical candlestick data for the given symbol and time interval
func (e *APIServices)FetchCandles(symbol, interval string, startTime, endTime int64) ([]model.Candle, error) {
	if e.Name != exchname{
		fmt.Println("Error Missmatch Exchange Name:")
		return []model.Candle{}, errors.New("Missmatch Exchange Name")		
	}
	candles, err := fetchHistoricalCandlesticks(symbol, e.BaseURL, e.ApiVersion, e.ApiKey, interval, startTime, endTime)
	if err != nil {
		fmt.Println("Error fetching Candle data:", err)
		return []model.Candle{}, err
	}
	mcandles := make([]model.Candle, 0, len(candles))
	ct := model.Candle{}

	for _, v := range candles{
		// fmt.Println(v)
		ct = model.Candle{
			ExchName: exchname,
			Timestamp: v.Timestamp,
			Open: v.Open,
			High: v.High,
			Low: v.Low,
			Close: v.Close,
			Volume: v.Volume, 
		}
		mcandles = append(mcandles, ct)
	}	
	fmt.Println()
	// fmt.Println("Candles fetched =", mcandles, " = ", len(mcandles), "counts")
	fmt.Println()
	return mcandles, nil
}
// func (e *APIServices)WriteCandleToDB(ClosePrice float64, Timestamp int64) error {
// 	return e.DataBase.WriteCandleToDB(ClosePrice, Timestamp)
// }
// func (e *APIServices)CloseDB()error{
// 	return e.DataBase.CloseDB()
// }
// FetchTicker fetches and displays real-time of a given symbol
func (e *APIServices)FetchTicker(symbol string)(CurrentPrice float64, err error){
		ticker, err := fetchTickerData(symbol, e.BaseURL, e.ApiVersion, e.ApiKey)
		if err != nil {
			fmt.Println("Error fetching ticker data:", err)
			return 0.0, err
		}
		// Display the ticker data
		// fmt.Printf("Symbol: %s\nPrice: %s\n", ticker.Symbol, ticker.Price)
		return helper.ParseStringToFloat(ticker.Price), nil
}
func (e *APIServices)WriteTickerToDB(ClosePrice float64, Timestamp int64)error{
	// _ = e.DataBase.WriteTickerToDB(ClosePrice, Timestamp)
	return nil
}

// fetchAndDisplay24hrTickerData fetches and displays 24-hour price change statistics for the given symbol
// func (e *APIServices)Fetch24hrChange(symbol string) (*model.Ticker24hrChange, error){
// 	ticker, err := fetch24hrTickerData(symbol, e.BaseURL, e.ApiVersion, e.ApiKey)
// 	if err != nil {
// 		fmt.Println("Error fetching 24-hour ticker data:", err)
// 		return &Ticker24hrChange{}, err
// 	}

// 	// Display the 24-hour price change statistics
// 	fmt.Printf("Symbol: %s\nLast Price: %s\nPrice Change: %s\nVolume: %s\n",
// 		ticker.Symbol, ticker.LastPrice, ticker.PriceChange, ticker.Volume)
// 		return ticker, nil
// }

// fetchAndDisplayOrderBook fetches and displays order book depth for the given symbol
// func (e *APIServices)FetchOrderBook(symbol string, limit int) (*model.OrderBookData, error) {
// 	orderBook, err := fetchOrderBook(symbol, e.BaseURL, e.ApiVersion, e.ApiKey, limit)
// 	if err != nil {
// 		fmt.Println("Error fetching order book data:", err)
// 		return &OrderBookData{}, err
// 	}

// 	// Display the order book data
// 	fmt.Printf("Symbol: %s\nTop %d Bids:\n", symbol, limit)
// 	for _, bid := range orderBook.Bids {
// 		fmt.Printf("Price: %s, Quantity: %s\n", bid[0], bid[1])
// 	}

// 	fmt.Printf("\nTop %d Asks:\n", limit)
// 	for _, ask := range orderBook.Asks {
// 		fmt.Printf("Price: %s, Quantity: %s\n", ask[0], ask[1])
// 	}
// 	return orderBook, nil
// }

