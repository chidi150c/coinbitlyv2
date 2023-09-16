package strategies

import (
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

	// Create the TradingSystem in the database
	tradeID, err := rdbServices.CreateDBTradingSystem(ts)
	assert.NoError(t, err) // Ensure there are no errors

	// Read the TradingSystem from the database using the ID obtained from creation
	readTS, err := rdbServices.ReadDBTradingSystem(tradeID)
	assert.NoError(t, err) // Ensure there are no errors

	// Assert that the read TradingSystem matches the original one
	assert.Equal(t, ts.Symbol, readTS.Symbol)
	assert.ElementsMatch(t, ts.ClosingPrices, readTS.ClosingPrices)
	// Add more assertions for other fields as needed

	// Clean up: Delete the created TradingSystem to keep the test environment clean
	err = rdbServices.DeleteDBTradingSystem(tradeID)
	assert.NoError(t, err) // Ensure there are no errors while deleting
}

