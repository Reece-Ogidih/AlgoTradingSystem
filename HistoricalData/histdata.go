package histdata // I made this a separate package since its not apart of main package

import (
	"encoding/json"
	"fmt"      // Standard formatting and printing library
	"io"       // To read/write the files
	"net/http" // To make Requests
	"sort"     // To sort the dataset time-wise
	"strconv"  // To parse the json response
	"sync"     // To ensure all workers finish
	"time"     // To convert date into required format for Binance's API

	models "github.com/Reece-Ogidih/CT-Bot/Models" // To use my declared types (for example Candlestick)
)

// Since Binance limits to 1000 candlesticks per call, will need to iteratively loop through sets of 1000 candlesticks
// Instead of sequentially looping, I speed up the process by doing these calls concurrently using a worker pool.
// So I need to define what vars are needed for each job, this is the start and end time of that chunk (in ms)

type Job struct {
	StartTime int64
	EndTime   int64
}

// To get this data concurrently the flow of work is as follows
// First the workers are initiated, then the jobs are passed into the job channel and the workers will read them.
// After this the workers will begin completing the jobs
// Once all the jobs are queued, the job channel is closed
// Then the system waits for all workers to finish their tasks and pass them to the result channel before closing the result channel
// Finally all the candles will have been obtained and can be appended to a dataset

// I had to create a helper function in order to prevent resource leaks by using defer resp.Body.Close() within the loop
func fetchData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // Can use defer here within the helper function

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// So I first setup the workers which will read the data from Binance, Parse it and then append them
func worker(id int, jobChan <-chan Job, resultChan chan<- []models.CandleStick, wg *sync.WaitGroup, interval string, maxCandlesPercall int64, rateLimiter <-chan time.Time) {
	defer wg.Done()

	for job := range jobChan {
		<-rateLimiter // I put in a rate limiter here so that it aligns with Binance API terms of use
		url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=SOLUSDT&interval=%s&startTime=%d&endTime=%d&limit=%d",
			interval,
			job.StartTime,
			job.EndTime,
			maxCandlesPercall,
		)

		body, err := fetchData(url)
		if err != nil {
			fmt.Printf("Worker %d, error fetching: %v", id, err)
			continue
		}

		// Now we need to parse the JSON response into [][]interface{}
		// This is basically just converting the Binance output into a datastructure that is viable for Go.
		var rawData [][]interface{}
		err = json.Unmarshal(body, &rawData)
		if err != nil {
			fmt.Printf("Worker %d, Error unmarshalling: %v", id, err)
			continue
		}

		// For each candle, need to extract the values we want and then append them to the job's dataset:
		var candles []models.CandleStick
		for _, item := range rawData {
			openTime := int64(item[0].(float64))
			open, _ := strconv.ParseFloat(item[1].(string), 64)
			high, _ := strconv.ParseFloat(item[2].(string), 64)
			low, _ := strconv.ParseFloat(item[3].(string), 64)
			closePrice, _ := strconv.ParseFloat(item[4].(string), 64)
			volume, _ := strconv.ParseFloat(item[5].(string), 64)
			closeTime := int64(item[6].(float64))
			numTrades := int64(item[8].(float64))
			// No need to worry about the technical indicators, Go will automatically fill them with 0vals

			candles = append(candles, models.CandleStick{
				OpenTime:    openTime,
				Open:        open,
				High:        high,
				Low:         low,
				Close:       closePrice,
				Volume:      volume,
				CloseTime:   closeTime,
				NumOfTrades: numTrades,
				IsFinal:     true, // all historical candles are closed
			})
		}

		resultChan <- candles
	}
}

// Now I create the overall system using the worker pool
// I can not have smaller intervals than 1min so I just use that here, ideally would use 5s candles.
func FetchCandles(start, end time.Time) (models.Dataset, error) {
	startMillis := start.UnixNano() / int64(time.Millisecond) //Converting into the required format for Binance API (so in milliseconds)
	endMillis := end.UnixNano() / int64(time.Millisecond)     // int64(time.Milliseconds) gives number of nanoseconds in a millisecond

	const maxCandlesPerCall = int64(1000) // This is the limit imposed by free Binance API calls
	const intervalMillis = int64(60 * 1000)
	const interval = "1m"
	const numWorkers = 5 // This is the number of workers I am using, since I am using a rate limiter of 15 per sec, more workers were observed to be unnecessary

	chunkMillis := maxCandlesPerCall * intervalMillis // Number of millisecs covered per API request
	totalMillis := endMillis - startMillis            // Number of millisecs over entire duration

	numChunks := totalMillis / chunkMillis
	if totalMillis%chunkMillis != 0 { // Add an if check for if there is a remainder
		numChunks++
	}

	var wg sync.WaitGroup

	// Create the channels

	jobChan := make(chan Job, numChunks)
	resultChan := make(chan []models.CandleStick, numChunks)

	rateLimit := 15 // To be safe, I chose a rate limit of 15 calls per sec
	rateLimiterChan := time.Tick(time.Second / time.Duration(rateLimit))

	// Initiate the workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobChan, resultChan, &wg, interval, maxCandlesPerCall, rateLimiterChan)
	}

	// Now we start to queue the jobs
	for i := int64(0); i < numChunks; i++ {
		chunkStart := startMillis + i*chunkMillis
		chunkEnd := chunkStart + chunkMillis - 1
		if chunkEnd > endMillis { // Add the end case
			chunkEnd = endMillis
		}

		// Want to actually push each chunk as a "job" and queue it on the Job channel
		jobChan <- Job{StartTime: chunkStart, EndTime: chunkEnd}
	}

	close(jobChan) // Close the queue once all of the jobs are queued

	// Need to ensure all workers are done before closing the results channel
	wg.Wait()
	close(resultChan)

	// Now we can collect all the candles
	var allCandles []models.CandleStick
	for candles := range resultChan {
		allCandles = append(allCandles, candles...)
	}

	// Because the workers are not guarenteed to finish in order, need to sort the candles by opentime to get the data in time order
	sort.Slice(allCandles, func(i, j int) bool {
		return allCandles[i].OpenTime < allCandles[j].OpenTime
	})

	return models.Dataset{Candles: allCandles}, nil
}

// Also need a simple function to get the most recent 50 candles
// Could improve efficiency of this by using worker pool later on
func RecentCandles(symbol, interval string, limit int) ([]models.CandleStick, error) {
	endTime := time.Now().UnixMilli()
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d&endTime=%d", symbol, interval, limit, endTime)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles: %v", err)
	}
	defer resp.Body.Close()

	var rawData [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	candles := make([]models.CandleStick, 0, len(rawData))
	for _, item := range rawData {
		openTime := int64(item[0].(float64))
		open, _ := strconv.ParseFloat(item[1].(string), 64)
		high, _ := strconv.ParseFloat(item[2].(string), 64)
		low, _ := strconv.ParseFloat(item[3].(string), 64)
		closePrice, _ := strconv.ParseFloat(item[4].(string), 64)
		volume, _ := strconv.ParseFloat(item[5].(string), 64)
		closeTime := int64(item[6].(float64))
		numTrades := int64(item[8].(float64))

		candles = append(candles, models.CandleStick{
			OpenTime:    openTime,
			Open:        open,
			High:        high,
			Low:         low,
			Close:       closePrice,
			Volume:      volume,
			CloseTime:   closeTime,
			NumOfTrades: numTrades,
			IsFinal:     true,
		})
	}

	return candles, nil
}
