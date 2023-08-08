package config

type Api struct{
	BaseURL string
	ApiVersion string
	ApiKey string
	SecretKey string
    Symbol string 
    Interval string 
}
type ExchConfig struct{
	Name string 
	Exch Api
	Symbols []string
}
func NewExchangesConfig()map[string]*ExchConfig{
	ExchConfigs := map[string]*ExchConfig{
		"Binance": {
			Name: "Binance",
			Symbols: []string{"BTCUSDT"},
			Exch: Api{
				BaseURL: "https://api.binance.com",
				ApiVersion: "api/v3",
				ApiKey: "", // Replace with your Binance API key
				SecretKey: "", // Replace with your Binance API secret key
				Symbol: "BTCUSDT",
				Interval: "30m",
			},
		},
		"HitBTC": { 
			Name: "HitBTC",
			Exch: Api{
				BaseURL: "https://api.hitbtc.com",
				ApiVersion: "api/3",
				ApiKey: "", // Replace with your Binance API key
				SecretKey: "", // Replace with your Binance API secret key
				Symbol: "BTCUSDT",
				Interval: "m30",
			},
		},
	}
	return ExchConfigs
}
