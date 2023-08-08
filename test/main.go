package main

import (
	"fmt"

	"coinbitly.com/strategies"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	// "gonum.org/v1/plot/vg/draw"
)

func main() {
	// Example usage
	data := []float64{50, 60, 70, 80, 90, 100, 110, 120, 130, 120, 110, 100, 90, 60, 50, 60, 70, 80, 90, 100, 110, 120, 130, 120, 110, 100, 90, 60}

	// Define the period for RSI and Stochastic RSI calculation
	rsiPeriod := 14
	stochRSIPeriod := 3
	smoothK := 3
	smoothD := 3

	// Generate and plot the Stochastic RSI and SmoothK RSI values
	plotRSI(data, rsiPeriod, stochRSIPeriod, smoothK, smoothD)

	fmt.Println("Plot saved to stochastic_rsi_plot.png")
}

func createPoints(data []float64) plotter.XYs {
	pts := make(plotter.XYs, len(data))
	for i, value := range data {
		pts[i].X = float64(i)
		pts[i].Y = value
	}
	return pts
}

