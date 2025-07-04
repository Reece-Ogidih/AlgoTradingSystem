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

	// Get the channel (receive only) of live candle data
	candleStream, err := bot.FetchLive(ctx, "SOLUSDT", "1m")
	if err != nil {
		log.Fatal(err)
	}

	// Now loop so that for each new entry on channel it will print to terminal.
	for candle := range candleStream {
		fmt.Printf("%+v\n", candle)
	}
}
