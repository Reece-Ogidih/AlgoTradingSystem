package histdata // I made this a separate package since its not apart of main package

import (
	"fmt"      // Standard formatting and printing library
	"io"       // To read/write the files
	"net/http" // To make Requests
	"time"     // To convert date into required format for CoinGecko's API
)

// My approach here is to set the base url and then we can add the required endpoint based upon which timeframe we choose.
// Crytpoworld generally runs on UTC, example of how we could make a boundary for time would be:
// example := time.Date(2023, 1, 1, 0, 0, 00, time.UTC)
// Date function has parameters: year, month, day, hour, minute, sec, nanosec, timezone.

func GetSolHist(from, to time.Time) ([]byte, error) {
	base := "https://api.coingecko.com/api/v3/coins/solana/market_chart/range"
	url := fmt.Sprintf("%s?vs_currency=usd&from=%d&to=%d",
		base,
		from.Unix(),
		to.Unix(),
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
