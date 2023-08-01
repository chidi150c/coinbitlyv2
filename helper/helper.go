package helper

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// rateLimitedTransport is a custom transport that handles API rate limiting
type RateLimitedTransport struct {
	Base http.RoundTripper
}

func (t RateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Implement rate limiting logic here (e.g., using a rate limiter)
	// For simplicity, we'll use a simple time.Sleep to limit the rate
	time.Sleep(100 * time.Millisecond)
	return t.Base.RoundTrip(req)
}

// ParseStringToFloat parses a string to a float64 value
func ParseStringToFloat(str string) float64 {
	val, _ := strconv.ParseFloat(str, 64)
	return val
}

func IntervalToDuration(interval string) time.Duration {
	switch interval {
	case "M1":
		return time.Duration(time.Minute.Seconds())
	case "M5":
		return 5 * time.Duration(time.Minute.Seconds())
	case "M15":
		return 15 * time.Duration(time.Minute.Seconds())
	case "H1":
		return time.Duration(time.Hour.Seconds())
	// Add more intervals as needed
	default:
		return 2 * time.Duration(time.Hour.Seconds())
	}
}

func CalculateCandleCount(startTime, endTime int64, interval int64) int64 {
	fmt.Printf("calculateCandleCount: startTime %v , endTime %v , interval %v\n",startTime, endTime, interval )
    return (endTime - startTime) / interval
}