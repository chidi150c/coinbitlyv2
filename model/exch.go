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

type APIServices interface {
	FetchCandles(symbol, interval string, startTime, endTime int64) ([]Candle, error)
	WriteCandleToDB(ClosePrice float64, Timestamp int64)error
}
