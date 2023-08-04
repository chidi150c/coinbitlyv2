package strategies

import (
	"image/color"
	"time"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func CreateLineChartWithSignals(timeSeries []int64, dataSeries []float64, signals []string) error {
	// Create a new plot with a title and axis labels.
	p := plot.New()
	p.Title.Text = "Line Chart with Trading Signals"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Data Value"

	// Create a new set of points based on the data.
	pts := make(plotter.XYs, len(timeSeries))
	for i := range pts {
		pts[i].X = float64(i)
		pts[i].Y = dataSeries[i]
	}

	// Create a new line plot and set its style.
	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	line.Color = color.RGBA{B: 255, A: 255} // Blue color

	// Add the line plot to the plot.
	p.Add(line)

	// Convert UNIX timestamps to formatted strings for X-axis labels.
	timeLabels := make([]string, len(timeSeries))
	for i, timestamp := range timeSeries {
		timeLabels[i] = time.Unix(timestamp, 0).Format("2006-01-02")
	}

	// Add custom tick marks for the X-axis (time labels).
	p.NominalX(timeLabels...)

	// Create separate scatter plots for Buy and Sell signals.
	var buySignalPoints, sellSignalPoints plotter.XYs
	for i, signal := range signals {
		if signal == "Buy" {
			buySignalPoints = append(buySignalPoints, plotter.XY{X: float64(i), Y: dataSeries[i]})
		} else if signal == "Sell" {
			sellSignalPoints = append(sellSignalPoints, plotter.XY{X: float64(i), Y: dataSeries[i]})
		}
	}

	// Create scatter plots for Buy and Sell signals.
	buyScatter, err := plotter.NewScatter(buySignalPoints)
	if err != nil {
		return err
	}
	buyScatter.Shape = plotutil.Shape(3) // Triangle shape for Buy signals
	buyScatter.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Green color for Buy signals
	buyScatter.GlyphStyle.Radius = vg.Points(5)

	sellScatter, err := plotter.NewScatter(sellSignalPoints)
	if err != nil {
		return err
	}
	sellScatter.Shape = plotutil.Shape(1) // Circle shape for Sell signals
	sellScatter.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red color for Sell signals
	sellScatter.GlyphStyle.Radius = vg.Points(5)

	// Add the scatter plots to the plot.
	p.Add(buyScatter, sellScatter)

	// Save the plot to a file (you can also display it in a window if you prefer).
	if err := p.Save(10*vg.Inch, 6*vg.Inch, "line_chart_with_signals.png"); err != nil {
		return err
	}

	return nil
}

