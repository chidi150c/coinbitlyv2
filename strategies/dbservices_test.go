package strategies

import (
	"fmt"
	"testing"

	// "coinbitly.com/model"
	"github.com/stretchr/testify/assert"
)
func TestCreateAndReadDBTradingSystem(t *testing.T) {
	// Initialize the RDBServices instance
	rdbServices := NewRDBServices("Testnet")

	// Create a sample TradingSystem
	ts := &TradingSystem{
		Symbol:        "BTC/USDT",
		ClosingPrices: []float64{100.0, 105.0, 110.0},
		// Initialize other fields as needed
	}
	var err error 
	// Create the TradingSystem in the database
	ts.ID, err = rdbServices.CreateDBTradingSystem(ts)
	assert.NoError(t, err) // Ensure there are no errors
	fmt.Printf("created ts.ID %d\n", ts.ID)
	// Read the TradingSystem from the database using the ID obtained from creation
	readTS, err := rdbServices.ReadDBTradingSystem(ts.ID)
	assert.NoError(t, err) // Ensure there are no errors
	fmt.Printf("Read ts.ID %d but created ts.ID %d\n", readTS.ID, ts.ID)
	
	// Assert that the read TradingSystem matches the original one
	assert.Equal(t, ts.Symbol, readTS.Symbol)
	assert.ElementsMatch(t, ts.ClosingPrices, readTS.ClosingPrices)
	// Add more assertions for other fields as needed

	// // Clean up: Delete the created TradingSystem to keep the test environment clean
	err = rdbServices.DeleteDBTradingSystem(ts.ID)
	assert.NoError(t, err) // Ensure there are no errors while deleting
	
	// Read the TradingSystem from the database using the ID obtained from creation
	// readTS, err = rdbServices.ReadDBTradingSystem(0)
	// assert.NoError(t, err) // Ensure there are no errors
	// fmt.Printf("Read ts.ID %d but created ts.ID %d\n", readTS.ID, ts.ID)
	// panic("")
}

