package config

import (
	"fmt"
	"os"
	"time"
)

type Api struct {
}
type ExchConfig struct {
	Name                  string
	BaseURL               string
	ApiVersion            string
	ApiKey                string
	SecretKey             string
	Symbol                string
	BaseCurrency          string
	QuoteCurrency	      string
	CandleInterval        string
	Symbols               []string
	CandleStartTime       int64
	CandleEndTime         int64
	DBOrgID               string
	LiveDBBucket          string
	LiveTimeRange         string
	LiveMeasurement       string
	LiveTag               string
	HistoricalDBBucket    string
	HistoricalTimeRange   string
	HistoricalMeasurement string
	HistoricalTag         string
}

const (
	timeRange int = -1
)

func NewExchangesConfig() map[string]*ExchConfig {
	ExchConfigs := map[string]*ExchConfig{
		"InfluxDB": {
			Name:    "InfluxDB",
			Symbols: []string{"BTCUSDT"},
			BaseURL: "http://influxdb-container:8086",
			//"http://localhost:8086" => local
			//"http://influxdb-container:8086" => production
			DBOrgID:   "Resoledge",
			SecretKey: "8tf6oI1nYCHpeosYrw9qcB_31tL6w4k1l3EFq2olfCylSyTJBL3y6Db0bBgQIul9CBKZExtvLYZJe_XYDiNI7A==",
			//"aXDeT9-0EX6K81D_94L-6q5G-w2eHS_4FJTIbsanUNqHlziMrFTOD3JULdCkCWgCTtVPvIuBhxUB0asbt8_AYw==" => local
			//"8tf6oI1nYCHpeosYrw9qcB_31tL6w4k1l3EFq2olfCylSyTJBL3y6Db0bBgQIul9CBKZExtvLYZJe_XYDiNI7A==" => production
			Symbol:                "BTCUSDT",
			BaseCurrency: "BTC",
			QuoteCurrency: "USDT",
			CandleInterval:        "30m",
			CandleStartTime:       time.Now().Add(time.Duration(timeRange) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:         time.Now().Unix(),
			HistoricalMeasurement: "historical_data",
			HistoricalTag:         "candle",
			HistoricalTimeRange:   fmt.Sprintf("%dd", timeRange),
			HistoricalDBBucket:    "AIMLDataSet",
			LiveMeasurement:       "Live_data",
			LiveTag:               "ticker",
			LiveTimeRange:         fmt.Sprintf("%dd", timeRange),
			LiveDBBucket:          "AIMLLiveData",
		},
		"Binance": {
			Name:            "Binance",
			Symbols:         []string{"BTCUSDT"},
			BaseURL:         "https://api.binance.com",
			ApiVersion:      "api/v3",
			ApiKey:          os.Getenv("BINANCE_API_KEY"),    // Replace with your Binance API key
			SecretKey:       os.Getenv("BINANCE_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			BaseCurrency: "BTC",
			QuoteCurrency: "USDT",
			CandleInterval:  "1m",
			CandleStartTime: time.Now().Add(time.Duration(-1) * 20 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix(),  // 3 days ago
		},
		"HitBTC": {
			Name:            "HitBTC",
			BaseURL:         "https://api.hitbtc.com",
			ApiVersion:      "api/3",
			ApiKey:          os.Getenv("HITBTC_API_KEY"),    // Replace with your Binance API key
			SecretKey:       os.Getenv("HITBTC_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			BaseCurrency: "BTC",
			QuoteCurrency: "USDT",
			CandleInterval:  "m30",
			CandleStartTime: time.Now().Add(time.Duration(timeRange) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Unix(),
		},
	}
	return ExchConfigs
}

//testing git ignore
