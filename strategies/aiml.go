// strategies/fundamental_analysis.go

package strategies

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	// "strconv"
    // "github.com/sjwhitworth/golearn/base"
    // "github.com/sjwhitworth/golearn/ensemble"
    // "github.com/sjwhitworth/golearn/evaluation"

	"coinbitly.com/model" // Import the DataPoint struct
)

func (ts *TradingSystem)DataPointtoCSV(data *model.DataPoint)error { //fundamentalAnalysis()
    ts.Log.Printf("Writing Data to CSV %v", data)
    err := ts.CSVWriter.Write([]string{
		data.Date.Format("02 15:04:05"),
        fmt.Sprintf("%f", data.L95EMA),
        fmt.Sprintf("%f", data.S15EMA),
        fmt.Sprintf("%f", data.L8EMA ),
        fmt.Sprintf("%f", data.S4EMA ),
        fmt.Sprintf("%f", data.DiffL95S15),
        fmt.Sprintf("%f", data.DiffL8S4),
        fmt.Sprintf("%f", data.RoCL95),
        fmt.Sprintf("%f", data.RoCS15),
        fmt.Sprintf("%f", data.MA5DiffL95S15),
        fmt.Sprintf("%f", data.MA5DiffL8S4 ),
        fmt.Sprintf("%f", data.StdDevL95),
        fmt.Sprintf("%f", data.StdDevS15),
        fmt.Sprintf("%d", data.CrossL95S15),
        fmt.Sprintf("%f", data.LaggedL95EMA),
        fmt.Sprintf("%f", data.LaggedS15EMA),
        fmt.Sprintf("%d", data.Label),
	})
	if err != nil {
        ts.Log.Printf("Writing Data to CSV ERROR %v", err)
        return err
	}
	return nil
}

func DataPointListtoCSV(data []*model.DataPoint) { //fundamentalAnalysis()
    // Save data to CSV file
    file, err := os.Create("./webclient/assets/data.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write headers to the CSV file

    headers := []string{
        "Date","L95EMA","S15EMA","L8EMA","S4EMA","DiffL95S15","DiffL8S4","RoCL95","RoCS15","MA5DiffL95S15",
        "MA5DiffL8S4","StdDevL95","StdDevS15","CrossL95S15","LaggedL95EMA","LaggedS15EMA","Label", 
    }
    err = writer.Write(headers)
    if err != nil {
        log.Fatal(err)
    }

    for _, d := range data {
        err := writer.Write([]string{
            d.Date.Format("02 15:04:05"),
            fmt.Sprintf("%f", d.L95EMA),
            fmt.Sprintf("%f", d.S15EMA),
            fmt.Sprintf("%f", d.L8EMA ),
            fmt.Sprintf("%f", d.S4EMA ),
            fmt.Sprintf("%f", d.DiffL95S15),
            fmt.Sprintf("%f", d.DiffL8S4),
            fmt.Sprintf("%f", d.RoCL95),
            fmt.Sprintf("%f", d.RoCS15),
            fmt.Sprintf("%f", d.MA5DiffL95S15),
            fmt.Sprintf("%f", d.MA5DiffL8S4 ),
            fmt.Sprintf("%f", d.StdDevL95),
            fmt.Sprintf("%f", d.StdDevS15),
            fmt.Sprintf("%d", d.CrossL95S15),
            fmt.Sprintf("%f", d.LaggedL95EMA),
            fmt.Sprintf("%f", d.LaggedS15EMA),
            fmt.Sprintf("%d", d.Label),
        })
        if err != nil {
            log.Fatal(err)
        }
    }
}

// func CSVtoDataPoint(filename string) ([]*model.DataPoint, error) {
//     var data []*model.DataPoint

//     file, err := os.Open(filename)
//     if err != nil {
//         return data, err
//     }
//     defer file.Close()

//     reader := csv.NewReader(file)
//     records, err := reader.ReadAll()
//     if err != nil {
//         return data, err
//     }

//     for _, record := range records {
//         // Convert record fields to appropriate types and populate DataPoint instances
//         DataPoint, _ := strconv.Atoi(record[0])
// 		ShortPeriod, _ := strconv.Atoi(record[3])
// 		LongPeriod, _ := strconv.Atoi(record[4])
// 		TargetProfit, _ := strconv.ParseFloat(record[18], 64)
// 		TargetStopLoss, _ := strconv.ParseFloat(record[19], 64)
// 		RiskPositionPercentage, _ := strconv.ParseFloat(record[20], 64)
// 		TotalProfitLoss, _ := strconv.ParseFloat(record[22], 64)

//         data = append(data, &model.DataPoint{
//             DataPoint:             DataPoint,
//             Strategy:          record[1],
//             ShortPeriod:       ShortPeriod,
//             LongPeriod:        LongPeriod,		
// 			TargetProfit: TargetProfit,
// 			TargetStopLoss: TargetStopLoss,
// 			RiskPositionPercentage: RiskPositionPercentage,
// 			TotalProfitLoss:	TotalProfitLoss,
//         })
//     }

//     return data, nil
// }