package main

import (
	"context"
	"fmt"
	"log"

	bot "github.com/Reece-Ogidih/CT-Bot/bot"
)

// For now I will be using this to test that the live data is being correctly obtained
func main() {
	ctx := context.Background()
	candleStream, err := bot.FetchLive(ctx, "SOLUSDT", "1m")
	if err != nil {
		log.Fatal(err)
	}

	for candle := range candleStream {
		fmt.Printf("%+v\n", candle)
	}
}
