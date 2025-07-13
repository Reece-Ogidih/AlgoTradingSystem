package bot // Currently placing this within the overall trading bot file and package

import (
	"context" // This is to contrul runtime/cancellations since using Websocket instead of polling
	"database/sql"
	"encoding/json" // To unmashall
	"fmt"           // Standard format lib
	"log"           // For logging errors
	"os"
	"strconv" // For when I need to convert the strings to different types
	"strings" // For manipulation of strings (specifically make sure symbol is lower case)

	models "github.com/Reece-Ogidih/CT-Bot/Models"
	"github.com/coder/websocket"       // Websocket library I decided to use (the ping/pong handling should be automatic)
	_ "github.com/go-sql-driver/mysql" // Need to save the data to the MySQL DB
	"github.com/joho/godotenv"         // Need to load secret info
)

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
