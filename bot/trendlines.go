package bot

import (
	models "github.com/Reece-Ogidih/CT-Bot/Models"
)

// My approach to Trendline calculation is to generate a sliding window, and to calculate the trendlines within that window.
// The functions are returning two lines, the type has been defined in types.go

// It will work as follows
// First an initial trendline is generated on the sliding window using CreateResLine() or CreateSupLine
// The line created will be dictated by direction of trend (ie +DI or -DI)
// They will use CheckTrendline1 to ensure no past violations of the trendline
// Now that the trendline is active, any new candles will be assessed individually to look for breakout using CheckTrenlines2 fuction
// Once a breakout is detected, new trendline will be created (ie support line -> resistance line)

// I named this CheckTrendline1 since it is going to check for a breakout in trendline over entire window, this is for trendline generation
func (sw *SlidingWindow) CheckTrendline1(t models.Trendline, startIdx, endIdx int, isResist bool) int {
	// First guard against edge case
	if startIdx >= endIdx {
		return -1
	}

	// Now define the anchor point for normalising. FOUND CANNOT WORK SINCE WILL SKEW THE EXPECTED PRICE VALUE
	//anchor1 := sw.Candles[startIdx]
	for i := endIdx - 1; i >= startIdx; i-- {
		candle := sw.Candles[i]

		// Can use OpenTime as x axis
		x := float64(candle.OpenTime)

		// Can now get the expected value of the price
		expected := t.Gradient*x + t.Intercept

		if isResist {
			price := candle.High
			if price > expected {
				return i
			}
		} else {
			price := candle.Low
			if price < expected {
				return i
			}
		}
	}
	return -1
}

// Next need some helper functions to find the max high or max low, it will output the index of this candle
func (sw *SlidingWindow) MaxHigh() int {
	maxIdx := -1
	maxHigh := -1.0
	for i := 0; i < len(sw.Candles); i++ {
		if sw.Candles[i].High > maxHigh {
			maxHigh = sw.Candles[i].High
			maxIdx = i
		}
	}
	return maxIdx
}

func (sw *SlidingWindow) MinLow() int {
	minIdx := -1
	minLow := 10000000000000000000000.0
	for i := 0; i < len(sw.Candles); i++ {
		if sw.Candles[i].Low < minLow {
			minLow = sw.Candles[i].Low
			minIdx = i
		}
	}
	return minIdx
}

func (sw *SlidingWindow) CreateResLine() (models.Trendline, bool) {
	maxIdx := sw.MaxHigh()
	if maxIdx == -1 {
		return models.Trendline{}, false
	}
	anchor1 := sw.Candles[maxIdx]
	endIdx := len(sw.Candles) - 1 // Decided to use the most recent completed candle
	anchor2 := sw.Candles[endIdx]

	for {
		// Cannot have identical timestamps
		if anchor2.OpenTime == anchor1.OpenTime {
			break
		}

		// Will get the required points to calculate equation of the line
		x1 := float64(anchor1.OpenTime)
		x2 := float64(anchor2.OpenTime)
		y1 := anchor1.High
		y2 := anchor2.High

		// Can now calculate the gradient and intercept of the Resistance Line
		gradient := (y2 - y1) / (x2 - x1)
		intercept := y1 - gradient*x1

		// Define the line
		line := models.Trendline{
			Gradient:  gradient,
			Intercept: intercept,
			A1Time:    anchor1.OpenTime,
			A2Time:    anchor2.OpenTime,
			A1Price:   y1,
			A2Price:   y2,
		}

		// Now we can check for any breakouts on past candles and recalculate the line
		breakoutIdx := sw.CheckTrendline1(line, maxIdx+1, endIdx, true)

		// If no breakouts then the line is complete
		if breakoutIdx == -1 {
			return line, true
		}

		// Otherwise we shift anchor 2 to the position of the breakout
		anchor2 = sw.Candles[breakoutIdx]
		endIdx = breakoutIdx
	}

	return models.Trendline{}, false
}

func (sw *SlidingWindow) CreateSupLine() (models.Trendline, bool) {
	minIdx := sw.MinLow()
	if minIdx == -1 {
		return models.Trendline{}, false
	}
	anchor1 := sw.Candles[minIdx]
	endIdx := len(sw.Candles) - 1 // Again, use the most recent completed candle
	anchor2 := sw.Candles[endIdx]

	for {
		// Cannot have identical timestamps
		if anchor2.OpenTime == anchor1.OpenTime {
			break
		}

		// Will get the required points to calculate equation of the line
		x1 := float64(anchor1.OpenTime)
		x2 := float64(anchor2.OpenTime)
		y1 := anchor1.Low
		y2 := anchor2.Low

		// Can now calculate the gradient and intercept of the Resistance Line
		gradient := (y2 - y1) / (x2 - x1)
		intercept := y1 - gradient*x1

		// Define the line
		line := models.Trendline{
			Gradient:  gradient,
			Intercept: intercept,
			A1Time:    anchor1.OpenTime,
			A2Time:    anchor2.OpenTime,
			A1Price:   y1,
			A2Price:   y2,
		}

		// Now we can check for any breakouts on past candles and recalculate the line
		breakoutIdx := sw.CheckTrendline1(line, minIdx+1, endIdx, false)

		// If no breakouts then the line is complete
		if breakoutIdx == -1 {
			return line, true
		}

		// Otherwise we shift anchor 2 to the position of the breakout
		anchor2 = sw.Candles[breakoutIdx]
		endIdx = breakoutIdx
	}

	return models.Trendline{}, false
}

// Now create CheckTrendlines2 which will check 1 candle for a breakout, this will be called as each new candle arrives live.
func (sw *SlidingWindow) CheckTrenlines2(line models.Trendline, candle models.CandleStick, isResist bool) bool {
	// First get the OpenTime as the x coordinate, and trendline price as y coordinate
	x2 := float64(candle.OpenTime)
	y2 := line.Gradient*x2 + line.Intercept

	// Can now break the logic into a break in either support or resitance trendlines
	if isResist {
		// For resistance trendline, breakout would be above the trendline price (since it is upper limit)
		if candle.High > y2 {
			return true
		}
	} else {
		// For support trendline, breakout would be below the trendline price (since it is lower limit)
		if candle.Low < y2 {
			return true
		}
	}

	// If no breakouts are detected can return false
	return false
}
