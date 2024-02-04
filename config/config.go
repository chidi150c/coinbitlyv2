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
	BaseCurrency    string
	QuoteCurrency   string
	CandleInterval  string
	Symbols         []string
	CandleStartTime int64
	CandleEndTime   int64
	InitialCapital  float64
}
type ModelConfig struct{	
    UserInput string
	ApiKey string
	Url string
	Model string
}
const (
	timeRange int = -1
)

func NewExchangeConfigs() map[string]*ExchConfig {
	ExchConfigs := map[string]*ExchConfig{
		"Binance": {
			Name:            "Binance",
			Symbols:         []string{"BTCUSDT"},
			BaseURL:         "https://api.binance.com", //testnet.  https://testnet.binance.vision/api/v3/klines
			ApiVersion:      "api/v3",
			ApiKey:          os.Getenv("BINANCE_API_KEY"),    // Replace with your Binance API key
			SecretKey:       os.Getenv("BINANCE_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			BaseCurrency:    "BTC",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.038193 + 26.47 + 54.2 + 86.5 + 100.0 + 16.6 + 58.0 + 56.72 + 18.0,
			CandleInterval:  "1m",
			CandleStartTime: time.Now().Add(time.Duration(-1) * 20 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix(),  // 3 days ago
		},
		"BinanceTestnet": {
			Name:            "BinanceTestnet",
			Symbols:         []string{"BTCUSDT"},
			BaseURL:         "https://testnet.binance.vision",
			ApiVersion:      "api/v3",
			ApiKey:          "Ob1mJrsNFb7Msfpb6eFwjc1IpQc3ivmGZdMaabxbttXwFipTlgASE6Zqjw2xqETZ", // os.Getenv("BINANCE_API_KEY"),    // Replace with your Binance API key
			SecretKey:       "tef6clCsc4zxTVqhGsD1neE7Od5CXEhEe8l8xScnyciA79OxcHpmIqcswLWOapfk", //os.Getenv("BINANCE_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			BaseCurrency:    "BTC",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.1,
			CandleInterval:  "1m",
			CandleStartTime: time.Now().Add(time.Duration(-3) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix(),  // 3 days ago
		},
		"BinanceTestnetWithDB": {
			Name:            "BinanceTestnetWithDB",
			Symbols:         []string{"BTCUSDT"},
			BaseURL:         "https://testnet.binance.vision",
			ApiVersion:      "api/v3",
			ApiKey:          "Ob1mJrsNFb7Msfpb6eFwjc1IpQc3ivmGZdMaabxbttXwFipTlgASE6Zqjw2xqETZ", // os.Getenv("BINANCE_API_KEY"),    // Replace with your Binance API key
			SecretKey:       "tef6clCsc4zxTVqhGsD1neE7Od5CXEhEe8l8xScnyciA79OxcHpmIqcswLWOapfk", //os.Getenv("BINANCE_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			BaseCurrency:    "BTC",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.1,
			CandleInterval:  "1m",
			CandleStartTime: time.Now().Add(time.Duration(-3) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix(),  // 3 days ago
		},
		"BinanceTestnetWithDBRemote": {
			Name:            "BinanceTestnetWithDBRemote",
			Symbols:         []string{"BTCUSDT"},
			BaseURL:         "https://testnet.binance.vision",
			ApiVersion:      "api/v3",
			ApiKey:          "Ob1mJrsNFb7Msfpb6eFwjc1IpQc3ivmGZdMaabxbttXwFipTlgASE6Zqjw2xqETZ", // os.Getenv("BINANCE_API_KEY"),    // Replace with your Binance API key
			SecretKey:       "tef6clCsc4zxTVqhGsD1neE7Od5CXEhEe8l8xScnyciA79OxcHpmIqcswLWOapfk", //os.Getenv("BINANCE_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			BaseCurrency:    "BTC",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.1,
			CandleInterval:  "1m",
			CandleStartTime: time.Now().Add(time.Duration(-3) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Add(time.Duration(-1) * 0 * time.Hour).Unix(),  // 3 days ago
		},
		"HitBTC": {
			Name:            "HitBTC",
			BaseURL:         "https://api.hitbtc.com",
			ApiVersion:      "api/3",
			ApiKey:          os.Getenv("HITBTC_API_KEY"),    // Replace with your Binance API key
			SecretKey:       os.Getenv("HITBTC_API_SECRET"), // Replace with your Binance API secret key
			Symbol:          "BTCUSDT",
			BaseCurrency:    "BTC",
			QuoteCurrency:   "USDT",
			InitialCapital:  54.1,
			CandleInterval:  "m30",
			CandleStartTime: time.Now().Add(time.Duration(timeRange) * 24 * time.Hour).Unix(), // 3 days ago
			CandleEndTime:   time.Now().Unix(),
		},
	}
	return ExchConfigs
}
func NewModelConfigs() map[string]ModelConfig{
	return map[string]ModelConfig{
		"gpt3": {
			UserInput : "your_topic_here", // Replace with your actual input
			ApiKey : os.Getenv("OPENAI_API_KEY"), // Ensure you have set your API key in your environment variables
			Url : "https://api.openai.com/v1/chat/completions",
			Model: "gpt-3.5-turbo",
		},
	}
}
