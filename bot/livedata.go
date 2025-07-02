package bot			// Currently placing this within the overall trading bot file and package

import {
	"fmt"
	"nhooyr.io/websocket"		// Websocket library I decided to use
	"context"					// 	This is to contrul runtime/cancellations since using Websocket instead of polling
	"github.com/Reece-Ogidih/CT-Bot/Models"
}

// First need to put in the Endpoint for Binance Websocket
// "<pair>_<contractType>@continuousKline_<interval>" I need to check this as I believe this is the stream I want
const Address = "wss://ws-fapi.binance.com/ws-fapi/v1"


// For the function's input and output declarations, I am passing a ctx var as input to allow the caller to timeout
// The output is going to be a channel of candles which will be showing he candle data in real time

// Whilst I am first aiming to only build the bot to work in Sol and with a time interval of 1min, it is good practice to make this scalable
// This is especially true since I plan to expand out to multiple coins later on, so I added these as input paramaters

func FetchLive(ctx context.Context, symbol string, interval string) (<-chan Candle, error) {

}