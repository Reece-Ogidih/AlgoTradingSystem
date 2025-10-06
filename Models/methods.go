package models

import (
	"time"
)

// This function is for the construction of the live candles
func (c *CandleStick) Update(price float64,
	deltaVol float64, deltaTrade int64) {

	// Will first check for new limits in candle window
	if price > c.High {
		c.High = price
	}

	if price < c.Low {
		c.Low = price
	}

	// Now can update the remaining information
	c.Close = price
	c.Volume += deltaVol
	c.NumOfTrades += deltaTrade
	c.CloseTime = time.Now().Unix()
}
