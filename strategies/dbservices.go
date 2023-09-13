package strategies

import (
	"encoding/json"
	"fmt"
	"strings"

	"coinbitly.com/model"
	"github.com/gorilla/websocket"
)

type RDBServices struct{
    
}

func NewRDBServices()*RDBServices{
    return &RDBServices{
    }
}

func(dbs *RDBServices)CreateDBTradingSystem(ts *TradingSystem) (tradeID uint, err error){	
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return 0, fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
	trade := model.TradingSystemData{
        Symbol:                   ts.Symbol,
        ClosingPrices:            ts.ClosingPrices[len(ts.ClosingPrices)-1],
        Timestamps:               ts.Timestamps[len(ts.Timestamps)-1],
        Signals:                  ts.Signals[len(ts.Signals)-1],
        CommissionPercentage:     ts.CommissionPercentage,
        InitialCapital:           ts.InitialCapital,
        PositionSize:             ts.PositionSize,
        InTrade:                  ts.InTrade,
        QuoteBalance:             ts.QuoteBalance,
        BaseBalance:              ts.BaseBalance,
        RiskCost:                 ts.RiskCost,
        DataPoint:                ts.DataPoint,
        CurrentPrice:             ts.CurrentPrice,
        TradeCount:               ts.TradeCount,
        EnableStoploss:           ts.EnableStoploss,
        StopLossTrigered:         ts.StopLossTrigered,
        StopLossRecover:          ts.StopLossRecover,
        RiskFactor:               ts.RiskFactor,
        MaxDataSize:              ts.MaxDataSize,
        RiskProfitLossPercentage: ts.RiskProfitLossPercentage,
        BaseCurrency:             ts.BaseCurrency,
        QuoteCurrency:            ts.QuoteCurrency,
        MiniQty:                  ts.MiniQty,
        MaxQty:                   ts.MaxQty,
        MinNotional:              ts.MinNotional,
        StepSize:                 ts.StepSize,
    }
    if len(ts.EntryPrice) >= 1 {
        trade.EntryCostLoss =  ts.EntryCostLoss
        trade.EntryQuantity =  ts.EntryQuantity
        trade.EntryPrice =     ts.EntryPrice
        trade.NextInvestBuYPrice = ts.NextInvestBuYPrice
        trade.NextProfitSeLLPrice = ts.NextProfitSeLLPrice
    }
    
	// Serialize the TradingSystemData object to JSON
	appDataJSON, err := json.Marshal(trade)
	if err != nil {
		return 0, fmt.Errorf("Error marshaling TradingSystemData to JSON: %v", err)
	}
	// Create a message (request) to send
	request := map[string]interface{}{
		"action": "create",
		"entity": "trading-system",
		"data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
	}   
	
    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return 0, fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return 0, fmt.Errorf("Failed to read WebSocket response: %v", err)
    }
    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return 0, fmt.Errorf("Invalid response format %v", ms)
    }
	// Perform assertions to verify the response
	uid, ok := response["data_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("Invalid response format %v ", uid)
	}
	fmt.Printf("%v with id: %v and ChanBuffer: %v\n", ms, uint(uid), len(ts.TSDataChan))
    if !strings.Contains(ms, "successfully"){
        return 0, fmt.Errorf("Something went wrong: %v", ms)
    }
	return uint(uid), err
}
func(dbs *RDBServices)ReadDBTradingSystem(tradeID uint) (ts *TradingSystem, err error){
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return nil, fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
    ap := model.TradingSystemData{ ID: tradeID}

    // Serialize the TradingSystemData object to JSON
    appDataJSON, err := json.Marshal(ap)
    if err != nil {
        return nil, fmt.Errorf("Error marshaling TradingSystemData to JSON: %v", err)
    }

    // Create a message (request) to send
    request := map[string]interface{}{
        "action": "read",
        "entity": "trading-system",
        "data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
    }

    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return nil, fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return nil, fmt.Errorf("Failed to read WebSocket response: %v", err)
    }

    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return nil, fmt.Errorf("Invalid response format %v", ms)
    }
    if !strings.Contains(ms, "successfully"){
        return nil, fmt.Errorf("Something went wrong: %v", ms)
    }
    var dbts model.TradingSystemData
    dataByte, _ := json.Marshal(response["data"])
    // Deserialize the WebSocket message directly into the struct
    if err := json.Unmarshal(dataByte, &dbts); err != nil {
        return nil, fmt.Errorf("Error parsing WebSocket message: %v", err)
    }

    //After reading data from the database in the model.TradingSystemData format 
    //it has to be converted to the TradingSystem format as follows:
    ts = &TradingSystem{}
    ts.ID = dbts.ID 
    ts.ClosingPrices = append(ts.ClosingPrices, dbts.ClosingPrices)
    ts.Timestamps = append(ts.Timestamps, dbts.Timestamps)
    ts.Signals = append(ts.Signals, dbts.Signals)
    ts.ClosingPrices = append(ts.ClosingPrices, dbts.ClosingPrices)
    ts.ClosingPrices = append(ts.ClosingPrices, dbts.ClosingPrices)      
    ts.NextInvestBuYPrice = dbts.NextInvestBuYPrice    
    ts.NextProfitSeLLPrice = dbts.NextProfitSeLLPrice 
    ts.EntryPrice = dbts.EntryPrice 
    ts.EntryQuantity = dbts.EntryQuantity 
    ts.EntryCostLoss = dbts.EntryCostLoss 
    ts.StopLossRecover = dbts.StopLossRecover
    ts.CommissionPercentage = dbts.CommissionPercentage
    ts.InitialCapital = dbts.InitialCapital
    ts.PositionSize   = dbts.PositionSize
    ts.InTrade      = dbts.InTrade
    ts.QuoteBalance   = dbts.QuoteBalance
    ts.BaseBalance    = dbts.BaseBalance
    ts.RiskCost       = dbts.RiskCost
    ts.DataPoint    = dbts.DataPoint
    ts.CurrentPrice   = dbts.CurrentPrice
    ts.TradeCount   = dbts.TradeCount
    ts.TradingLevel = dbts.TradingLevel
    ts.ClosedWinTrades  = dbts.ClosedWinTrades
    ts.EnableStoploss    = dbts.EnableStoploss
    ts.StopLossTrigered  = dbts.StopLossTrigered
    ts.RiskFactor     = dbts.RiskFactor
    ts.MaxDataSize      = dbts.MaxDataSize
    ts.RiskProfitLossPercentage = dbts.RiskProfitLossPercentage
    ts.BaseCurrency    = dbts.BaseCurrency
    ts.QuoteCurrency    = dbts.QuoteCurrency
    ts.MiniQty        = dbts.MiniQty
    ts.MaxQty         = dbts.MaxQty
    ts.MinNotional    = dbts.MinNotional
    ts.StepSize       = dbts.StepSize
	fmt.Printf("%v with id: %v\n", ms, ts.ID)
    return ts, nil
}
func(dbs *RDBServices)UpdateDBTradingSystem(trade *TradingSystem)(err error){
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
	
	// Serialize the TradingSystemData object to JSON
	appDataJSON, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("Error marshaling TradingSystemData to JSON: %v", err)
	}

	// Create a message (request) to send
	request := map[string]interface{}{
		"action": "update",
		"entity": "trading-system",
		"data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
	}
	
    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return fmt.Errorf("Failed to read WebSocket response: %v", err)
    }

    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return fmt.Errorf("Invalid response format %v", ms)
    }
	fmt.Printf("%v\n", ms)
	return nil
}
func(dbs *RDBServices)DeleteDBTradingSystem(tradeID uint) (err error){
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
	ap := model.AppData{ ID: tradeID}
	
	// Serialize the TradingSystemData object to JSON
	appDataJSON, err := json.Marshal(ap)
	if err != nil {
		return fmt.Errorf("Error marshaling TradingSystemData to JSON: %v", err)
	}

	// Create a message (request) to send
	request := map[string]interface{}{
		"action": "delete",
		"entity": "trading-system",
		"data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
	}
	
    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return fmt.Errorf("Failed to read WebSocket response: %v", err)
    }

    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return fmt.Errorf("Invalid response format %v and resp: %v", ms, response)
    }
	fmt.Printf("%v\n", ms)
	return err
}
func(dbs *RDBServices)CreateDBAppData(data *model.AppData) (id uint, err error){
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return 0, fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
	
	// Serialize the DBAppData object to JSON
	appDataJSON, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("Error marshaling DBAppData to JSON: %v", err)
	}

	// Create a message (request) to send
	request := map[string]interface{}{
		"action": "create",
		"entity": "app-data",
		"data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
	}	
    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return 0, fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return 0, fmt.Errorf("Failed to read WebSocket response: %v", err)
    }

    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return 0, fmt.Errorf("Invalid response format %v", ms)
    }
    // Perform assertions to verify the response
    uid, ok := response["data_id"].(float64)
    if !ok {
        return 0, fmt.Errorf("Invalid response format %v", id)
    }
	fmt.Printf("%v with id: %v\n", ms, uint(uid))
    if !strings.Contains(ms, "successfully"){
        return 0, fmt.Errorf("Something went wrong: %v", ms)
    }
	return uint(uid), nil
}
func(dbs *RDBServices)ReadDBAppData(dataID uint) (ad *model.AppData, err error){
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return nil, fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
	ap := model.AppData{ ID: dataID}
	
	// Serialize the DBAppData object to JSON
	appDataJSON, err := json.Marshal(ap)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling DBAppData to JSON: %v", err)
	}

	// Create a message (request) to send
	request := map[string]interface{}{
		"action": "read",
		"entity": "app-data",
		"data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
	}
	
    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return nil, fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return nil, fmt.Errorf("Failed to read WebSocket response: %v", err)
    }

    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return nil, fmt.Errorf("Invalid response format %v and resp: %v", ms, response)
    }
	fmt.Printf("tradrID in uint correct: %v\n", ms)
	return ad, err
}
func(dbs *RDBServices)UpdateDBAppData(data *model.AppData)(err error){
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
	
	// Serialize the DBAppData object to JSON
	appDataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error marshaling DBAppData to JSON: %v", err)
	}

	// Create a message (request) to send
	request := map[string]interface{}{
		"action": "update",
		"entity": "app-data",
		"data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
	}
	
    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return fmt.Errorf("Failed to read WebSocket response: %v", err)
    }

    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return fmt.Errorf("Invalid response format %v and resp: %v", ms, response)
    }
	fmt.Printf("tradrID in uint correct: %v\n", ms)
	return err
}
func(dbs *RDBServices)DeleteDBAppData(dataID uint) (err error){
    // Create a mock WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial("ws://my-database-app:35261/database-services/ws", nil)
    if err != nil {
        return fmt.Errorf("Failed to connect to WebSocket: %v", err)
    }
    defer conn.Close()
	ap := model.AppData{ ID: dataID}
	
	// Serialize the DBAppData object to JSON
	appDataJSON, err := json.Marshal(ap)
	if err != nil {
		return fmt.Errorf("Error marshaling DBAppData to JSON: %v", err)
	}

	// Create a message (request) to send
	request := map[string]interface{}{
		"action": "delete",
		"entity": "app-data",
		"data":   json.RawMessage(appDataJSON), // RawMessage to keep it as JSON
	}
	
    // Send the message
    err = conn.WriteJSON(request)
    if err != nil {
        return fmt.Errorf("Failed to send WebSocket message: %v", err)
    }

    // Receive and parse the response
    var response map[string]interface{}
    err = conn.ReadJSON(&response)
    if err != nil {
        return fmt.Errorf("Failed to read WebSocket response: %v", err)
    }

    // Perform assertions to verify the response
    ms, ok := response["message"].(string)
    if !ok {
        return fmt.Errorf("Invalid response format %v and resp: %v", ms, response)
    }
	fmt.Printf("tradrID in uint correct: %v\n", ms)
	return err
}