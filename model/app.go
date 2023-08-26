package model

// type AppID uint64

type AppData struct{
	DataPoint int
	Strategy string
	ShortPeriod  int   
	LongPeriod  int      
	ShortEMA float64
	LongEMA float64 
	ShortMACDPeriod  int     
	LongMACDPeriod  int 
	SignalMACDPeriod  int 
	RSIPeriod  int
	StochRSIPeriod     int
	SmoothK            int
	SmoothD            int
	RSIOverbought 	   float64 
	RSIOversold        float64 
	StRSIOverbought    float64 
	StRSIOversold      float64
	BollingerPeriod  int    
	BollingerNumStdDev float64
	TargetProfit float64
	TargetStopLoss float64
	RiskPositionPercentage float64
	TotalProfitLoss        float64
}

type BacTServices interface{
	WriteToInfluxDB(backTData *AppData) error
}