package histdata        // I made this a separate package since its not apart of main package

import (
	"log"               // For debugging
	"net/http"          // To make Requests
	"io"                // To read/write the files
	"time"              // To convert date into required format for CoinGecko's API 
	"fmt"               // Standard formatting and printing library
)

func GetSolHist() {
	base := "https://www.coingecko.com/en/coins/solana/market_chart/range"
	url := fmt.Sprintf("%s?vs_currency=usd&from=%d&to=%%d",
		base,
		from.Unix(),
		to.Unix()
	)
resp, err := http.get(url)
if err != nill {
	return nill, err
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nill {
	return nill, err
}
return body, nill
}

