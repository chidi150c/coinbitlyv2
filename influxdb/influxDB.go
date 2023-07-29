package influxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const (
	InfluxDBOrgID = "Resoledge"
	influxDBURL    = "http://localhost:8086" // Replace with your InfluxDB URL
	influxDBToken  = "aXDeT9-0EX6K81D_94L-6q5G-w2eHS_4FJTIbsanUNqHlziMrFTOD3JULdCkCWgCTtVPvIuBhxUB0asbt8_AYw=="   // Replace with your InfluxDB token
	influxDBBucket = "market_data"           // Use "market_data" as the bucket/measurement name
)

type MarketData struct {
	Exchange    string
	Symbol      string
	Price       float64
	Volume      float64
	BidPrice    float64
	BidQuantity float64
	AskPrice    float64
	AskQuantity float64
	Timestamp   time.Time
}

func NewMarketData(exname, symbol string) *MarketData {
	return &MarketData{
		Exchange: exname,
		Symbol:   symbol,
	}
}

// NewDB creates and initializes the InfluxDB client
func NewDB() (influxdb2.Client, error) {
	// Connect to InfluxDB
	client := influxdb2.NewClient(influxDBURL, influxDBToken)
	// defer client.Close()

	// Check if the client connection was successful
	ctx := context.Background()
	_, err := client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("error connecting to InfluxDB: %v", err)
	}
	return client, nil
}

// WriteDataPoint writes the market data point to InfluxDB
func WriteDataPoint(writeAPI api.WriteAPI, data *MarketData) error {
	// Create a new data point
	p := influxdb2.NewPointWithMeasurement(influxDBBucket).
		AddTag("exchange", data.Exchange).
		AddTag("symbol", data.Symbol).
		AddField("price", data.Price).
		AddField("volume", data.Volume).
		AddField("bid_price", data.BidPrice).
		AddField("bid_quantity", data.BidQuantity).
		AddField("ask_price", data.AskPrice).
		AddField("ask_quantity", data.AskQuantity).
		SetTime(data.Timestamp)

	// Write the data point to InfluxDB
	writeAPI.WritePoint(p)

	// Flush writes
    writeAPI.Flush()

	return nil
}
