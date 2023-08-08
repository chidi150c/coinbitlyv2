package hitbtcapi

import (
	"errors"
	"fmt"

	"coinbitly.com/config"
	"coinbitly.com/helper"
	"coinbitly.com/model"
)

const (
	exchname = "HitBTC"
)
type ExchServices struct{
	*config.ExchConfig
}

func NewExchServices(exchConfig map[string]*config.ExchConfig)(*ExchServices, error){
	// Check if the key "HitBTC" exists in the map
	if val, ok := exchConfig["HitBTC"]; ok {	
		// Check if the environment variables are set
		if val.ApiKey == "" || val.SecretKey == "" {
			fmt.Println("Error: API credentials not set.")
			return nil, errors.New("Error: API credentials not set.")
		}	
		return &ExchServices{val}, nil
	} else {		
		fmt.Println("Error Internal: Exch name not HitBTC")
		return nil, errors.New("Exch name not HitBTC")
	}
}

// FetchHistoricalCandlesticks fetches historical candlestick data for the given symbol and time interval
func (e *ExchServices)FetchHistoricalCandlesticks(symbol, interval string, startTime, endTime int64) ([]model.Candlestick, error) {
	if e.Name != exchname{
		fmt.Println("Error Missmatch Exchange Name:")
		return []model.Candlestick{}, errors.New("Missmatch Exchange Name")		
	}
	ticker, err := fetchHistoricalCandlesticks(symbol, e.BaseURL, e.ApiVersion, e.ApiKey, interval, startTime, endTime)
	if err != nil {
		fmt.Println("Error fetching Candle data:", err)
		return []model.Candlestick{}, err
	}
	mticker := make([]model.Candlestick, 0, len(ticker))
	ct := model.Candlestick{}
	for _, v := range ticker{
		ct = model.Candlestick{
			ExchName: exchname,
			Timestamp: v.Timestamp.Unix(),
			Open: helper.ParseStringToFloat(v.Open),
			High: helper.ParseStringToFloat(v.High),
			Low: helper.ParseStringToFloat(v.Low),
			Close: helper.ParseStringToFloat(v.Close),
			Volume: helper.ParseStringToFloat(v.Volume), 
		}
		mticker = append(mticker, ct)
	}
	return mticker, nil
}

// // FetchTickerData fetches and displays real-time of a given symbol
// func (e *ExchServices)FetchTickerData(symbol string) (*model.TickerData, error) {
// 	ticker, err := fetchTickerData(symbol)
// 	if err != nil {
// 		fmt.Println("Error fetching ticker data:", err)
// 		return &TickerData{}, err
// 	}

// 	// Display the ticker data
// 	fmt.Printf("Symbol: %s\nPrice: %s\n", ticker.Symbol, ticker.Price)
// 	return ticker, nil
// }

// fetchAndDisplay24hrTickerData fetches and displays 24-hour price change statistics for the given symbol
// func (e *ExchServices)Fetch24hrChange(symbol string) (*model.Ticker24hrChange, error){
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
// func (e *ExchServices)FetchOrderBook(symbol string, limit int) (*model.OrderBookData, error) {
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

