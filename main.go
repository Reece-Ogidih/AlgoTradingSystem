package main

import (
	"fmt"
	"time"

	histdata "github.com/Reece-Ogidih/CT-Bot/HistoricalData" // Temporarily checking that the API request works and the data is correctly obtained.
)

func main() {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	data, err := histdata.GetSolHist(from, to)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(data))
}
