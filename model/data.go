package model

import "time"

// DataPoint represents a single data point in the dataset
type DataPoint struct {
	Date            time.Time
	L95EMA          float64
	S15EMA          float64
	L8EMA           float64
	S4EMA           float64
	DiffL95S15      float64
	DiffL8S4        float64
	RoCL95          float64
	RoCS15          float64
	MA5DiffL95S15   float64
	MA5DiffL8S4     float64
	StdDevL95       float64
	StdDevS15       float64
	CrossL95S15     int
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
}
