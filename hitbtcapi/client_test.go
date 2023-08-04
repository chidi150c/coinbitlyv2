package hitbtcapi


import (
	"testing"
	"time"
	)

func TestFetchHistoricalCandlesticks(t *testing.T){
	ticker, err := fetchHistoricalCandlesticks("BTCUSDT", "https://api.hitbtc.com", "api/3", "055a27622c79b892c197cdbf963ec641", "M15", time.Now().Add(-1 * time.Hour).Unix(), time.Now().Unix())
	if err != nil {
		t.Error("Error fetching Candle data:", ticker, err)
	}
}

