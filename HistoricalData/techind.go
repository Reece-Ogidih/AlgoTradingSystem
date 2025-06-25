package histdata

import (
	"math"
)

// Looking into the mathematical forumulation, many technical indicator calculations depend on the previous x candles' technical indicators
// As a result, depending on the period, the first n candles of data can not have a value for these technical indicators
// I decided against imputing as this will bring bias into the ML model, so will remove these entries after calculating starting averages

// First start with the helper function for Exponential Moving Average
func Calc_EMA(candles []CandleStick, period int) []float64 {
	ema := make([]float64, len(candles))

	// For debugging, will add an if statement to ensure there are enough candles to calculate ema over specified period
	if len(candles) < period {
		return ema
	}

	// The first EMA is just a simple average
	var sum float64
	for i := 0; i < period; i++ {
		sum += candles[i].Close
		ema[i] = math.NaN() // Will flag these first candles as NaN for now and remove after
	}
	sma := sum / float64(period)
	ema[period-1] = sma

	alpha := 2.0 / float64(period)

	// Now compute remaining EMA's
	for i := period; i < len(candles); i++ {
		ema[i] = (candles[i].Close-ema[i-1])*alpha + ema[i-1]
	}

	return ema

}

// Going to look into actually learning trading to a better level and also fully flesh out desired trading bot pipeline
// Specifically, sliding windows and trendline channels as well as drawdown support will be some of the primary trading logic
// ML algorithm will output weighted value between 0 and 1 which can be used by bot when deciding capital to invest into trade.
