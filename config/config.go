package config

import (
	"os"
	"time"
)

type Api struct {
}
type ExchConfig struct {
	Name            string
	BaseURL         string
	ApiVersion      string
	ApiKey          string
	SecretKey       string
	Symbol          string
	CandleInterval  string
	Symbols         []string
	CandleStartTime int64
	CandleEndTime   int64
}

func NewExchangesConfig() map[string]*ExchConfig {
	ExchConfigs := map[string]*ExchConfig{
		"Binance": {
			Name:            "Binance",
			Symbols:         []string{"BTCUSDT"},
			BaseURL:         "https://api.binance.com",
			ApiVersion:      "api/v3",
			ApiKey:          os.Getenv("BINANCE_API_KEY"), // Replace with your Binance API key
			SecretKey:       os.Getenv("BINANCE_API_SECRET"),  // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			CandleInterval:  "30m",
			CandleStartTime: time.Now().Add(-10 * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Unix(),
		},
		"HitBTC": {
			Name: "HitBTC",
			BaseURL:    "https://api.hitbtc.com",
			ApiVersion: "api/3",
			ApiKey:  os.Getenv("HITBTC_API_KEY"), // Replace with your Binance API key
			SecretKey: os.Getenv("HITBTC_API_SECRET"), // Replace with your Binance API secret key
			Symbol:     "BTCUSDT",
			CandleInterval:   "m30",
			CandleStartTime: time.Now().Add(-10 * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Unix(),
		},
	}
	return ExchConfigs
}
