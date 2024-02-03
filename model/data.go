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
	RoCL8          float64
	RoCS4          float64
	MA5DiffL95S15   float64
	MA5DiffL8S4     float64
	StdDevL95       float64
	StdDevS15       float64
	LaggedL95EMA    float64
	LaggedS15EMA    float64
	ProfitLoss float64
	CurrentPrice    float64
	StochRSI float64
	SmoothKRSI float64
	MACDLine float64
	MACDSigLine float64
	MACDHist float64
	OBV int64
	ATR float64
	Label           int
	Label2          int
}
	// CrossL95S15DN     bool
	// CrossL8S4UP     bool
	// CrossL8S4DN     bool