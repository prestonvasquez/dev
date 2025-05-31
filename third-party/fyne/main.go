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
)

// OHLC represents one “candlestick” worth of data.
type OHLC struct {
	Open, High, Low, Close float64
}

func main() {
	// ----------------------------------------------------------------
	// 1) Window setup
	// ----------------------------------------------------------------
	myApp := app.New()
	w := myApp.NewWindow("Dynamic Candlestick Chart")
	w.Resize(fyne.NewSize(600, 400))

	// ----------------------------------------------------------------
	// 2) Graphing parameters
	// ----------------------------------------------------------------
	const numCandles = 30                      // how many candles visible at once
	const graphW, graphH = 500.0, 300.0        // pixel size of the chart area
	const originX, originY = 50.0, 350.0       // bottom-left corner of the chart (in window coords)
	candleSlot := graphW / float64(numCandles) // horizontal “slot” per candle
	candleWidth := candleSlot * 0.6            // width of each candle body

	// Seed the RNG once:
	rand.Seed(time.Now().UnixNano())

	// ----------------------------------------------------------------
	// 3) Generate initial random OHLC data
	// ----------------------------------------------------------------
	// Start from an arbitrary base price, e.g. 100.0
	basePrice := 100.0
	ohlcWindow := make([]OHLC, numCandles)
	for i := 0; i < numCandles; i++ {
		ohlcWindow[i] = randomCandle(basePrice)
		basePrice = ohlcWindow[i].Close
	}

	// ----------------------------------------------------------------
	// 4) Pre–allocate shapes for “wick” (line) and “body” (rectangle)
	// ----------------------------------------------------------------
	wicks := make([]*canvas.Line, numCandles)
	bodies := make([]*canvas.Rectangle, numCandles)

	// Create a container with no auto layout (absolute coords)
	content := container.NewWithoutLayout()

	// Create static axes (x-axis and y-axis)
	xAxis := canvas.NewLine(color.Gray{Y: 100})
	xAxis.StrokeWidth = 1
	xAxis.Position1 = fyne.NewPos(float32(originX), float32(originY))
	xAxis.Position2 = fyne.NewPos(float32(originX+graphW), float32(originY))
	content.Add(xAxis)

	yAxis := canvas.NewLine(color.Gray{Y: 100})
	yAxis.StrokeWidth = 1
	yAxis.Position1 = fyne.NewPos(float32(originX), float32(originY-graphH))
	yAxis.Position2 = fyne.NewPos(float32(originX), float32(originY))
	content.Add(yAxis)

	// Function to map a price in [priceMin, priceMax] → a y‐coordinate in the chart
	mapPriceToY := func(price, priceMin, priceMax float64) float64 {
		// Normalize price into [0,1], then map to [originY-graphH .. originY]:
		if priceMax == priceMin {
			// edge‐case: avoid division by zero
			return originY - graphH/2.0
		}
		normalized := (price - priceMin) / (priceMax - priceMin)
		// y grows downward, so top of chart = originY - graphH, bottom = originY.
		return originY - normalized*graphH
	}

	// Initialize the candle shapes based on initial ohlcWindow
	updateShapes := func() {
		// 1) Find sliding‐window min & max (for vertical scaling)
		priceMin := math.Inf(1)
		priceMax := math.Inf(-1)
		for _, c := range ohlcWindow {
			if c.Low < priceMin {
				priceMin = c.Low
			}
			if c.High > priceMax {
				priceMax = c.High
			}
		}

		// 2) For each candle, compute wick‐line and body‐rectangle coords & color
		for i, candle := range ohlcWindow {
			// Horizontal x‐coordinates for this candle
			xCenter := originX + float64(i)*candleSlot + candleSlot/2.0
			// “Wick” is a vertical line from High→Low
			yHigh := mapPriceToY(candle.High, priceMin, priceMax)
			yLow := mapPriceToY(candle.Low, priceMin, priceMax)
			line := wicks[i]
			if line == nil {
				// Create a new wick‐line if not already built
				line = canvas.NewLine(color.Black)
				line.StrokeWidth = 1
				wicks[i] = line
				content.Add(line)
			}
			line.Position1 = fyne.NewPos(float32(xCenter), float32(yHigh))
			line.Position2 = fyne.NewPos(float32(xCenter), float32(yLow))
			canvas.Refresh(line)

			// “Body” is a rectangle between Open→Close
			yOpen := mapPriceToY(candle.Open, priceMin, priceMax)
			yClose := mapPriceToY(candle.Close, priceMin, priceMax)

			// Top of the body is max(Open,Close), bottom is min(Open,Close)
			yTop := math.Min(yOpen, yClose)
			yBottom := math.Max(yOpen, yClose)
			bodyHeight := yBottom - yTop
			body := bodies[i]
			if body == nil {
				body = canvas.NewRectangle(color.Transparent)
				body.StrokeWidth = 1
				bodies[i] = body
				content.Add(body)
			}
			// Move and resize
			body.Move(fyne.NewPos(float32(xCenter-candleWidth/2.0), float32(yTop)))
			body.Resize(fyne.NewSize(float32(candleWidth), float32(bodyHeight)))

			// Fill‐color: green if Close ≥ Open, red otherwise
			if candle.Close >= candle.Open {
				body.FillColor = color.NRGBA{R: 34, G: 139, B: 34, A: 255} // dark green
				body.StrokeColor = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
			} else {
				body.FillColor = color.NRGBA{R: 178, G: 34, B: 34, A: 255} // firebrick red
				body.StrokeColor = color.NRGBA{R: 139, G: 0, B: 0, A: 255}
			}
			canvas.Refresh(body)
		}
	}

	// Build initial shapes
	updateShapes()

	// Tell the window to use our “content” container
	w.SetContent(content)

	// ----------------------------------------------------------------
	// 5) Animate: every 100 ms, drop the oldest candle, generate a new random one,
	//    shift the window, and redraw.
	// ----------------------------------------------------------------
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			// 5.1) Remove first element and shift the slice left
			ohlcWindow = ohlcWindow[1:]

			// 5.2) Generate a new candle based on the previous close
			prevClose := ohlcWindow[len(ohlcWindow)-1].Close
			newCandle := randomCandle(prevClose)

			// 5.3) Append the new candle to the end
			ohlcWindow = append(ohlcWindow, newCandle)

			// 5.4) Use fyne.Do to ensure UI‐thread safety
			fyne.Do(func() {
				updateShapes()
			})
		}
	}()

	// 6) Show the window (blocking call)
	w.ShowAndRun()
}

// randomCandle generates a plausible random OHLC candle given an “open” price.
// - open = prevClose
// - high = open + rand*volatility
// - low = open – rand*volatility
// - close is chosen randomly between low and high
func randomCandle(prevClose float64) OHLC {
	// Choose a small “volatility” relative to the previous price
	volatility := prevClose * 0.02 // ±2%
	// high = open + Δ
	hi := prevClose + rand.Float64()*volatility
	// low = open – Δ
	lo := prevClose - rand.Float64()*volatility
	// close = somewhere between lo and hi
	cl := lo + rand.Float64()*(hi-lo)
	return OHLC{
		Open:  prevClose,
		High:  hi,
		Low:   lo,
		Close: cl,
	}
}
