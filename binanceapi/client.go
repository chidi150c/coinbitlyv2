package binanceapi

import (
	"errors"
	"fmt"
	"log"

	"coinbitly.com/config"
	"coinbitly.com/helper"
	"coinbitly.com/model"
)

type APIServices struct {
	*config.ExchConfig
}

func NewAPIServices(exchConfig *config.ExchConfig) (*APIServices, error) {
	// Check if the environment variables are set
	if exchConfig.ApiKey == "" || exchConfig.SecretKey == "" {
		fmt.Println("Error: Binance API credentials not set.")
		return nil, errors.New("Error: Binance API credentials not set.")
	}
	return &APIServices{exchConfig}, nil
}

// FetchHistoricalCandlesticks fetches historical candlestick data for the given symbol and time interval
func (e *APIServices) FetchCandles(symbol, interval string, startTime, endTime int64) ([]model.Candle, error) {
	candles, err := fetchHistoricalCandlesticks(symbol, e.BaseURL, e.ApiVersion, e.ApiKey, interval, startTime, endTime)
	if err != nil {
		fmt.Println("Error fetching Candle data:", err)
		return []model.Candle{}, err
	}
	mcandles := make([]model.Candle, 0, len(candles))
	ct := model.Candle{}

	for _, v := range candles {
		// fmt.Println(v)
		ct = model.Candle{
			ExchName:  e.Name,
			Timestamp: v.Timestamp,
			Open:      v.Open,
			High:      v.High,
			Low:       v.Low,
			Close:     v.Close,
			Volume:    v.Volume,
		}
		mcandles = append(mcandles, ct)
	}
	fmt.Println()
	// fmt.Println("Candles fetched =", mcandles, " = ", len(mcandles), "counts")
	fmt.Println()
	return mcandles, nil
}
func (e *APIServices) FetchTicker(symbol string) (CurrentPrice float64, err error) {
	ticker, err := fetchTickerData(symbol, e.BaseURL, e.ApiVersion, e.ApiKey)
	if err != nil {
		fmt.Println("Error fetching ticker data:", err)
		return 0.0, err
	}
	// Display the ticker data
	// fmt.Printf("Symbol: %s\nPrice: %s\n", ticker.Symbol, ticker.Price)
	return helper.ParseStringToFloat(ticker.Price), nil
}

func (e *APIServices) FetchExchangeEntities(symbol string) (minQty, maxQty, stepSize, minNotional float64, err error) {
	exchangeInfo, err := fetchExchangeInfo(symbol, e.BaseURL, e.ApiVersion, e.ApiKey)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// Find the trading pair in the response
	found := false
	for _, symbolInfo := range exchangeInfo["symbols"].([]interface{}) {
		symbolMap := symbolInfo.(map[string]interface{})
		if symbolMap["symbol"] == symbol {
			filters := symbolMap["filters"].([]interface{})
			for _, filter := range filters {
				filterMap := filter.(map[string]interface{})
				filterType := filterMap["filterType"].(string)

				switch filterType {
				case "NOTIONAL":
					minNotional = helper.ParseStringToFloat(filterMap["minNotional"].(string))
					fmt.Printf("MinNotional: %.8f\n", minNotional)
				case "LOT_SIZE":
					minQty = helper.ParseStringToFloat(filterMap["minQty"].(string))
					maxQty = helper.ParseStringToFloat(filterMap["maxQty"].(string))
					stepSize = helper.ParseStringToFloat(filterMap["stepSize"].(string))
					fmt.Printf("Symbol: %s\nMin Quantity: %.8f\nMax Quantity: %.8f\nStep Size: %.8f\n", symbol, minQty, maxQty, stepSize)
				}
			}
			found = true
			break
		}
	}
	if !found {
		return 0, 0, 0, 0, fmt.Errorf("Symbol not found: %s", symbol)
	}
	return minQty, maxQty, stepSize, minNotional, nil
}

func (e *APIServices) PlaceLimitOrder(symbol, side string, price, quantity float64) (orderResp model.Response, err error) {
	// return model.Response{
	// 	OrderID:            int64(23211),
	// 	ExecutedQty:        quantity,
	// 	ExecutedPrice:      price,
	// 	Commission:         0.00075 * quantity,
	// 	CumulativeQuoteQty: quantity * price,
	// 	Status:             "",
	// }, err
	orderType := "LIMIT"
	timeInForce := "GTC"
	Price := fmt.Sprintf("%.8f", price)
	Quantity := fmt.Sprintf("%.8f", quantity)
	response, err := placeOrder(symbol, side, orderType, timeInForce, Price, Quantity, e.BaseURL, e.ApiVersion, e.ApiKey, e.SecretKey)
	if err != nil {
		log.Fatal(err)
	}
	// Convert relevant fields to appropriate numeric types
	orderResp.ExecutedQty = helper.ParseStringToFloat(response.ExecutedQty)
	orderResp.ExecutedPrice = helper.ParseStringToFloat(response.Price)
	if len(response.Fills) > 0 {
		// Access the first fill
		orderResp.Commission = helper.ParseStringToFloat(response.Fills[0].Commission)
	}
	// else {
	// 	orderResp.Commission = (orderResp.ExecutedQty * orderResp.ExecutedPrice) * 0.0009
	// }
	orderResp.OrderID = int64(response.OrderID)
	orderResp.CumulativeQuoteQty = helper.ParseStringToFloat(response.CumulativeQuoteQty)
	return orderResp, nil
}

func (e *APIServices) GetQuoteAndBaseBalances(symbol string) (quoteBalance float64, baseBalance float64, err error) {
	quoteBalance, baseBalance, err = getQuoteAndBaseBalances(e.ApiKey, symbol)
	if err != nil {
		fmt.Println("Error fetching balances:", err)
		return 0.0, 0.0, err
	}

	fmt.Printf("Quote Balance for %s: %.8f\n", symbol, quoteBalance)
	fmt.Printf("Base Balance for %s: %.8f\n", symbol, baseBalance)

	return quoteBalance, baseBalance, nil
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
