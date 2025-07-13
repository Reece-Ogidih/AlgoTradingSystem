package bot

import (
	models "github.com/Reece-Ogidih/CT-Bot/Models"
)

// My approach to Trendline calculation is to generate a sliding window, and to calculate the trendlines within that window.
// The function is returning two lines, the type has been defined in types.go
// We first need to define two helper functions
// One will check and compute the MSE of a trendline and the close prices
// The other will take in a slice of candles and calculates the trendlines
//func checkTrendlines(support bool, pivot int, slope float64, y []models.CandleStick) err float64 {}

func computeTrendLines(window []models.CandleStick) bool { // (supLine, resLine models.Trendline) {
	// Here will need to implement logic into choosing the anchor points from the window and constructing the support and resistence trenlines.
	return true
}
