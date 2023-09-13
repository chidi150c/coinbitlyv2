package strategies

import (
	"fmt"
	"image/color"
	"time"

	"coinbitly.com/model"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func (ts *TradingSystem)CreateLineChartWithSignals(md *model.AppData, timeSeries []int64, dataSeries []float64, signals []string, graph string) error {
	// ts.ChartChan <- model.ChartData{
	// 	ClosingPrices: dataSeries[len(dataSeries)-1], 
	// 	Timestamps: timeSeries[len(dataSeries)-1], 
	// 	Signals: signals[len(dataSeries)-1], 
	// 	ShortEMA: 0.0,
	// 	LongEMA: 0.0,
	// }
	// Create a new plot with a title and axis labels.
	p := plot.New()
	p.Title.Text = ts.BaseCurrency+ " Price and Buy/Sell Signal Trading Chart"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = ts.BaseCurrency+" Price in USDT"

	// Create a new set of points based on the data.
	pts := make(plotter.XYs, len(dataSeries))
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
	if len(timeLabels) > 0 {
		p.NominalX(timeLabels...)
	}

	// Create separate scatter plots for Buy and Sell signals.
	var buySignalPoints, sellSignalPoints plotter.XYs
	for i, signal := range signals {
		if i < len(dataSeries){
			if signal == "Buy" {
				buySignalPoints = append(buySignalPoints, plotter.XY{X: float64(i), Y: dataSeries[i]})
			} else if signal == "Sell" {
				sellSignalPoints = append(sellSignalPoints, plotter.XY{X: float64(i), Y: dataSeries[i]})
			}			
		}
	}

	// Create scatter plots for Buy and Sell signals.
	buyScatter, err := plotter.NewScatter(buySignalPoints)
	if err != nil {
		return err
	}
	buyScatter.Shape = plotutil.Shape(3) // Triangle shape for Buy signals
	buyScatter.Color = color.RGBA{R: 213, G: 184, B: 255, A: 255} // Green color for Buy signals
	buyScatter.GlyphStyle.Radius = vg.Points(7)

	sellScatter, err := plotter.NewScatter(sellSignalPoints)
	if err != nil {
		return err
	}
	sellScatter.Shape = plotutil.Shape(1) // Circle shape for Sell signals
	sellScatter.Color = color.RGBA{R: 213, G: 184, B: 255, A: 255} // (191, 85, 236);, 1Red color for Sell signals
	sellScatter.GlyphStyle.Radius = vg.Points(7)

	// Add the scatter plots to the plot.
	p.Add(buyScatter, sellScatter)

	
	// Create a legend and add entries for your plots.
	p.Legend.Add(ts.BaseCurrency+" Price", line)
	p.Legend.Add("Buy Signals", buyScatter)
	p.Legend.Add("Sell Signals", sellScatter)
	p.Legend.Top = true
	p.Legend.Left = false
	p.Legend.Padding = 5.0
	p.Legend.TextStyle.Color = color.White
    p.BackgroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	p.X.Color = color.White
    p.Y.Color = color.White
	
    // Set labels to white color
	p.X.Label.TextStyle.Color = color.White
	p.Y.Label.TextStyle.Color = color.White

    // Set scaling labels to white color
    p.X.Tick.Label.Color = color.White
    p.Y.Tick.Label.Color = color.White

	// Save the plot to a file (you can also display it in a window if you prefer).
	if err := p.Save(10*vg.Inch, 6*vg.Inch, "./webclient/assets/"+graph+"line_chart_with_signals.png"); err != nil {
		return err
	}
	return nil
}

func (ts *TradingSystem)CreateLineChartWithSignalsV3(md *model.AppData, timeSeries []int64, dataSeries []float64, greenDataSeries []float64, yellowDataSeries []float64, signals []string, graph string) error {
	// ts.ChartChan <- model.ChartData{
	// 	ClosingPrices: dataSeries[len(dataSeries)-1], 
	// 	Timestamps: timeSeries[len(dataSeries)-1], 
	// 	Signals: signals[len(dataSeries)-1], 
	// 	ShortEMA: greenDataSeries[len(dataSeries)-1],
	// 	LongEMA: yellowDataSeries[len(dataSeries)-1],
	// }
	// Create a new plot with a title and axis labels.
	p := plot.New()
	p.Title.Text = ts.BaseCurrency+ " Price, EMA and Buy/Sell Signal Trading Chart"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = ts.BaseCurrency+" Price in USDT"

	// Create a new set of points based on the data.
	pts := make(plotter.XYs, len(dataSeries))
	for i := range pts {
		pts[i].X = float64(i)
		pts[i].Y = dataSeries[i]
	}

	// Create a new set of points based on the green data series.
	greenPts := make(plotter.XYs, len(greenDataSeries))
	for i := range greenPts {
		greenPts[i].X = float64(i)
		greenPts[i].Y = greenDataSeries[i]
	}

	// Create a new set of points based on the yellow data series.
	yellowPts := make(plotter.XYs, len(yellowDataSeries))
	for i := range yellowPts {
		yellowPts[i].X = float64(i)
		yellowPts[i].Y = yellowDataSeries[i]
	}

	// Create a new line plot and set its style.
	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	line.Color = color.RGBA{B: 255, A: 255} // Blue color

	// Create a new line plot for the green data series and set its style.
	greenLine, err := plotter.NewLine(greenPts)
	if err != nil {
		return err
	}
	greenLine.Color = color.RGBA{G: 255, A: 255} // Green color

	// Create a new line plot for the yellow data series and set its style.
	yellowLine, err := plotter.NewLine(yellowPts)
	if err != nil {
		return err
	}
	yellowLine.Color = color.RGBA{R: 255, G: 242, A: 255} // yellow color


	// Add the line plot to the plot.
	p.Add(line)

	// Add the green line plot to the plot.
	p.Add(greenLine)

	// Add the yellow line plot to the plot.
	p.Add(yellowLine)

	// Calculate the start and end times for the 7-day range.
	endTime := timeSeries[len(timeSeries)-1]
	startTime := timeSeries[0]

	// Calculate the number of tick marks you want.
	numTicks := 10
	timeRange := endTime - startTime
	stepSize := timeRange / int64(numTicks)
	nextMark := startTime + stepSize

	// Create a slice to hold the selected time labels.
	selectedTimeLabels := make([]string, len(timeSeries))
	for i, timeVal := range timeSeries {
		if timeVal >= nextMark {
			selectedTimeLabels[i] = time.Unix(0, timeVal*int64(time.Millisecond)).Format("02 15:04:05")
			nextMark += stepSize
		} else {
			selectedTimeLabels[i] = ""
		}
	}

	// Add custom tick marks for the X-axis using the selected time labels.
	p.NominalX(selectedTimeLabels...)

	// Create separate scatter plots for Buy and Sell signals.
	var buySignalPoints, sellSignalPoints plotter.XYs
	for i, signal := range signals {
		if i < len(dataSeries){
			if signal == "Buy" {
				buySignalPoints = append(buySignalPoints, plotter.XY{X: float64(i), Y: dataSeries[i]})
			} else if signal == "Sell" {
				sellSignalPoints = append(sellSignalPoints, plotter.XY{X: float64(i), Y: dataSeries[i]})
			}
		}
	}

	// Create scatter plots for Buy and Sell signals.
	buyScatter, err := plotter.NewScatter(buySignalPoints)
	if err != nil {
		return err
	}
	buyScatter.Shape = plotutil.Shape(3) // Triangle shape for Buy signals
	buyScatter.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red color for Buy signals
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

	// Create a legend and add entries for your plots.
	p.Legend.Add(ts.BaseCurrency+" Price", line)
	p.Legend.Add(fmt.Sprintf("%d", md.ShortPeriod)+"PeriodEMA", greenLine)
	p.Legend.Add(fmt.Sprintf("%d", md.LongPeriod)+"PeriodEMA", yellowLine)
	p.Legend.Add("Buy Signals", buyScatter)
	p.Legend.Add("Sell Signals", sellScatter)
	p.Legend.Top = true
	p.Legend.Left = false
	p.Legend.Padding = 5.0
	
	
	p.Legend.TextStyle.Color = color.White
    p.BackgroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	p.X.Color = color.White
    p.Y.Color = color.White

    // Set labels to white color
	p.X.Label.TextStyle.Color = color.White
	p.Y.Label.TextStyle.Color = color.White

    // Set scaling labels to white color
    p.X.Tick.Label.Color = color.White
    p.Y.Tick.Label.Color = color.White

	// Save the plot to a file (you can also display it in a window if you prefer).
	if err := p.Save(10*vg.Inch, 6*vg.Inch, "./webclient/assets/"+graph+"line_chart_with_signals.png"); err != nil {
		return err
	}

	return nil
}
// PlotEMA compares the calculated EMAs with data obtained from Exch.
func PlotEMA(timestamps []int64, m55 []float64, closingPrices []float64) error {
	// Create a new plot.
	p := plot.New()

	// Add data to the plot.
	emaData := plotter.XYs{}
	for i, timestamp := range timestamps {
		emaData = append(emaData, struct{ X, Y float64 }{float64(timestamp), m55[i]})
	}
	err := plotutil.AddLinePoints(p, "EMA55", emaData)
	if err != nil {
		return fmt.Errorf("error adding EMA55 to plot: %v", err)
	}

	closingPriceData := plotter.XYs{}
	for i, timestamp := range timestamps {
		closingPriceData = append(closingPriceData, struct{ X, Y float64 }{float64(timestamp), closingPrices[i]})
	}
	err = plotutil.AddLinePoints(p, "Closing Prices", closingPriceData)
	if err != nil {
		return fmt.Errorf("error adding closing prices to plot: %v", err)
	}

	// Save the plot to a file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "ema_plot.png"); err != nil {
		return fmt.Errorf("error saving plot: %v", err)
	}

	return nil
}

func PlotRSI(stochRSI, smoothKRSI []float64) {

	// Create a new plot
	p := plot.New()

	// Create line plots for Stochastic RSI and SmoothK RSI points
	stochRSILine, err := plotter.NewLine(createPoints(stochRSI))
	if err != nil {
		panic(err)
	}
	smoothKRSILine, err := plotter.NewLine(createPoints(smoothKRSI))
	if err != nil {
		panic(err)
	}

	// Add the line plots to the plot
	p.Add(stochRSILine, smoothKRSILine)

	// Save the plot to a PNG file
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "stochastic_rsi_plot.png"); err != nil {
		panic(err)
	}
}

func createPoints(data []float64) plotter.XYs {
	pts := make(plotter.XYs, len(data))
	for i, value := range data {
		pts[i].X = float64(i)
		pts[i].Y = value
	}
	return pts
}
