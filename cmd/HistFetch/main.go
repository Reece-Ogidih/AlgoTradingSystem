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

	for i := 0; i < 5 && i < len(data.Candles); i++ {
		fmt.Printf("%v\n", data.Candles[i])
	}
}
