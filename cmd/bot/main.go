package main

import (
	"context"
	"fmt"
	"log"

	models "github.com/Reece-Ogidih/CT-Bot/Models"
	bot "github.com/Reece-Ogidih/CT-Bot/bot"
)

// For now I will be using this to test that the live data is being correctly obtained
func main() {
	// Create the context object
	ctx := context.Background()

	// Get the channel (receive only) of live candle data
	candleStream, err := bot.FetchLive(ctx, "SOLUSDT", "1m")
	if err != nil {
		log.Fatal(err)
	}

	// Now can add in calculation of ADX
	adxCalc := bot.ADXCalculator{Period: 14}
	var prev models.CandleStick

	count := 0
	// Now loop so that for each new entry on channel it will print to terminal.
	for candle := range candleStream {
		if !candle.IsFinal { // Only want to be calculating once per candle
			continue
		}

		if count == 0 { // First candle needs to be set as prev in order to use the Update method
			prev = candle
			count = 1
			continue
		}

		// Now can calculate ADX
		adx, ok := adxCalc.Update(prev, candle)
		if ok {
			fmt.Printf("ADX:%.2f, %+v\n", adx, candle)
		}
		prev = candle
	}
}
