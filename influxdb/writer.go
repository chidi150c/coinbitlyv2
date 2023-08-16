package influxdb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"coinbitly.com/model"
	"coinbitly.com/config"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type CandleServices struct{
	InfluxDBBucket string
	Client influxdb2.Client
	QueryAPI api.QueryAPI
	WriteAPI api.WriteAPIBlocking
	TimeRange string
	Measurement string
}

func NewAPIServices(configParam *config.ExchConfig) (*CandleServices, error){	
	// Create a new InfluxDB client 
	client := influxdb2.NewClientWithOptions(configParam.BaseURL, configParam.SecretKey, influxdb2.DefaultOptions())
	queryAPI := client.QueryAPI(configParam.DBOrgID)
	writeAPI := client.WriteAPIBlocking(configParam.DBOrgID, configParam.DBBucket)
	return &CandleServices{
		InfluxDBBucket: configParam.DBBucket,
		Client: client,
		QueryAPI: queryAPI,
		WriteAPI: writeAPI,
		TimeRange: configParam.TimeRange,
		Measurement: configParam.Measurement,
	}, nil
}

func (c *CandleServices)FetchCandles(symbol, interval string, startTime, endTime int64) ([]model.Candle, error){
	tags := map[string]string{
		"Historical" : "HitBTC",
	}
	return readHistoricalData(c.QueryAPI, c.InfluxDBBucket, c.TimeRange, c.Measurement, tags)
}

func (c *CandleServices)WriteCandleToDB(cdl model.Candle) error {
	tags := map[string]string{
		"Historical" : cdl.ExchName,
	}
	timestamp := cdl.Timestamp

	exists, err := dataPointExists(c.QueryAPI, c.InfluxDBBucket, c.TimeRange, c.Measurement, tags, timestamp)
	if err != nil {		
		fmt.Println("Error checking data point:", err)
		return fmt.Errorf("Error checking data point: %v", err)
	}

	if exists {
		fmt.Println("Data point already exists. Skipping write.")
		return fmt.Errorf("Data point already exists. Skipping write.")
	}

    point2 := influxdb2.NewPointWithMeasurement(c.Measurement).
		AddTag("Historical", cdl.ExchName).
        AddField("Close", cdl.Close).
        AddField("High", cdl.High).
        AddField("Low", cdl.Low).
        AddField("Open", cdl.Open).
        AddField("Volume", cdl.Volume).
        SetTime(time.Unix(cdl.Timestamp, 0))	

	// Write the point to the database
	err = c.WriteAPI.WritePoint(context.Background(), point2)
	if err != nil {
		return err
	}
	// fmt.Println("Candle inserted successfully!")
	return nil
}

func readHistoricalData(queryAPI api.QueryAPI, InfluxDBBucket, TimeRange, InfluxDBMeasurement string, tags map[string]string)([]model.Candle, error){
	query := fmt.Sprintf(`from(bucket: "%s")
	|> range(start: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s"`, InfluxDBBucket, TimeRange, InfluxDBMeasurement)

	for tagKey, tagValue := range tags {
		query += fmt.Sprintf(` and r["%s"] == "%s"`, tagKey, tagValue)
	}

	query += `)
	|> keep(columns: ["_time", "_value", "_field"])
	|> filter(fn: (r) => r["_field"] == "Close" and r["_value"] != 0.0)` // Filter for "Close" field values

	
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}
	defer result.Close()
	var mdcdls []model.Candle
	i := 0
	for result.Next() {
		// Parse row data
		record := result.Record()
		// fmt.Println(record.Time(), record.Field(), record.Value(), "i = ", i)	
		cdlTime := record.Time().Unix()
		v := fmt.Sprintf("%v",record.Value())
		cdlClose, _ := strconv.ParseFloat(v, 64)				
		mdcdls = append(mdcdls, model.Candle{Timestamp: cdlTime, Close: cdlClose})	
		i++	
	}
	return mdcdls, nil
}

