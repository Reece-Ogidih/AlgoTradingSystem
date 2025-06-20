package histdata        // I made this a separate package since its not apart of main package

import (
	"log"               // For debugging
	"net/http"          // To make Requests
	"io"                // To read/write the files
	"time"              // To convert date into required format for CoinGecko's API 
)

func GetSolHist() {
	base := "https://www.coingecko.com/en/coins/solana/market_chart/range"
	url := 
}

