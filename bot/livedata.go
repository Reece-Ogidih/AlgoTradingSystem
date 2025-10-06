// Switched to using the DEX Screener API since it aggregates real-time on-chain swaps
// across multiple Solana DEXs, giving more accurate price data than Binance, which can
// diverge from actual on-chain liquidity conditions.

// My original implementation with a Binance Websocket connection can be found at the bottom of this file

package bot // Currently placing this within the overall trading bot file and package

import (
	"context" // This is to contrul runtime/cancellations since using Websocket instead of polling (Binance implementation)
	"database/sql"
	"encoding/json" // To unmashall
	"fmt"           // Standard format lib
	"log"           // For logging errors
	"net/http"      // To make the HTTP Requests to DEX Screener
	"os"
	"strconv" // For when I need to convert the strings to different types (Binance implementation)
	"strings" // For manipulation of strings (specifically make sure symbol is lower case) (Binance implementation)
	"time"

	models "github.com/Reece-Ogidih/CT-Bot/Models"
	"github.com/coder/websocket"       // Websocket library I decided to use (the ping/pong handling should be automatic) (Binance implementation)
	_ "github.com/go-sql-driver/mysql" // Need to save the data to the MySQL DB
	"github.com/joho/godotenv"         // Need to load secret info (Binance implementation)
)

// Will first define a helper function to help with pooling
// Will use a weighted average to aggregate price and other metrics
func AggregateDexData(
	pairs []models.DexPair) (models.AggregateSnapshot, error) {
	var totalWeighted float64 // Initially is 0.0
	var totalLiquidity float64
	var totalVol, totalBuys, totalSells float64

	for _, p := range pairs {
		price, err1 := p.PriceUSD.Float64()
		liq, err2 := p.Liquidity.USD.Float64()
		volM5, err3 := p.Volume.M5.Float64()

		// Skip broken entries
		// We still want to collect tx counts if possible
		if err1 == nil && err2 == nil && liq > 0 {
			totalWeighted += price * liq
			totalLiquidity += liq
		}

		if err3 == nil {
			totalVol += volM5
		}
		totalWeighted += price * liq
		totalLiquidity += liq

		totalVol += volM5
		totalBuys += float64(p.Txns.M5.Buys)
		totalSells += float64(p.Txns.M5.Sells)

		// Want to add a guard agains div by 0
		if totalLiquidity == 0 {
			return models.AggregateSnapshot{},
				fmt.Errorf("no valid liquidity")
		}
	}
	price := totalWeighted / totalLiquidity
	return models.AggregateSnapshot{
		PriceUSD:   price,
		VolumeM5:   totalVol,
		TxnsBuyM5:  int64(totalBuys),
		TxnsSellM5: int64(totalSells),
	}, nil
}

func FetchSOLUSDT() (<-chan models.CandleStick, error) {

	// Make the channel
	candleChan := make(chan models.CandleStick)

	go func() {
		// Will defer close the channel
		defer close(candleChan)

		// DEX Screener imposes a rate limit of 5 requests per sec
		// To ensure to complpy with DEX Screener's rate limits, will use a time system to fire off the requests
		ticker := time.NewTicker(210 * time.Millisecond) // I added a 10ms buffer to keep bot from hitting exact edge of rate limit
		defer ticker.Stop()

		// Will declare the url only for solana token for now, this function is easily adapted to make it token non-specific
		solToken := "So11111111111111111111111111111111111111112"
		url := fmt.Sprintf("https://api.dexscreener.com/latest/dex/tokens/%s", solToken)

		// Need a variable to track if this is the first run (Outdated)
		//firstRun := true

		// Need a variable to track the current candle
		var currCandle *models.CandleStick

		// Also now need to declare variables for calculations
		var currVol float64
		var currBuy int64
		var currSell int64

		// Will just set an infinite loop which starts a new loop in compoliance with our set rate
		for range ticker.C {
			// First, make the http request using DEX Screener API
			resp, err := http.Get(url)
			if err != nil {
				log.Println("Fetch error,", err)
				continue
			}

			defer resp.Body.Close()

			// body, _ := io.ReadAll(resp.Body)
			// log.Println(string(body))
			// resp.Body = io.NopCloser(bytes.NewBuffer(body))

			var parsed models.DexResponse
			if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
				log.Println("Parse error:,", err)
				continue
			}

			// Aggregate the pools here
			snap, err := AggregateDexData(parsed.Pairs)
			if err != nil {
				log.Println("Aggregate Error:", err)
				continue
			}
			// Since there is no need for multiple workers, dont need to worry about preserving chronological order
			// Need to parse the info and extract the necessary information here
			// Important to Note API return 5min averages
			// Therefore will calculate delta values and use those
			// Changed to averaging to per min values
			// This is due to the nature of the 5min averages returned
			price := snap.PriceUSD
			currVol = snap.VolumeM5
			buys := snap.TxnsBuyM5 // Already declared int64
			sells := snap.TxnsSellM5

			currBuy = buys
			currSell = sells

			// This is some outdated logic kept for interest
			// Was used with delta approach
			// if firstRun {
			// 	prevVol = currVol
			// 	prevBuy = currBuy
			// 	prevSell = currSell
			// 	firstRun = false
			// 	continue // skip first candle
			// }

			// deltaVol := currVol - prevVol
			// deltaBuy := currBuy - prevBuy
			// deltaSell := currSell - prevSell
			// deltaTrades := int64(deltaBuy + deltaSell)

			// DEX Screener reports rolling 5-minute metrics
			// Normalize to per-minute values for stable candles
			deltaVol := currVol / 5
			deltaBuy := currBuy / 5
			deltaSell := currSell / 5

			deltaTrades := int64(deltaBuy + deltaSell)

			now := time.Now().Unix()

			// Now check and start new candle if needed
			if currCandle == nil || now >= currCandle.OpenTime+60 {
				if currCandle != nil {
					currCandle.IsFinal = true
					candleChan <- *currCandle // Push finished candle
				}
				currCandle = &models.CandleStick{
					OpenTime:    now - (now % 60), // start of this minute
					Open:        price,
					High:        price,
					Low:         price,
					Close:       price,
					Volume:      deltaVol,
					NumOfTrades: deltaTrades,
				}
			}

			// To update current candle use Update method
			// Is defined in methods.go
			currCandle.Update(price, deltaVol, deltaTrades)

			// The following is outdated logic kept for interest
			// At the end of each loop, will need to set the data
			// This ensures that can continue calculating deltas
			// prevVol = currVol
			// prevBuy = currBuy
			// prevSell = currSell
		}
	}()
	return candleChan, nil
}