func readAppData(queryAPI api.QueryAPI, InfluxDBBucket, TimeRange, InfluxDBMeasurement string, tags map[string]string)([]interface{}, error){
	// Construct the InfluxQL query
	query := fmt.Sprintf(`from(bucket: "%v")
	|> range(start: %s)
	|> filter(fn: (r) => r["_measurement"] == "%s"`, InfluxDBBucket, TimeRange, InfluxDBMeasurement)

	for tagKey, tagValue := range tags {
		query += fmt.Sprintf(` and r["%s"] == "%s"`, tagKey, tagValue)
	}

	query += ")"
	
	// result, err := queryDB(cli, `SELECT "bid_quantity" FROM "newmarket_data" WHERE "exchange"='Binance' AND "symbol"='BTCUSDT'`)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}
	defer result.Close()
	var (	
		md model.AppData
		mdcdls []interface{}
	)	
	for result.Next() {
		// Parse row data
		record := result.Record()
		measurement := record.Measurement()
		
		switch measurement{
		case "strategy_data":
			md.Strategy = "MACD"
			md.Count = record.ValueByKey("Count").(int)
			md.Strategy = record.ValueByKey("Strategy").(string)
			md.StrategyCombLogic = record.ValueByKey("StrategyCombLogic").(string)
			md.ShortPeriod = record.ValueByKey("ShortPeriod").(int)
			md.LongPeriod = record.ValueByKey("LongPeriod").(int)
			md.ShortMACDPeriod = record.ValueByKey("ShortMACDPeriod").(int)
			md.LongMACDPeriod = record.ValueByKey("LongMACDPeriod").(int)
			md.SignalMACDPeriod = record.ValueByKey("SignalMACDPeriod").(int)
			md.RSIPeriod = record.ValueByKey("RSIPeriod").(int)
			md.StochRSIPeriod = record.ValueByKey("StochRSIPeriod").(int)
			md.SmoothK = record.ValueByKey("SmoothK").(int)
			md.SmoothD = record.ValueByKey("SmoothD").(int)
			md.RSIOverbought = record.ValueByKey("RSIOverbought").(float64)
			md.RSIOversold = record.ValueByKey("RSIOversold").(float64)
			md.StRSIOverbought = record.ValueByKey("StRSIOverbought").(float64)
			md.StRSIOversold = record.ValueByKey("StRSIOversold").(float64)
			md.BollingerPeriod = record.ValueByKey("BollingerPeriod").(int)
			md.BollingerNumStdDev = record.ValueByKey("BollingerNumStdDev").(float64)
			md.TargetProfit = record.ValueByKey("TargetProfit").(float64)
			md.TargetStopLoss = record.ValueByKey("TargetStopLoss").(float64)
			md.RiskPositionPercentage = record.ValueByKey("RiskPositionPercentage").(float64)
			md.TotalProfitLoss = record.ValueByKey("TotalProfitLoss").(float64) 
			fmt.Printf("strategy_data: %+v\n", md)
			mdcdls = append(mdcdls, md)
		}
	}
	return mdcdls, nil
}

func (c *CandleServices)WriteAppDataToDB(md *model.AppData, timestamp int64) error {

    // Convert the structs to appropriate InfluxDB line protocol format
    point1 := influxdb2.NewPointWithMeasurement("strategy_data").
		AddTag("Strategy", md.Strategy).
		AddField("Count", md.Count).
		AddField("Strategy", md.Strategy).
		AddField("StrategyCombLogic", md.StrategyCombLogic).
		AddField("ShortPeriod", md.ShortPeriod).
		AddField("LongPeriod", md.LongPeriod).
		AddField("ShortMACDPeriod", md.ShortMACDPeriod).
		AddField("LongMACDPeriod", md.LongMACDPeriod).
		AddField("SignalMACDPeriod", md.SignalMACDPeriod).
		AddField("RSIPeriod", md.RSIPeriod).
		AddField("StochRSIPeriod", md.StochRSIPeriod).
		AddField("SmoothK", md.SmoothK).
		AddField("SmoothD", md.SmoothD).
		AddField("RSIOverbought", md.RSIOverbought).
		AddField("RSIOversold", md.RSIOversold).
		AddField("StRSIOverbought", md.StRSIOverbought).
		AddField("StRSIOversold", md.StRSIOversold).
		AddField("BollingerPeriod", md.BollingerPeriod).
		AddField("BollingerNumStdDev", md.BollingerNumStdDev).
		AddField("TargetProfit", md.TargetProfit).
		AddField("TargetStopLoss", md.TargetStopLoss).
		AddField("RiskPositionPercentage", md.RiskPositionPercentage).
		AddField("TotalProfitLoss", md.TotalProfitLoss). 
        SetTime(time.Unix(timestamp, 0))

	// Write the point to the database
	err := c.WriteAPI.WritePoint(context.Background(), point1)
	if err != nil {
		return err
	}
	fmt.Println("AppData inserted successfully!")

	return nil
}

func dataPointExists(queryAPI api.QueryAPI, InfluxDBBucket, TimeRange, measurement string, tags map[string]string, timestamp int64) (bool, error) {
	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: %s)
		|> filter(fn: (r) => r["_measurement"] == "%s"`,InfluxDBBucket, TimeRange, measurement)

	for tagKey, tagValue := range tags {
		query += fmt.Sprintf(` and r["%s"] == "%s"`, tagKey, tagValue)
	}

	query += fmt.Sprintf(` and r["_time"] == %d)`, timestamp)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return false, err
	}
	defer result.Close()

	return result.Next(), nil
}

func (c *CandleServices)DeleteBucket(client influxdb2.Client) {
	// Execute the DROP query to delete the bucket
	query := fmt.Sprintf("DROP mlai_data \"%s\"", c.InfluxDBBucket)
	_, err := c.QueryAPI.Query(context.Background(), query)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return
	}

	// Bucket deleted successfully
	fmt.Printf("Bucket %s deleted successfully!\n", c.InfluxDBBucket)
}

