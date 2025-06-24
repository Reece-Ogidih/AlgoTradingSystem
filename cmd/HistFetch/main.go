package main

import (
	"fmt"
	"time"

	histdata "github.com/Reece-Ogidih/CT-Bot/HistoricalData"
)

// Time-wise we start with 1 year of data since it is short term day trading bot so should be sufficient, if model doesnt seem robust can extend to 2 years.

func main() {
	from := time.Date(2024, 6, 22, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC)

	data, err := histdata.FetchCandles(from, to)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// To validify everything works properly we use some outputs
	// First we get a look at number of candles we have retrieved (would expect 525600 since there are that many minutes in a year)
	fmt.Printf("Total number of Candles	fetched: %d\n", len(data.Candles))

	// Then we take a snapshot by checking first 5 candles to ensure candle data looks correct
	for i := 0; i < 5 && i < len(data.Candles); i++ {
		fmt.Printf("%v\n", data.Candles[i])
	}
}
