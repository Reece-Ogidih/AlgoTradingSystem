package bot

import (
	"log"
	"math"

	histdata "github.com/Reece-Ogidih/CT-Bot/HistoricalData"
	models "github.com/Reece-Ogidih/CT-Bot/Models"
)

// First helper function I need is a function to sum a sice of numbers:
func sum(data []float64) float64 {
	total := 0.0
	for _, v := range data {
		total += v
	}
	return total
}

// Looking into the mathematical forumulation, many technical indicator calculations depend on the previous x candles' technical indicators
// As a result, depending on the period, the first n candles of data can not have a value for these technical indicators
// I decided against imputing as this will bring bias into the ML model, so will remove these entries after calculating starting averages

// First start with the helper function for Exponential Moving Average
func Calc_EMA(candles []models.CandleStick, period int) []float64 {
	ema := make([]float64, len(candles))

	// For debugging, will add an if statement to ensure there are enough candles to calculate ema over specified period
	if len(candles) < period {
		return ema
	}

	// The first EMA is just a simple average
	var closesum float64
	for i := 0; i < period; i++ {
		closesum += candles[i].Close
		ema[i] = math.NaN() // Will flag these first candles as NaN for now and remove after
	}
	sma := closesum / float64(period)
	ema[period-1] = sma

	alpha := 2.0 / float64(period)

	// Now compute remaining EMA's
	for i := period; i < len(candles); i++ {
		ema[i] = (candles[i].Close-ema[i-1])*alpha + ema[i-1]
	}

	return ema

}

// Will need to completely rework EMA helper function if I decide to use it

// Now will need a helper functions to calculate ADX for both the historical dataset as well as for the live candle stream

// Start with the calculator for the historical data
// First make the base function that takes input of some candles and the period and calculates the ADX
// Note that this is a lagging indicator and so the N-1 candles wont be able to have an ADX value, but loss of a few candles is negligible
func CalculateADX(candles []models.CandleStick, period int) (
	adxVals []float64,
	plusDIvals []float64,
	minusDIvals []float64,
	prevTR float64,
	prevPosDM float64, // Need all these additional outputs for the Live ADX calculator function below
	prevNegDM float64,
	prevADX float64,
	err error) {

	// Will define the types here
	var posDM, negDM float64
	var posDMs, negDMs []float64
	var TRs []float64
	var ADXs = make([]float64, len(candles))   // Final result
	plusDIvals = make([]float64, len(candles)) // Need these for direction indication
	minusDIvals = make([]float64, len(candles))

	// Due to it being a lagging indicator the first N candles will be used to generate first ADX (where N is the period)
	for i := 1; i < len(candles); i++ {
		prev := candles[i-1]
		curr := candles[i]
		// First up we need to calculate +DI and -DI
		UpMove := curr.High - prev.High
		DownMove := prev.Low - curr.Low

		if UpMove > DownMove && UpMove > 0 {
			posDM = UpMove
		} else {
			posDM = 0
		}
		if DownMove > UpMove && DownMove > 0 {
			negDM = DownMove
		} else {
			negDM = 0
		}

		// Next we calculate TR (True Range)
		TR := math.Max(
			curr.High-curr.Low,
			math.Max( // Have to set it up like this since math.Max() only takes 2 arguments
				math.Abs(curr.High-prev.Close),
				math.Abs(curr.Low-prev.Close)),
		)

		// Store these values
		TRs = append(TRs, TR) // note that this would only have 49 values since start from i= 1
		posDMs = append(posDMs, posDM)
		negDMs = append(negDMs, negDM)
	}

	// Now compute the first smoothed values (note I am titling them as prev since they are the base for the for loop later on)
	prevTR = sum(TRs[0:period])
	prevPosDM = sum(posDMs[0:period])
	prevNegDM = sum(negDMs[0:period])

	// Can use these these to calculate the first +DI, -DI, DX and ADX
	prevplusDI := 100 * (prevPosDM / prevTR)
	prevminusDI := 100 * (prevNegDM / prevTR)
	prevDX := 100 * math.Abs(prevplusDI-prevminusDI) / (prevplusDI + prevminusDI)
	ADXs[period] = prevDX
	plusDIvals[period] = prevplusDI
	minusDIvals[period] = prevminusDI

	// Now can compute the ADX for remaining candles
	for i := period + 1; i < len(TRs); i++ { // Have to use TRs here since it only has 49 entries

		smoothedTR := (prevTR*float64(period-1) + TRs[i]) / float64(period)
		smoothedPosDM := (prevPosDM*float64(period-1) + posDMs[i]) / float64(period)
		smoothedNegDM := (prevNegDM*float64(period-1) + negDMs[i]) / float64(period)

		plusDI := 100 * (smoothedPosDM / smoothedTR)
		minusDI := 100 * (smoothedNegDM / smoothedTR)

		DX := 100 * math.Abs(plusDI-minusDI) / (plusDI + minusDI)
		ADXs[i] = (ADXs[i-1]*float64(period-1) + DX) / float64(period)

		// Update the position
		prevTR = smoothedTR
		prevPosDM = smoothedPosDM
		prevNegDM = smoothedNegDM

		// Store the directional indicators
		plusDIvals[i] = plusDI
		minusDIvals[i] = minusDI
	}
	// Return the output (ADXs[len(ADXs) - 1] is the ADX value of most recent candle)
	return ADXs, plusDIvals, minusDIvals, prevTR, prevPosDM, prevNegDM, ADXs[len(ADXs)-1], nil
}

