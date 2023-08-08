package hitbtcapi

import (
	"testing"
	"time"
	"fmt"
	"coinbitly.com/config"
)

func TestFetchHistoricalCandlesticks(t *testing.T){
	ticker, err := fetchHistoricalCandlesticks("BTCUSDT", "https://api.hitbtc.com", "api/3", "055a27622c79b892c197cdbf963ec641", "M15", time.Now().Add(-1 * time.Hour).Unix(), time.Now().Unix())
	if err != nil {
		t.Error("Error fetching Candle data:", ticker, err)
	}
}

func TestGetBalance(t *testing.T){
	var (
		baseURL string
		 apiVersion string
		  apiKey string
		   apiSecret string
	)
		exch := config.NewExchangesConfig()

		if val, ok := exch["HitBTC"]; ok {
			baseURL = val.BaseURL
			apiVersion = val.ApiVersion
			apiKey = val.ApiKey
			apiSecret = val.SecretKey
		}
		// Example: Placing a buy order for 0.01 BTC at the price of 40000 USD
		// symbol := "BTCUSD"
		// side := "buy"
		// quantity := 0.01
		// price := 40000.0

		if !checkBalance("USDT", 0.0, baseURL, apiVersion, apiKey, apiSecret){
			t.Error("Error:")
		}else{
			fmt.Println("Balance got USDT")
		}

		// bal, err := getBalance("USDT", baseURL, apiVersion, apiKey, apiSecret)
		// if err != nil {
		// 	t.Error("Error:", err)
		// 	return
		// }else{
		// 	fmt.Println("Balance got successfully: ", bal)
		// }

		order, err := createSpotOrder("BCTUSDT", "buy", 0.000001, 29074, baseURL, apiVersion, apiKey, apiSecret)
		if err != nil {
			t.Error("Error:", err)
			return
		}else{
			t.Log("Order placed successfully. Order ID: ", order.ID)
		}
	}	

