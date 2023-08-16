package influxdb

import (
	"testing"

	"coinbitly.com/config"
	// "coinbitly.com/model"
)

func TestWriter(t *testing.T){	
	//count,Strategy,StrategyCombLogic,ShortPeriod,LongPeriod,ShortMACDPeriod,LongMACDPeriod,
	//SignalMACDPeriod,RSIPeriod,StochRSIPeriod,SmoothK,SmoothD,RSIOverbought,RSIOversold,
	//StRSIOverbought,StRSIOversold,BollingerPeriod,BollingerNumStdDev,TargetProfit,
	//TargetStopLoss,RiskPositionPercentage,Scalping,
	// testkit := []model.BackT{  //StochRSI
		// {1,"MACD StrategiesOnly","",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {2,"RSI StrategyOnly","",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {3,"EMA DataStrategies","",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {4,"StochR StrategiesOnly","",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {5,"Bollinger DataStrategies","",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {6,"MACD,RSI","AND",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {7,"RSI,MACD","AND",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {8,"EMA,MACD","AND",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {9,"StochR,MACD","AND",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {10,"Bollinger,MACD","AND",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {11,"MACD,RSI","OR",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {12,"RSI,MACD","OR",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {13,"EMA,MACD","OR",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {14,"StochR,MACD","OR",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
		// {15,"Bollinger,MACD","OR",6,16,12,29,9,14,14,3,3,0.6,0.4,0.7,0.2,20,2.0,1.0,0.5,0.1,"",0.0},
	// }

	// if err := WriteToInfluxDB(testkit[0]); err != nil{
	// 	t.Errorf("Error writing to influxDB: %v", err)
	// }
	ic,_ := NewAPIServices(config.NewExchangesConfig()["InfluxDB"])
	_,_ = ic.CandlesFromInfluxDB()
}