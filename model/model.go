package model

// Candlestick represents a single candlestick data
type Candlestick struct {
	ExchName string `json:"exchname"`
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

type ExchServices interface{
	FetchHistoricalCandlesticks(symbol, interval string, startTime, endTime int64 )([]Candlestick, error)
}