package influxdb

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"coinbitly.com/config"
	"coinbitly.com/model"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type CandleServices struct{
	HistoricalDBBucket string	
	LiveDBBucket string
	Client influxdb2.Client
	QueryAPI api.QueryAPI
	LiveWriteAPI api.WriteAPIBlocking
	LiveTimeRange string
	LiveMeasurement string
	LiveTag string	
	// HistoricalQueryAPI api.QueryAPI
	HistoricalWriteAPI api.WriteAPIBlocking
	HistoricalTimeRange string
	HistoricalMeasurement string
	HistoricalTag string
}

func NewAPIServices(configParam *config.ExchConfig) (*CandleServices, error){	
	// Create a new InfluxDB client 
	client := influxdb2.NewClientWithOptions(configParam.BaseURL, configParam.SecretKey, influxdb2.DefaultOptions())
	queryAPI := client.QueryAPI(configParam.DBOrgID)
	LivewriteAPI := client.WriteAPIBlocking(configParam.DBOrgID, configParam.LiveDBBucket)
	HistoricalwriteAPI := client.WriteAPIBlocking(configParam.DBOrgID, configParam.HistoricalDBBucket)
	return &CandleServices{
		HistoricalDBBucket: configParam.HistoricalDBBucket,
		LiveDBBucket: configParam.LiveDBBucket,
		Client: client,
		QueryAPI: queryAPI,
		HistoricalWriteAPI: HistoricalwriteAPI,
		HistoricalTimeRange: configParam.HistoricalTimeRange,
		HistoricalMeasurement: configParam.HistoricalMeasurement,
		HistoricalTag: configParam.HistoricalTag,		
		// LiveQueryAPI: queryAPI,
		LiveWriteAPI: LivewriteAPI,
		LiveTimeRange: configParam.LiveTimeRange,
		LiveMeasurement: configParam.LiveMeasurement,
		LiveTag: configParam.LiveTag,
	}, nil
}

func (c *CandleServices)FetchCandles(symbol, interval string, startTime, endTime int64) ([]model.Candle, error){
	tags := map[string]string{
		"Historical" : c.HistoricalTag,
	}
	return readHistoricalData(c.QueryAPI, c.HistoricalDBBucket, c.HistoricalTimeRange, c.HistoricalMeasurement, tags)
}
func (c *CandleServices)WriteCandleToDB(ClosePrice float64, Timestamp int64) error {
	return nil
	tags := map[string]string{
		"Historical" : c.HistoricalTag,
	}
	timestamp := Timestamp

	exists, err := dataPointExists(c.QueryAPI, c.HistoricalDBBucket, c.HistoricalTimeRange, c.HistoricalMeasurement, tags, timestamp)
	if err != nil {		
		fmt.Println("Error checking data point:", err)
		return fmt.Errorf("Error checking data point: %v", err)
	}

	if exists {
		fmt.Println("Data point already exists. Skipping write.")
		return fmt.Errorf("Data point already exists. Skipping write.")
	}

    point2 := influxdb2.NewPointWithMeasurement(c.HistoricalMeasurement).
		AddTag("Historical", c.HistoricalTag).
        AddField("Close", ClosePrice).
        SetTime(time.Unix(Timestamp, 0))	

	// Write the point to the database
	err = c.HistoricalWriteAPI.WritePoint(context.Background(), point2)
	if err != nil {
		return err
	}
	// fmt.Println("Candle inserted successfully!")
	return nil
}
func (c *CandleServices)FetchTicker(symbol string)(CurrentPrice float64, err error){
	CurrentPrice = rand.Float64()
	return CurrentPrice, err
}
func (c *CandleServices)WriteTickerToDB(ClosePrice float64, Timestamp int64) error {
	return nil
	tags := map[string]string{
		"Live" : c.LiveTag,
	}
	timestamp := Timestamp

	exists, err := dataPointExists(c.QueryAPI, c.LiveDBBucket, c.LiveTimeRange, c.LiveMeasurement, tags, timestamp)
	if err != nil {		
		fmt.Println("Error checking Livedata point:", err)
		return fmt.Errorf("Error checking Livedata point: %v", err)
	}

	if exists {
		fmt.Println("LiveData point already exists. Skipping write.")
		return fmt.Errorf("LiveData point already exists. Skipping write.")
	}

    point2 := influxdb2.NewPointWithMeasurement(c.LiveMeasurement).
		AddTag("Live", c.LiveTag).
        AddField("Close", ClosePrice).
        SetTime(time.Unix(Timestamp, 0))	

	// Write the point to the database
	err = c.LiveWriteAPI.WritePoint(context.Background(), point2)
	if err != nil {
		return err
	}
	// fmt.Println("Candle inserted successfully!")
	return nil
}
func (c *CandleServices)CloseDB()error{
	c.Client.Close()
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
	fmt.Println("Interval:", TimeRange, "CandleCount:", len(mdcdls))
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
	return nil
    // Convert the structs to appropriate InfluxDB line protocol format
    point1 := influxdb2.NewPointWithMeasurement("strategy_data").
		AddTag("Strategy", md.Strategy).
		AddField("Count", md.Count).
		AddField("Strategy", md.Strategy).
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
	err := c.HistoricalWriteAPI.WritePoint(context.Background(), point1)
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

// func (c *CandleServices)DeleteBucket(client influxdb2.Client) {
// 	// Execute the DROP query to delete the bucket
// 	query := fmt.Sprintf("DROP mlai_data \"%s\"", c.HistoricalDBBucket)
// 	_, err := c.QueryAPI.Query(context.Background(), query)
// 	if err != nil {
// 		fmt.Println("Error executing query:", err)
// 		return
// 	}

// 	// Bucket deleted successfully
// 	fmt.Printf("Bucket %s deleted successfully!\n", c.HistoricalDBBucket)
// }

