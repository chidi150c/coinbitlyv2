package model

import "time"

// PredictionResponse represents the structure of the JSON response from the Flask app
type PredictionResponse struct {
    Prediction int `json:"prediction"`
}

// DataPoint represents a single data point in the dataset
type DataPoint struct {
	Date            time.Time
	DiffL95S15      float64
	DiffL8S4        float64
	RoCL95          float64
	RoCS15          float64
	MA5DiffL95S15   float64
	MA5DiffL8S4     float64
	StdDevL95       float64
	StdDevS15       float64
	LaggedL95EMA    float64
	LaggedS15EMA    float64
	Label           int
	ProfitLoss float64
	Asset           float64
	QuoteBalance    float64
	BaseBalance     float64
	CurrentPrice    float64
	LowestPrice     float64
	HighestPrice    float64
	RSI float64
	OBV float64
	ATR float64
	MACD float64
	PSAR float64
}
	// CrossL95S15DN     bool
	// CrossL8S4UP     bool
	// CrossL8S4DN     bool