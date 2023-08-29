package model

// type AppID uint64

type AppData struct{
	DataPoint int
	Strategy string
	ShortPeriod  int   
	LongPeriod  int      
	ShortEMA float64
	LongEMA float64 
	TargetProfit float64
	TargetStopLoss float64
	RiskPositionPercentage float64
	TotalProfitLoss        float64
}

type ChartData struct{
	ClosingPrices float64 `json:"ClosingPrices"`
	Timestamps    int64   `json:"Timestamps"`
	Signals       string  `json:"Signals"`
	ShortEMA      float64 `json:"ShortEMA"`
	LongEMA       float64 `json:"LongEMA"`
}
type BacTServices interface{
	WriteToInfluxDB(backTData *AppData) error
}