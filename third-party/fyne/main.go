package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// OHLC represents one candlestick’s price data.
type OHLC struct {
	Open, High, Low, Close float64
}

func main() {
	// ----------------------------------------
	// 1) App and Window setup
	// ----------------------------------------
	myApp := app.New()
	w := myApp.NewWindow("Static Candlestick + Button Demo")
	w.Resize(fyne.NewSize(600, 450))

	// ----------------------------------------
	// 2) Chart parameters
	// ----------------------------------------
	const numCandles = 5                       // number of candles
	const graphW, graphH = 500.0, 300.0        // pixel size of chart area
	const originX, originY = 50.0, 350.0       // bottom-left of chart in window coords
	candleSlot := graphW / float64(numCandles) // horizontal slot per candle
	candleWidth := candleSlot * 0.6            // width of candle body (60% of slot)

	// ----------------------------------------
	// 3) Generate initial OHLC data (static)
	// ----------------------------------------
	rand.Seed(time.Now().UnixNano())
	ohlcData := generateRandomOHLC(numCandles)

	// ----------------------------------------
	// 4) Prepare shapes & container for chart
	// ----------------------------------------
	// Pre-allocate slices for wick (line) and body (rectangle) shapes
	wicks := make([]*canvas.Line, numCandles)
	bodies := make([]*canvas.Rectangle, numCandles)

	// A container without layout allows absolute positioning
	chartContainer := container.NewWithoutLayout()

	// Static axes: X‐axis (horizontal) and Y‐axis (vertical)
	xAxis := canvas.NewLine(color.Gray{Y: 128})
	xAxis.StrokeWidth = 1
	xAxis.Position1 = fyne.NewPos(float32(originX), float32(originY))
	xAxis.Position2 = fyne.NewPos(float32(originX+graphW), float32(originY))
	chartContainer.Add(xAxis)

	yAxis := canvas.NewLine(color.Gray{Y: 128})
	yAxis.StrokeWidth = 1
	yAxis.Position1 = fyne.NewPos(float32(originX), float32(originY-graphH))
	yAxis.Position2 = fyne.NewPos(float32(originX), float32(originY))
	chartContainer.Add(yAxis)

	// Helper: map a price in [minPrice, maxPrice] → y‐coordinate on canvas
	mapPriceToY := func(price, minPrice, maxPrice float64) float64 {
		if maxPrice == minPrice {
			// avoid division by zero
			return originY - graphH/2.0
		}
		normalized := (price - minPrice) / (maxPrice - minPrice)
		// Top of chart is originY - graphH; bottom is originY
		return originY - normalized*graphH
	}

	// drawChart updates or creates shapes based on current ohlcData
	drawChart := func() {
		// 1) Find window’s min and max prices
		minPrice := math.Inf(1)
		maxPrice := math.Inf(-1)
		for _, c := range ohlcData {
			if c.Low < minPrice {
				minPrice = c.Low
			}
			if c.High > maxPrice {
				maxPrice = c.High
			}
		}

		// 2) For each candle, compute wick and body coordinates + color
		for i, candle := range ohlcData {
			// X-center of this candle slot
			xCenter := originX + float64(i)*candleSlot + candleSlot/2.0

			// Wick: Line from High → Low
			yHigh := mapPriceToY(candle.High, minPrice, maxPrice)
			yLow := mapPriceToY(candle.Low, minPrice, maxPrice)
			var line *canvas.Line
			if wicks[i] == nil {
				line = canvas.NewLine(color.Black)
				line.StrokeWidth = 1
				wicks[i] = line
				chartContainer.Add(line)
			} else {
				line = wicks[i]
			}
			line.Position1 = fyne.NewPos(float32(xCenter), float32(yHigh))
			line.Position2 = fyne.NewPos(float32(xCenter), float32(yLow))
			canvas.Refresh(line)

			// Body: Rectangle between Open → Close
			yOpen := mapPriceToY(candle.Open, minPrice, maxPrice)
			yClose := mapPriceToY(candle.Close, minPrice, maxPrice)

			// Top is min(yOpen, yClose); bottom is max(yOpen, yClose)
			yTop := math.Min(yOpen, yClose)
			yBottom := math.Max(yOpen, yClose)
			height := yBottom - yTop

			var rect *canvas.Rectangle
			if bodies[i] == nil {
				rect = canvas.NewRectangle(color.Transparent)
				rect.StrokeWidth = 1
				bodies[i] = rect
				chartContainer.Add(rect)
			} else {
				rect = bodies[i]
			}
			rect.Move(fyne.NewPos(float32(xCenter-candleWidth/2.0), float32(yTop)))
			rect.Resize(fyne.NewSize(float32(candleWidth), float32(height)))

			// Fill color: green if Close ≥ Open; red otherwise
			if candle.Close >= candle.Open {
				rect.FillColor = color.NRGBA{R: 34, G: 139, B: 34, A: 255} // dark green
				rect.StrokeColor = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
			} else {
				rect.FillColor = color.NRGBA{R: 178, G: 34, B: 34, A: 255} // firebrick red
				rect.StrokeColor = color.NRGBA{R: 139, G: 0, B: 0, A: 255}
			}
			canvas.Refresh(rect)
		}
	}

	// Draw the initial chart
	drawChart()

	// ----------------------------------------
	// 5) Create a button that “does something”
	//    In this case, it generates a brand‐new random OHLC dataset
	//    and redraws the chart.
	// ----------------------------------------
	button := widget.NewButton("Randomize Data", func() {
		ohlcData = generateRandomOHLC(numCandles)
		drawChart()
	})

	// Place the button above the chart using a Border layout
	content := container.NewBorder(button, nil, nil, nil, chartContainer)
	w.SetContent(content)
	w.ShowAndRun()
}

// generateRandomOHLC returns a slice of count random OHLC candles.
// We start at some base price (e.g. 100.0) and do a small random walk.
func generateRandomOHLC(count int) []OHLC {
	out := make([]OHLC, count)
	base := 100.0
	for i := 0; i < count; i++ {
		vol := base * 0.02 // ±2% volatility
		high := base + rand.Float64()*vol
		low := base - rand.Float64()*vol
		closePrice := low + rand.Float64()*(high-low)
		out[i] = OHLC{
			Open:  base,
			High:  high,
			Low:   low,
			Close: closePrice,
		}
		base = closePrice
	}
	return out
}
