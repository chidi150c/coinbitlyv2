package strategies

import (
	"fmt"
	"log"
	"testing"
)

func TestBacktestStrategy(t *testing.T) {
    
	// Initialize the trading system.
	ts := &TradingSystem{
		ShortPeriod:        12, // Define moving average short period for the strategy.
		LongPeriod:         26, // Define moving average long period for the strategy.
	}

    err := ts.Initialize()
	if err != nil {
		log.Fatal("Error initializing trading system:", err)
		return
	}
    fmt.Println(ts.ClosingPrices)
	ts.Signals = make([]string, len(ts.ClosingPrices)) // Holder of generated trading signals

	ts.Strategy.UseBollinger = true
	
    //Test SMA
    sma := CalculateSimpleMovingAverage(ts.ClosingPrices, ts.LongPeriod)
    expected := 28000.0
    fmt.Println(sma)
    if sma[0] <= expected {
        t.Errorf("Expected SMA %f, but got %v\n", expected, sma[len(sma)-1])        
    }else{
        fmt.Printf("Test SMS Successfully: last : %f\n", sma[len(sma)-1])
    }

    //Test EMA
    ema := CalculateExponentialMovingAverage(ts.ClosingPrices, ts.LongPeriod)
    expected = 28000.0
    if ema[len(ema)-ts.LongPeriod] <= expected {
        t.Errorf("Expected EMA %f, but got %v\n", expected, ema[len(ema)-ts.LongPeriod])        
    }else{        
        fmt.Printf("Test EMA Successfully: last : %f\n", ema[len(ema)-ts.LongPeriod])
    }

    //Test CandleEMA
    longEMA, shortEMA, _, _ := ts.CandleExponentialMovingAverage()  

    if ema[len(ema)-ts.LongPeriod] > longEMA[len(longEMA)-1] {
        t.Errorf("Expected candleLongEMA %f, but got EMA: %v\n", longEMA[len(longEMA)-1], ema[len(ema)-ts.LongPeriod])        
    }else{        
        fmt.Printf("Test candleEMA Successfully: longEMA: %v, shortEMA : %f\n", longEMA[len(longEMA)-1], shortEMA[len(shortEMA)-1])
    }

    //MACD
    mac, sig, hist := ts.CalculateMACD()
    if true {
        t.Errorf("Got mac %v, sig %v, hist %v\n", mac[len(mac)-1], sig[len(sig)-1], hist[len(hist)-1])        
    }else{

    }   

}