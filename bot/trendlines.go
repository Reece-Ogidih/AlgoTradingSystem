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

	// Now define the anchor point for normalising.
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

// This is the end of what will be used for live bot

// Now need to basically duplicate everything for the historical simullation run to dev ML model

// func (sw *SlidingWindowTrain) CheckTrendline1(t models.Trendline, startIdx, endIdx int, isResist bool) int {
// 	// First guard against edge case
// 	if startIdx >= endIdx {
// 		return -1
// 	}

// 	// Now define the anchor point for normalising.
// 	//anchor1 := sw.Candles[startIdx]
// 	for i := endIdx - 1; i >= startIdx; i-- {
// 		candle := sw.Candles[i]

// 		// Can use OpenTime as x axis
// 		x := float64(candle.OpenTime)

// 		// Can now get the expected value of the price
// 		expected := t.Gradient*x + t.Intercept

// 		if isResist {
// 			price := candle.High
// 			if price > expected {
// 				return i
// 			}
// 		} else {
// 			price := candle.Low
// 			if price < expected {
// 				return i
// 			}
// 		}
// 	}
// 	return -1
// }

func (sw *SlidingWindowTrain) CheckTrendline1(t models.Trendline, startIdx, endIdx int, isResist bool) int {
	// Guard against edge case
	if startIdx >= endIdx {
		return -1
	}

	for i := endIdx - 1; i >= startIdx; i-- {
		// Instead of OpenTime, use index from sw.Idxs
		x := float64(sw.Idxs[i])

		expected := t.Gradient*x + t.Intercept

		if isResist {
			if sw.Candles[i].High > expected {
				return i
			}
		} else {
			if sw.Candles[i].Low < expected {
				return i
			}
		}
	}
	return -1
}

func (sw *SlidingWindowTrain) MaxHigh() int {
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

func (sw *SlidingWindowTrain) MinLow() int {
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

func (sw *SlidingWindowTrain) CreateResLine() (models.Trendline, bool) {
	maxIdx := sw.MaxHigh()
	if maxIdx == -1 {
		return models.Trendline{}, false
	}
	anchor1 := sw.Candles[maxIdx]
	endIdx := len(sw.Candles) - 1 // Decided to use the most recent completed candle
	anchor2 := sw.Candles[endIdx]

	for {
		// Find positions of anchor1 and anchor2 inside the window
		pos1 := -1
		pos2 := -1
		for i, c := range sw.Candles {
			if c.OpenTime == anchor1.OpenTime {
				pos1 = i
			}
			if c.OpenTime == anchor2.OpenTime {
				pos2 = i
			}
		}

		// Will check to ensure that both of the candles are within the window
		if pos1 == -1 || pos2 == -1 {
			break
		}

		// Cannot have identical timestamps (Note: sw.Idxs[pos1] = index of anchor1 and sw.Idx[pos2] = index of anchor2)
		if sw.Idxs[pos1] == sw.Idxs[pos2] {
			break
		}

		// Need to ensure that the line has negative gradient
		if anchor2.High >= anchor1.High {
			// To guard against infinite loop, have to make sure to update the anchor2 before exiting the loop
			if pos2-1 <= pos1 {
				break // There are no more candles left to check
			}
			anchor2 = sw.Candles[pos2-1] // Can now update the 2nd anchor and skip to the next loop
			continue
		}

		// Will get the required points to calculate equation of the line
		x1 := float64(sw.Idxs[pos1])
		x2 := float64(sw.Idxs[pos2])
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

// Want to use index positions instead of OpenTime for the timestamp. this is because it would otherwise result in practically 0 gradient
func (sw *SlidingWindowTrain) CreateSupLine() (models.Trendline, bool) {
	minIdx := sw.MinLow()
	if minIdx == -1 {
		return models.Trendline{}, false
	}
	anchor1 := sw.Candles[minIdx]
	endIdx := len(sw.Candles) - 1 // Again, use the most recent completed candle to start
	anchor2 := sw.Candles[endIdx]

	for {
		// Find positions of anchor1 and anchor2 inside the window
		pos1 := -1
		pos2 := -1
		for i, c := range sw.Candles {
			if c.OpenTime == anchor1.OpenTime {
				pos1 = i
			}
			if c.OpenTime == anchor2.OpenTime {
				pos2 = i
			}
		}

		// Will check to ensure that both of the candles are within the window
		if pos1 == -1 || pos2 == -1 {
			break
		}

		// Cannot have identical timestamps (Note: sw.Idxs[pos1] = index of anchor1 and sw.Idx[pos2] = index of anchor2)
		if sw.Idxs[pos1] == sw.Idxs[pos2] {
			break
		}

		// Need to ensure that the line has positive gradient
		if anchor2.Low <= anchor1.Low {
			// To guard against infinite loop, have to make sure to update the anchor2 before exiting the loop
			if pos2-1 <= pos1 {
				break // There are no more candles left to check
			}
			anchor2 = sw.Candles[pos2-1] // Can now update the 2nd anchor and skip to the next loop
			continue
		}

		// Will get the required points to calculate equation of the line
		x1 := float64(sw.Idxs[pos1])
		x2 := float64(sw.Idxs[pos2])
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
// func (sw *SlidingWindowTrain) CheckTrenlines2(line models.Trendline, candle models.EnrichedCandle, isResist bool, currPos int) bool {
// 	// First get the OpenTime as the x coordinate, and trendline price as y coordinate
// 	x2 := float64(candle.OpenTime)
// 	y2 := line.Gradient*x2 + line.Intercept

// 	// Can now break the logic into a break in either support or resitance trendlines

// 	// If not already in a trade, only need to do standard check
// 	if currPos == 0 {
// 		if isResist {
// 			// For resistance trendline, breakout would be above the trendline price (since it is upper limit)
// 			return candle.High > y2
// 		} else {
// 			// For support trendline, breakout would be below the trendline price (since it is lower limit)
// 			return candle.Low < y2
// 		}
// 	}
// 	// In a long → only support breaks matter
// 	if currPos == 1 {
// 		if !isResist {
// 			return candle.Low < y2
// 		}
// 		return false
// 	}
// 	// In a short → only resistance breaks matter
// 	if currPos == -1 {
// 		if isResist {
// 			return candle.High > y2
// 		}
// 		return false
// 	}
// 	// If no breakouts are detected can return false
// 	return false
// }

func (sw *SlidingWindowTrain) CheckTrendlines2(line models.Trendline, candle models.EnrichedCandle, isResist bool, currPos int, nextIdx int) bool {
	// This function is for the live candles being fed in, so we know that the Idx would end up being the one more than the last Idx in the window
	x := float64(nextIdx)
	y := line.Gradient*x + line.Intercept

	if currPos == 0 {
		if isResist {
			return candle.High > y
		}
		return candle.Low < y
	}
	if currPos == 1 {
		if !isResist {
			return candle.Low < y
		}
		return false
	}
	if currPos == -1 {
		if isResist {
			return candle.High > y
		}
		return false
	}
	return false
}

// Will be easier to use in the bot with an additional function to compute the two trendlines
func (sw *SlidingWindowTrain) ComputeTrendlines() {
	if line, ok := sw.CreateResLine(); ok {
		sw.ResLine = line
	}
	if line, ok := sw.CreateSupLine(); ok {
		sw.SupLine = line
	}
}
