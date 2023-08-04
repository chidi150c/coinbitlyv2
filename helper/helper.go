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
	case "m1":
		return time.Duration(time.Minute.Seconds())
	case "m3":
		return 3 * time.Duration(time.Minute.Seconds())
	case "m5":
		return 5 * time.Duration(time.Minute.Seconds())
	case "m15":
		return 15 * time.Duration(time.Minute.Seconds())
	case "m30":
		return 30 * time.Duration(time.Minute.Seconds())
	case "H1":
		return time.Duration(time.Hour.Seconds())
	case "H4":
		return 4 * time.Duration(time.Hour.Seconds())
	case "D1":
		return 24 * time.Duration(time.Hour.Seconds())
	case "W1":
		return 7 * 24 * time.Duration(time.Hour.Seconds())
	case "M1":
		return 4 * 7 * 24 * time.Duration(time.Hour.Seconds())
	case "1m":
		return time.Duration(time.Minute.Seconds())
	case "3m":
		return 3 * time.Duration(time.Minute.Seconds())
	case "5m":
		return 5 * time.Duration(time.Minute.Seconds())
	case "15m":
		return 15 * time.Duration(time.Minute.Seconds())
	case "30m":
		return 30 * time.Duration(time.Minute.Seconds())
	case "1H":
		return time.Duration(time.Hour.Seconds())
	case "4H":
		return 4 * time.Duration(time.Hour.Seconds())
	case "1D":
		return 24 * time.Duration(time.Hour.Seconds())
	case "1W":
		return 7 * 24 * time.Duration(time.Hour.Seconds())
	case "1M":
		return 4 * 7 * 24 * time.Duration(time.Hour.Seconds())
	default:
		return time.Duration(time.Hour.Seconds())
	}
}

func CalculateCandleCount(startTime, endTime int64, interval int64) int64 {
	fmt.Printf("calculateCandleCount: startTime %v , endTime %v , interval %v\n",startTime, endTime, interval )
    return (endTime - startTime) / interval
}
