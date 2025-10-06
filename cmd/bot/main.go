package main

import (
	"context"
	"fmt"
	"log"

	bot "github.com/Reece-Ogidih/CT-Bot/bot"
)

// For now I will be using this to test that the live data is being correctly obtained
func main() {
	// Create the context object
	ctx := context.Background()
	candleStream, err := bot.FetchSOLUSDT()

	// Fallback method
	// Get the channel (receive only) of live candle data
	if err != nil {
		candleStream, err = bot.FetchLive(ctx, "SOLUSDT", "1m")
		if err != nil {
			log.Fatal(err)
		}
	}
	// Now can add in calculation of ADX
	adxCalc := bot.ADXCalculator{Period: 14, Count: 0}

	// Now loop so that for each new entry on channel it will print to terminal.
	for candle := range candleStream {
		//fmt.Printf("%+v\n", candle) // This is just a checker for if the live candle stream works
		if !candle.IsFinal { // Only want to be calculating once per candle
			continue
		}

		// Now can calculate ADX
		adx, ok := adxCalc.Update(candle)
		if ok {
			fmt.Printf("ADX:%.2f, %+v\n", adx, candle)
		}
	}
}
