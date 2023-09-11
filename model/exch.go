package model

// Candlestick represents a single candlestick data
type Candle struct {
	ExchName  string  `json:"exchname"`
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}
type Response struct{
	OrderID int64
	ExecutedQty float64
	ExecutedPrice float64
	Commission float64
	CumulativeQuoteQty float64
	Status string
}

type APIServices interface {
	FetchCandles(symbol, interval string, startTime, endTime int64) ([]Candle, error)
	FetchTicker(symbol string) (float64, error)
	FetchExchangeEntities(symbol string) (minQty, maxQty, stepSize, minNotional float64, err error)
	PlaceLimitOrder(symbol, side string, price, quantity float64) (Response, error)
	GetQuoteAndBaseBalances(symbol string) (quoteBalance float64, baseBalance float64, err error) 
}

type DBServices interface{	
	FetchCandlesFromDB(symbol, interval string, startTime, endTime int64) ([]Candle, error)
	WriteCandleToDB(ClosePrice float64, Timestamp int64)error
	WriteTickerToDB(ClosePrice float64, Timestamp int64)error
	CloseDB()error
}