// To best calculate the ADX for live candle stream will use a method on a custom struct so will declare the struct here
// Did it here rather than types.go since can not define a method on a non-local type
type ADXCalculator struct {
	Period     int
	Count      int
	PrevTR     float64
	PrevPosDM  float64
	PrevNegDM  float64
	PrevADX    float64
	PrevCandle models.CandleStick
}

// Now can calculate the ADX by implementing a method on the custom struct
func (a *ADXCalculator) Update(curr models.CandleStick) (adx float64, ok bool) {
	// First time this loop is run will need to get the ADX using historical (most recent) candles
	if a.Count == 0 {
		histCandles, err := histdata.RecentCandles("SOLUSDT", "1m", 50)
		if err != nil {
			log.Fatal(err)
		}
		if len(histCandles) <= a.Period { // Debugging check for if there are enough candles being retrieved.
			log.Fatalf("not enough candles: expected at least %d, got %d", a.Period+1, len(histCandles))
		}
		a.PrevCandle = histCandles[len(histCandles)-1]

		_, _, _, a.PrevTR, a.PrevPosDM, a.PrevNegDM, a.PrevADX, err = CalculateADX(histCandles, a.Period) // Do not need list of ADX vals
		if err != nil {
			log.Fatal(err)
		}
		a.Count = 1
	}

	UpMove := curr.High - a.PrevCandle.High
	DownMove := a.PrevCandle.Low - curr.Low

	var posDM, negDM float64
	if UpMove > DownMove && UpMove > 0 {
		posDM = UpMove
	}
	if DownMove > UpMove && DownMove > 0 {
		negDM = DownMove
	}

	TR := math.Max(
		curr.High-curr.Low,
		math.Max( // Have to set it up like this since math.Max() only takes 2 arguments
			math.Abs(curr.High-a.PrevCandle.Close),
			math.Abs(curr.Low-a.PrevCandle.Close)),
	)
	a.PrevTR = (a.PrevTR*(float64(a.Period-1)) + TR) / float64(a.Period)
	a.PrevPosDM = (a.PrevPosDM*(float64(a.Period-1)) + posDM) / float64(a.Period)
	a.PrevNegDM = (a.PrevNegDM*(float64(a.Period-1)) + negDM) / float64(a.Period)

	// Now calculate the directional index
	plusDI := 100 * (a.PrevPosDM / a.PrevTR)
	minusDI := 100 * (a.PrevNegDM / a.PrevTR)
	denom := plusDI + minusDI // Adding guards for divide by 0
	if denom == 0 {
		return 0, false
	}
	DX := 100 * math.Abs(plusDI-minusDI) / denom

	a.PrevADX = ((a.PrevADX * float64(a.Period-1)) + DX) / float64(a.Period) // calculate the ADX

	a.PrevCandle = curr // Update the stored candle

	return a.PrevADX, true
}
