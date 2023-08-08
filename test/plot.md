<!-- package main

import (
	"fmt"

	"coinbitly.com/strategies"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// ... (Previous code for RSI, ExponentialMovingAverage, StochasticRSI)

// CustomCircleGlyph is a custom glyph that represents a circle.
type CustomCircleGlyph struct{}

// DrawGlyph draws a circle glyph at the given point using the provided style.
func (g CustomCircleGlyph) DrawGlyph(c draw.Canvas, sty draw.GlyphStyle, pt vg.Point) {
	c.DrawCircle(pt, vg.Points(3), sty)
}

// CustomPyramidGlyph is a custom glyph that represents a pyramid.
type CustomPyramidGlyph struct{}

// DrawGlyph draws a pyramid glyph at the given point using the provided style.
func (g CustomPyramidGlyph) DrawGlyph(c draw.Canvas, sty draw.GlyphStyle, pt vg.Point) {
	side := vg.Points(6)
	halfSide := side / 2
	pts := []vg.Point{
		{X: pt.X - halfSide, Y: pt.Y - halfSide}, // Bottom left
		{X: pt.X + halfSide, Y: pt.Y - halfSide}, // Bottom right
		{X: pt.X, Y: pt.Y + halfSide},            // Top
	}
	c.FillPolygon(sty, pts)
}

func createPoints(data []float64) plotter.XYs {
	pts := make(plotter.XYs, len(data))
	for i, value := range data {
		pts[i].X = float64(i)
		pts[i].Y = value
	}
	return pts
}

func plotRSI(data []float64, rsiPeriod, stochRSIPeriod, smoothK, smoothD int) {
	// Calculate RSI
	rsi := strategies.CalculateRSI(data, rsiPeriod)

	// Calculate Stochastic RSI and SmoothK, SmoothD values
	stochRSI, smoothKRSI := strategies.StochasticRSI(rsi, stochRSIPeriod, smoothK, smoothD)

	// Create a new plot
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// Create a scatter plot for Stochastic RSI points
	stochRSIScatter, err := plotter.NewScatter(createPoints(stochRSI))
	if err != nil {
		panic(err)
	}
	stochRSIScatter.GlyphStyle.Shape = CustomCircleGlyph{}

	// Create a scatter plot for SmoothK RSI points
	smoothKRSIScatter, err := plotter.NewScatter(createPoints(smoothKRSI))
	if err != nil {
		panic(err)
	}
	smoothKRSIScatter.GlyphStyle.Shape = CustomPyramidGlyph{}

	// Add the scatter plots to the plot
	p.Add(stochRSIScatter, smoothKRSIScatter)

	// Save the plot to a PNG file
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "stochastic_rsi_plot.png"); err != nil {
		panic(err)
	}
}

func main() {
	// Example usage
	data := []float64{50, 60, 70, 80, 90, 100, 110, 120, 130, 120, 110, 100}

	// Define the period for RSI and Stochastic RSI calculation
	rsiPeriod := 14
	stochRSIPeriod := 3
	smoothK := 3
	smoothD := 3

	// Generate and plot the Stochastic RSI and SmoothK RSI values
	plotRSI(data, rsiPeriod, stochRSIPeriod, smoothK, smoothD)

	fmt.Println("Plot saved to stochastic_rsi_plot.png")
} -->
