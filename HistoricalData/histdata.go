package histdata // I made this a separate package since its not apart of main package

import (
	"encoding/json"
	"fmt"      // Standard formatting and printing library
	"io"       // To read/write the files
	"net/http" // To make Requests
	"strconv"
	"time" // To convert date into required format for CoinGecko's API
)

// First I define type of candlestick which then has the data I pull from each candlestick through Binance API
// Then define the overall dataset to be a collection of these candlesticks
type CandleStick struct {
	OpenTime    int64
	Open        float64
	High        float64
	Low         float64
	Close       float64
	Volume      float64
	NumOfTrades int64
	CloseTime   int64
}

type Dataset struct {
	Candles []CandleStick
}

// Since Binance limits to 1000 candlesticks per call, will need to iteratively loop through sets of 1000 candlesticks
// Time-wise we start with 1 year of data since it is short term day trading bot, if model doesnt seem robust can extend to 2 years.
// Also I can not have smaller intervals than 1min so I just use that here

func FetchCandles(start, end time.Time) (Dataset, error) {
	startMillis := start.UnixNano() / int64(time.Millisecond) //Converting into the required format for Binance API (so in milliseconds)
	endMillis := end.UnixNano() / int64(time.Millisecond)     // int64(time.Milliseconds) gives number of nanoseconds in a millisecond

	var dataset Dataset

	const maxCandlesPerCall = int64(1000) // This is the limit imposed by free Binance API calls
	const intervalMillis = int64(60 * 1000)
	const interval = "1m"

	chunkMillis := maxCandlesPerCall * intervalMillis // Number of millisecs covered per API request
	totalMillis := endMillis - startMillis            // Number of millisecs over entire duration

	numChunks := totalMillis / chunkMillis
	if totalMillis%chunkMillis != 0 { // Add an if check for if there is a remainder
		numChunks++
	}

	// Now loop iteratively over the time period to get all the candles
	for i := int64(0); i < numChunks; i++ {
		chunkStart := startMillis + i*chunkMillis
		chunkEnd := chunkStart + chunkMillis - 1
		if chunkEnd > endMillis { // Add the end case
			chunkEnd = endMillis
		}

		url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=SOLUSDT&interval=%s&startTime=%d&endTime=%d&limit=%d",
			interval,
			chunkStart,
			chunkEnd,
			maxCandlesPerCall,
		)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error fetching candles", err)
			continue // Used continue here so that it will just skip this iteration
		}
		defer resp.Body.Close() // Need to close at end to prevent memory leak

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			continue
		}

		// Now we need to parse the JSON response into [][]interface{}
		// This is basically just converting the Binance output into a datastructure that is viable for Go.
		var rawData [][]interface{}
		err = json.Unmarshal(body, &rawData)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			continue
		}

		// For each candle, need to extract the values we want and then append them to the dataset:
		for _, item := range rawData {
			openTime := int64(item[0].(float64))
			open, _ := strconv.ParseFloat(item[1].(string), 64)
			high, _ := strconv.ParseFloat(item[2].(string), 64)
			low, _ := strconv.ParseFloat(item[3].(string), 64)
			closePrice, _ := strconv.ParseFloat(item[4].(string), 64)
			volume, _ := strconv.ParseFloat(item[5].(string), 64)
			closeTime := int64(item[6].(float64))
			numTrades := int64(item[8].(float64))

			candle := CandleStick{
				OpenTime:    openTime,
				Open:        open,
				High:        high,
				Low:         low,
				Close:       closePrice,
				Volume:      volume,
				NumOfTrades: numTrades,
				CloseTime:   closeTime,
			}

			dataset.Candles = append(dataset.Candles, candle)
		}
	}
	return dataset, nil
}

// My approach here is to set the base url and then we can add the required endpoint based upon which timeframe we choose.
// Crytpoworld generally runs on UTC, example of how we could make a boundary for time would be:
// example := time.Date(2023, 1, 1, 0, 0, 00, time.UTC)
// Date function has parameters: year, month, day, hour, minute, sec, nanosec, timezone.

// func GetSolHist(from, to time.Time) ([]byte, error) {
// 	base := "https://api.coingecko.com/api/v3/coins/solana/market_chart/range" // Will have to change Data Provider due to limitations of granuity
// 	url := fmt.Sprintf("%s?vs_currency=usd&from=%d&to=%d",
// 		base,
// 		from.Unix(),
// 		to.Unix(),
// 	)
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return body, nil
// }
