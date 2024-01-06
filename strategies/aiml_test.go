package strategies

import (
	"encoding/csv"
	"log"
	"os"
	"testing"
	"time"

	"coinbitly.com/model"
	"github.com/stretchr/testify/assert"
)

// TestDataPointtoCSV tests the DataPointtoCSV function
func TestDataPointtoCSV(t *testing.T) {
	// Open or create a log file for appending
	logFile, err := os.OpenFile("test_output.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening or creating log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Initialize your TradingSystem instance with a real CSVWriter
	ts := &TradingSystem{
		Log:       log.New(logFile, "", log.LstdFlags),
		CSVWriter: csv.NewWriter(logFile),
	}

	// Define a sample DataPoint
	data := &model.DataPoint{
		Date:  time.Now(),
		// Fill in other fields with appropriate values for testing
	}

	// Call the function being tested
	err = ts.DataPointtoCSV(data)

	// Check for expected behavior
	assert.NoError(t, err) // Assert that no error occurred during writing

	// Flush the CSVWriter to ensure data is written to the file
	ts.CSVWriter.Flush()
	assert.NoError(t, ts.CSVWriter.Error())
}

