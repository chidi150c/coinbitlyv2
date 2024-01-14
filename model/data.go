package model

import "time"
// DataPointRequest represents the JSON structure for the POST request
type DataPointRequest struct {
    DiffL95S15      float64 `json:"DiffL95S15"`
    DiffL8S4        float64 `json:"DiffL8S4"`
    RoCL95          float64 `json:"RoCL95"`
    RoCS15          float64 `json:"RoCS15"`
    MA5DiffL95S15   float64 `json:"MA5DiffL95S15"`
    MA5DiffL8S4     float64 `json:"MA5DiffL8S4"`
    StdDevL95       float64 `json:"StdDevL95"`
    StdDevS15       float64 `json:"StdDevS15"`
    LaggedL95EMA    float64 `json:"LaggedL95EMA"`
    LaggedS15EMA    float64 `json:"LaggedS15EMA"`
    CurrentPrice    float64 `json:"CurrentPrice"`
}

// PredictionResponse represents the structure of the JSON response from the Flask app
type PredictionResponse struct {
    Prediction int `json:"prediction"`
}

// DataPoint represents a single data point in the dataset
type DataPoint struct {
	Date            time.Time
	CrossUPTime time.Time
	CrossL95S15UP     bool
	PriceDownGoingDown bool
	PriceDownGoingUp bool
	PriceUpGoingUp bool
	PriceUpGoingDown bool
	MarketDownGoingDown bool
	MarketDownGoingUp bool
	MarketUpGoingUp bool
	MarketUpGoingDown bool
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
	TotalProfitLoss float64
	Asset           float64
	QuoteBalance    float64
	BaseBalance     float64
	CurrentPrice    float64
	TargetProfit    float64
	TargetStopLoss  float64
	LowestPrice     float64
	HighestPrice    float64
	MarketUpGoneUp bool
	MarketDownGoneDown bool
	L95EMA6 float64
	L95EMA3 float64
	L95EMA1 float64
	L95EMA0 float64
	S15EMA6 float64
	S15EMA3 float64
	S15EMA1 float64
	S15EMA0 float64
}

	// CrossL95S15DN     bool
	// CrossL8S4UP     bool
	// CrossL8S4DN     bool