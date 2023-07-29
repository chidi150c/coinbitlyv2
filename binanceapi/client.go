package binanceapi

import (
	"fmt"
)


// fetchAndDisplayTickerData fetches and displays real-time market data for the given symbol
func FetchAndDisplayTickerData(symbol string) (*TickerData, error) {
	ticker, err := fetchTickerData(symbol)
	if err != nil {
		fmt.Println("Error fetching ticker data:", err)
		return &TickerData{}, err
	}

	// Display the ticker data
	fmt.Printf("Symbol: %s\nPrice: %s\n", ticker.Symbol, ticker.Price)
	return ticker, nil
}

// fetchAndDisplay24hrTickerData fetches and displays 24-hour price change statistics for the given symbol
func FetchAndDisplay24hrTickerData(symbol string) (*Ticker24hrData, error){
	ticker, err := fetch24hrTickerData(symbol)
	if err != nil {
		fmt.Println("Error fetching 24-hour ticker data:", err)
		return &Ticker24hrData{}, err
	}

	// Display the 24-hour price change statistics
	fmt.Printf("Symbol: %s\nLast Price: %s\nPrice Change: %s\nVolume: %s\n",
		ticker.Symbol, ticker.LastPrice, ticker.PriceChange, ticker.Volume)
		return ticker, nil
}

// fetchAndDisplayOrderBook fetches and displays order book depth for the given symbol
func FetchAndDisplayOrderBook(symbol string, limit int) (*OrderBookData, error) {
	orderBook, err := fetchOrderBook(symbol, limit)
	if err != nil {
		fmt.Println("Error fetching order book data:", err)
		return &OrderBookData{}, err
	}

	// Display the order book data
	fmt.Printf("Symbol: %s\nTop %d Bids:\n", symbol, limit)
	for _, bid := range orderBook.Bids {
		fmt.Printf("Price: %s, Quantity: %s\n", bid[0], bid[1])
	}

	fmt.Printf("\nTop %d Asks:\n", limit)
	for _, ask := range orderBook.Asks {
		fmt.Printf("Price: %s, Quantity: %s\n", ask[0], ask[1])
	}
	return orderBook, nil
}