// BINANCE WEB-SOCKET IMPLEMENTATION

// For the function's input and output declarations, I am passing a ctx var as input to allow the caller to timeout (ctx short for context)
// The output is going to be a channel of candles which will be showing he candle data in real time

// Whilst I am first aiming to only build the bot to work in Sol and with a time interval of 1min, it is good practice to make this scalable
// This is especially true since I plan to expand out to multiple coins later on, so I added these as input paramaters

// Before creating the function to setup the live candle stream, will create a helper function
// This function is so that as the candle data is sent through the channel, it is stored in MySQL database for future checks/calculations

// First a function to load the .env info
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Format the string used to connect to the database here
func getDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
}

// Because of how I plan to call the function to insert to DB it is inefficient to open connection every time
// Instead create a function to do this at the start, can close at the end
var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("mysql", getDSN())
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
}

// Now the function to send the candles to the database
func candleToDB(candle models.CandleStick) {
	// Reuse shared db connection instead of opening every time so dont need to initiate connection
	// Also do not need to use db.Prepare here since only doing one entry
	_, err := db.Exec(`
        INSERT INTO live_candles_1m (open_times_ms, close, is_final)
        VALUES (?, ?, ?)`,
		candle.OpenTime, candle.Close, candle.IsFinal)

	if err != nil {
		log.Println("DB insert failed for candle:", candle.OpenTime, "err:", err)
	}
}

func FetchLive(ctx context.Context, symbol string, interval string) (<-chan models.CandleStick, error) {
	// First need to put in the Endpoint for Binance Websocket incorporating the input variables
	Address := fmt.Sprintf("wss://fstream.binance.com/ws/%s_perpetual@continuousKline_%s", strings.ToLower(symbol), interval)

	// Use the Dial function from websocket package to initiate a websocket connection
	conn, _, err := websocket.Dial(ctx, Address, nil) // We can ignore the HTTP response hence the _
	if err != nil {
		return nil, err
	}

	// Initialise the Database to store the candles
	InitDB()

	// Create the channel for candles data
	candleChan := make(chan models.CandleStick)

	// Start a background goroutine to read messages, using a goroutine otherwise everything would be blocked by the infinite loop
	go func() {
		// To ensure the connection and the channel close after, we defer them
		defer conn.Close(websocket.StatusNormalClosure, "Closing the connection")
		defer close(candleChan)

		// Initiate an infinite loop
		for {
			_, message, err := conn.Read(ctx)
			if err != nil {
				log.Println("WebSocket read error:", err)
				return // Stop on error
			}

			// Need to unmarshall the JSON message
			var msg models.BinanceKlineWrapper
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println("Unmarshal error:", err)
				continue
			}

			// Convert the strings to float64 and construct the Candle with type Candlestick
			open, _ := strconv.ParseFloat(msg.Kline.Open, 64)
			close, _ := strconv.ParseFloat(msg.Kline.Close, 64)
			high, _ := strconv.ParseFloat(msg.Kline.High, 64)
			low, _ := strconv.ParseFloat(msg.Kline.Low, 64)
			volume, _ := strconv.ParseFloat(msg.Kline.Volume, 64)

			candle := models.CandleStick{
				OpenTime:    msg.Kline.OpenTime,
				Open:        open,
				High:        high,
				Low:         low,
				Close:       close,
				Volume:      volume,
				CloseTime:   msg.Kline.CloseTime,
				NumOfTrades: msg.Kline.NumOfTrades,
				IsFinal:     msg.Kline.IsFinal,
			}

			// Send through the channel
			candleChan <- candle
			if candle.IsFinal {
				go candleToDB(candle)
			}
		}
	}()
	return candleChan, nil
}
