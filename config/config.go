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
				ApiKey: "1CHA2mXswJdHjfossO43t4WRa82HPzFaeZOt2entAgajkAYIUaf55f7CepLt58YK", // Replace with your Binance API key
				SecretKey: "0DiulGqjlOuQlQHcVQYLYjKfpkq6Qs2rNxNVHHTtF1s2uy7n5clugQRxTXltjqFj", // Replace with your Binance API secret key
				Symbol: "BTCUSDT",
				Interval: "30m",
			},
		},
		"HitBTC": { 
			Name: "HitBTC",
			Exch: Api{
				BaseURL: "https://api.hitbtc.com",
				ApiVersion: "api/3",
				ApiKey: "055a27622c79b892c197cdbf963ec641", // Replace with your Binance API key
				SecretKey: "ca2cc7c96ea711765c0f248c678a1863", // Replace with your Binance API secret key
				Symbol: "BTCUSDT",
				Interval: "m30",
			},
		},
	}
	return ExchConfigs
}