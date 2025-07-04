package bot // Currently placing this within the overall trading bot file and package

import (
	"context"       // This is to contrul runtime/cancellations since using Websocket instead of polling
	"encoding/json" // To unmashall
	"fmt"           // Standard format lib
	"log"           // For logging errors
	"strconv"       // For when I need to convert the strings to different types
	"strings"       // For manipulation of strings (specifically make sure symbol is lower case)

	models "github.com/Reece-Ogidih/CT-Bot/Models"
	"github.com/coder/websocket" // Websocket library I decided to use (the ping/pong handling should be automatic)
)

// For the function's input and output declarations, I am passing a ctx var as input to allow the caller to timeout (ctx short for context)
// The output is going to be a channel of candles which will be showing he candle data in real time

// Whilst I am first aiming to only build the bot to work in Sol and with a time interval of 1min, it is good practice to make this scalable
// This is especially true since I plan to expand out to multiple coins later on, so I added these as input paramaters

func FetchLive(ctx context.Context, symbol string, interval string) (<-chan models.CandleStick, error) {

	// First need to put in the Endpoint for Binance Websocket incorporating the input variables
	Address := fmt.Sprintf("wss://fstream.binance.com/ws/%s_perpetual@continuousKline_%s", strings.ToLower(symbol), interval)

	// Use the Dial function from websocket package to initiate a websocket connection
	conn, _, err := websocket.Dial(ctx, Address, nil) // We can ignore the HTTP response hence the _
	if err != nil {
		return nil, err
	}

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
			}

			// Send through the channel
			candleChan <- candle
		}
	}()
	return candleChan, nil
}
