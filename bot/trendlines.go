package bot

import (
	"context"
	"log"

	models "github.com/Reece-Ogidih/CT-Bot/Models"
)

// My approach to Trendline calculation is to generate a sliding window, and to calculate the trendlines within that window.
// The function is returning two lines, the type has been defined in types.go
// We first need to define a helper function that takes in a slice of candles and calculates the trendlines
func computeTrendLines(window []models.CandleStick) bool { // (supLine, resLine models.Trendline) {
	// Here will need to implement logic into choosing the anchor points from the window and constructing the support and resistence trenlines.
	return true
}

func DetectTrendLines(candles []models.CandleStick, windowsize int) bool { // (supLine, resLine models.Trendline) {
	ctx := context.Background()
	candleStream, err := FetchLive(ctx, "SOLUSDT", "1m")
	if err != nil {
		log.Fatal(err)
	}

	// Will need to define the sliding window
	var window []models.CandleStick

	for candle := range candleStream {
		if !candle.IsFinal { // Only want to be calculating on closed candles
			continue
		}

		// Add the candle to the window
		window = append(window, candle)

		// To keep the window at a fixed size, remove the oldest candle when a new one arrives
		if len(window) > windowsize {
			window = window[1:]
		}

		// Need a check for when there are not enough candles yet
		if len(window) < windowsize {
			continue // Since there isnt enough data for the trendline yet
		}

		// Now the window is ready and trendlines can be calculated

	}
	return true
}
