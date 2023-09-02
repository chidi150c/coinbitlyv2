package config

import (
	"fmt"
	"os"
	"time"
)

type Api struct {
}
type DBConfig struct {
	Name                  string
	BaseURL               string
	ApiKey                string
	SecretKey             string
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
type ExchConfig struct {
	Name            string
	BaseURL         string
	ApiVersion      string
	ApiKey          string
	SecretKey       string
	Symbol          string
	BaseCurrency    string
	QuoteCurrency   string
	CandleInterval  string
	Symbols         []string
	CandleStartTime int64
	CandleEndTime   int64
	InitialCapital  float64
}

const (
	timeRange int = -1
)

func NewDataBaseConfigs() map[string]*DBConfig {
	DBConfigs := map[string]*DBConfig{
		"InfluxDB": {
			Name:                  "InfluxDB",
			BaseURL:               "http://influxdb-container:8086",
			DBOrgID:               "Resoledge",
			SecretKey:             "8tf6oI1nYCHpeosYrw9qcB_31tL6w4k1l3EFq2olfCylSyTJBL3y6Db0bBgQIul9CBKZExtvLYZJe_XYDiNI7A==",
			HistoricalMeasurement: "historical_data",
			HistoricalTag:         "candle",
			HistoricalTimeRange:   fmt.Sprintf("%dd", timeRange),
			HistoricalDBBucket:    "AIMLDataSet",
			LiveMeasurement:       "Live_data",
			LiveTag:               "ticker",
			LiveTimeRange:         fmt.Sprintf("%dd", timeRange),
			LiveDBBucket:          "AIMLLiveData",
		},
		"InfluxDBLocal": {
			Name:                  "InfluxDB",
			BaseURL:               "http://localhost:8086",
			DBOrgID:               "Resoledge",
			SecretKey:             "aXDeT9-0EX6K81D_94L-6q5G-w2eHS_4FJTIbsanUNqHlziMrFTOD3JULdCkCWgCTtVPvIuBhxUB0asbt8_AYw==",
			HistoricalMeasurement: "historical_data",
			HistoricalTag:         "candle",
			HistoricalTimeRange:   fmt.Sprintf("%dd", timeRange),
			HistoricalDBBucket:    "AIMLDataSet",
			LiveMeasurement:       "Live_data",
			LiveTag:               "ticker",
			LiveTimeRange:         fmt.Sprintf("%dd", timeRange),
			LiveDBBucket:          "AIMLLiveData",
		},
	}
	return DBConfigs
}
func NewExchangeConfigs() map[string]*ExchConfig {
	ExchConfigs := map[string]*ExchConfig{
		"Binance": {
			Name:            "Binance",
			Symbols:         []string{"BNBUSDT"},
			BaseURL:         "https://api.binance.com", //testnet.  https://testnet.binance.vision/api/v3/klines
			ApiVersion:      "api/v3",
			ApiKey:          os.Getenv("BINANCE_API_KEY"),    // Replace with your Binance API key
			SecretKey:       os.Getenv("BINANCE_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BNBUSDT",
			BaseCurrency:    "BNB",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.06239372,
			CandleInterval:  "1m",
			CandleStartTime: time.Now().Add(time.Duration(-1) * 20 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix(),  // 3 days ago
		},
		"BinanceTestnet": {
			Name:            "BinanceTestnet",
			Symbols:         []string{"BNBUSDT"},
			BaseURL:         "https://testnet.binance.vision",
			ApiVersion:      "api/v3",
			ApiKey:          "OHgORL3TN42JHFIwUQWzYDCWTaXJbHYHJA0YlAWvP84LusLLTFT80dX7kuWTpmGI", // os.Getenv("BINANCE_API_KEY"),    // Replace with your Binance API key
			SecretKey:       "ZX3J0nA5XOdMdP60ZieQlcnFuUbipMhTLmP2kufERZcZ8rawSVpDW1b5goWLYkb5", //os.Getenv("BINANCE_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BNBUSDT",
			BaseCurrency:    "BNB",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.06239372,
			CandleInterval:  "1m",
			CandleStartTime: time.Now().Add(time.Duration(-2) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix(),  // 3 days ago
		},
		"HitBTC": {
			Name:            "HitBTC",
			BaseURL:         "https://api.hitbtc.com",
			ApiVersion:      "api/3",
			ApiKey:          os.Getenv("HITBTC_API_KEY"),    // Replace with your Binance API key
			SecretKey:       os.Getenv("HITBTC_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BNBUSDT",
			BaseCurrency:    "BNB",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.06239372,
			CandleInterval:  "m30",
			CandleStartTime: time.Now().Add(time.Duration(timeRange) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Unix(),
		},
	}
	return ExchConfigs
}
