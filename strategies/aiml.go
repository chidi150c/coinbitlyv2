// strategies/fundamental_analysis.go

package strategies

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

    // "github.com/sjwhitworth/golearn/linear_models"

	"coinbitly.com/model" // Import the AppData struct
)

func AppDatatoCSV(data []*model.AppData) { //fundamentalAnalysis()
    // Save data to CSV file
    file, err := os.Create("data.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    for _, d := range data {
        err := writer.Write([]string{
            fmt.Sprintf("%d", d.Count),
			d.Strategy,
			d.StrategyCombLogic,
			fmt.Sprintf("%d", d.ShortPeriod),
			fmt.Sprintf("%d", d.LongPeriod),
			fmt.Sprintf("%d", d.ShortMACDPeriod),
			fmt.Sprintf("%d", d.LongMACDPeriod),
			fmt.Sprintf("%d", d.SignalMACDPeriod),
			fmt.Sprintf("%d", d.RSIPeriod),
			fmt.Sprintf("%d", d.StochRSIPeriod),
			fmt.Sprintf("%d", d.SmoothK),
			fmt.Sprintf("%d", d.SmoothD),
			fmt.Sprintf("%f", d.RSIOverbought),
			fmt.Sprintf("%f", d.RSIOversold),
			fmt.Sprintf("%f", d.StRSIOverbought),
			fmt.Sprintf("%f", d.StRSIOversold),
			fmt.Sprintf("%d", d.BollingerPeriod),
			fmt.Sprintf("%f", d.BollingerNumStdDev),
			fmt.Sprintf("%f", d.TargetProfit),
			fmt.Sprintf("%f", d.TargetStopLoss),
			fmt.Sprintf("%f", d.RiskPositionPercentage),
			d.Scalping,
			fmt.Sprintf("%f", d.TotalProfitLoss),
        })
        if err != nil {
            log.Fatal(err)
        }
    }
}

func CSVtoAppData(filename string) ([]*model.AppData, error) {
    var data []*model.AppData

    file, err := os.Open(filename)
    if err != nil {
        return data, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return data, err
    }

    for _, record := range records {
        // Convert record fields to appropriate types and populate AppData instances
        Count, _ := strconv.Atoi(record[0])
		ShortPeriod, _ := strconv.Atoi(record[3])
		LongPeriod, _ := strconv.Atoi(record[4])
		ShortMACDPeriod, _ := strconv.Atoi(record[5])
		LongMACDPeriod, _ := strconv.Atoi(record[6])
		SignalMACDPeriod, _ := strconv.Atoi(record[7])
		RSIPeriod, _ := strconv.Atoi(record[8])
		StochRSIPeriod, _ := strconv.Atoi(record[9])
		SmoothK, _ := strconv.Atoi(record[10])
		SmoothD, _ := strconv.Atoi(record[11])
		RSIOverbought, _ := strconv.ParseFloat(record[12], 64)
		RSIOversold, _ := strconv.ParseFloat(record[13], 64)
		StRSIOverbought, _ := strconv.ParseFloat(record[14], 64)
		StRSIOversold, _ := strconv.ParseFloat(record[15], 64)
		BollingerPeriod, _ := strconv.Atoi(record[16])
		BollingerNumStdDev, _ := strconv.ParseFloat(record[17], 64)
		TargetProfit, _ := strconv.ParseFloat(record[18], 64)
		TargetStopLoss, _ := strconv.ParseFloat(record[19], 64)
		RiskPositionPercentage, _ := strconv.ParseFloat(record[20], 64)
		TotalProfitLoss, _ := strconv.ParseFloat(record[22], 64)

        data = append(data, &model.AppData{
            Count:             Count,
            Strategy:          record[1],
            StrategyCombLogic: record[2],
            ShortPeriod:       ShortPeriod,
            LongPeriod:        LongPeriod,		
			ShortMACDPeriod:	ShortMACDPeriod,
			LongMACDPeriod :	LongMACDPeriod,
			SignalMACDPeriod:	SignalMACDPeriod,
			RSIPeriod:RSIPeriod,
			StochRSIPeriod:	StochRSIPeriod,
			SmoothK:	SmoothK,
			SmoothD:	SmoothD,
			RSIOverbought:	RSIOverbought,
			RSIOversold:	RSIOversold,
			StRSIOverbought: StRSIOverbought,
			StRSIOversold: StRSIOversold,
			BollingerPeriod: BollingerPeriod,
			BollingerNumStdDev: BollingerNumStdDev,
			TargetProfit: TargetProfit,
			TargetStopLoss: TargetStopLoss,
			RiskPositionPercentage: RiskPositionPercentage,
			Scalping: record[21],
			TotalProfitLoss:	TotalProfitLoss,
        })
    }

    return data, nil
}

// func MLPrediction()string{

//     // Read data from CSV (uses your AppData struct)
//     data, err := CSVtoAppData("data.csv")
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Convert data to suitable format for SVM (feature matrix X, target vector y)
//     var X [][]float64
//     var y []float64
//     for _, d := range data {
//         // Extract relevant features and target (price movement)
//         featureVector := []float64{
//             float64(d.ShortPeriod),
//             float64(d.LongPeriod),
//             // Include other features as needed...
//         }
//         target := computePriceMovement(d) // Implement this function to compute price movement
//         X = append(X, featureVector)
//         y = append(y, target)
//     }

//     // Train a Support Vector Machine (SVM) model
//     model := linear_models.NewLinearSVC()
//     err = model.Fit(X, y)
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Simulated input features for prediction (replace with real-time data)
//     inputFeatures := []float64{10.0, 30.0} // Example values for ShortPeriod and LongPeriod

//     // Make predictions using the trained SVM model
//     predictedPriceMovement, err := model.Predict(inputFeatures)
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Interpret the predicted price movement
//     predictedDirection := "unknown"
//     if predictedPriceMovement > 0 {
//         predictedDirection = "up"
//     } else if predictedPriceMovement < 0 {
//         predictedDirection = "down"
//     }

//     fmt.Printf("Predicted price movement: %s\n", predictedDirection)
// 	return predictedDirection
// }

func computePriceMovement(data *model.AppData) float64 {
    // Implement a function to compute price movement
    // This function should use your AppData struct and perform the necessary calculations
    // Return a value indicating the predicted price movement (e.g., positive for an expected price increase)

    // For the purpose of this example, let's return a simulated price movement
    // Replace this with a meaningful calculation based on your trading strategy
    return float64(data.Count) * 0.01 // Assuming a simple linear relation
}


