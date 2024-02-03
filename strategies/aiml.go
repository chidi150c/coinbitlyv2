// strategies/fundamental_analysis.go

package strategies

import (
	// "encoding/csv"
	"fmt"
	// "log"
	// "os"
	// "strconv"
    // "github.com/sjwhitworth/golearn/base"
    // "github.com/sjwhitworth/golearn/ensemble"
    // "github.com/sjwhitworth/golearn/evaluation"

	"coinbitly.com/model" // Import the DataPoint struct
)

func (ts *TradingSystem)DataPointtoCSV(data *model.DataPoint)error { //fundamentalAnalysis()
    err := ts.CSVWriter.Write([]string{		
        data.Date.Format("02 15:04:05"),
		fmt.Sprintf("%f", data.DiffL95S15),
		fmt.Sprintf("%f", data.DiffL8S4),
		fmt.Sprintf("%f", data.RoCL8),
		fmt.Sprintf("%f", data.RoCS4),
		fmt.Sprintf("%f", data.MA5DiffL95S15),
		fmt.Sprintf("%f", data.MA5DiffL8S4),
		fmt.Sprintf("%f", data.StdDevL95),
		fmt.Sprintf("%f", data.StdDevS15),
		fmt.Sprintf("%f", data.LaggedL95EMA),
		fmt.Sprintf("%f", data.LaggedS15EMA),
		fmt.Sprintf("%f", data.ProfitLoss),
		fmt.Sprintf("%f", data.CurrentPrice),
		fmt.Sprintf("%f", data.StochRSI),
		fmt.Sprintf("%f", data.SmoothKRSI),
		fmt.Sprintf("%f", data.MACDLine),
		fmt.Sprintf("%f", data.MACDSigLine),
		fmt.Sprintf("%f", data.MACDHist),
		fmt.Sprintf("%d", data.OBV),
		fmt.Sprintf("%f", data.ATR),
		fmt.Sprintf("%d", data.Label),
		fmt.Sprintf("%d", data.Label2),
	})
	if err != nil {
        ts.Log.Printf("Writing Data to CSV ERROR %v", err)
        return err
	}
	// Flush the CSVWriter to ensure data is written to the file
    ts.CSVWriter.Flush()
	return nil
}
