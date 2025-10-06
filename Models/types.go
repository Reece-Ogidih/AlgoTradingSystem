package models

import "encoding/json"

// This file is for storing all type declarations which are not local to a single file
// For a different file to use one of these types it will need to preface it with models. (For example models.Candlestick)

// First I define type of candlestick which then has the data I pull from each candlestick
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
	IsFinal     bool
}

type Dataset struct {
	Candles []CandleStick
}

// Next will need to define the type for the response we get from DEX Screener HTTP request
// I use json.Number for safe parsing of both strings and floats, since it could change
// type DexResponse struct {
// 	ChainID  string   `json:"chainId"` // These are struct tags that help the encoding/json package and help map json fieldnames to Go
// 	DexID    string   `json:"dexId"`   // Example: Orca
// 	URL      string   `json:"url"`     // Example "https://dexscreener.com/solana/..."
// 	PairAddr string   `json:"pairAddress"`
// 	Labels   []string `json:"labels"`

// 	BaseToken struct {
// 		Address string `json:"address"` // Example: Solana's address "So111...2"
// 		Name    string `json:"name"`    // Example: Solana
// 		Symbol  string `json:"symbol"`  // Example SOL
// 	} `json:"baseToken"`

// 	QuoteToken struct { // Example for this field: USDT's information
// 		Address string `json:"address"`
// 		Name    string `json:"name"`
// 		Symbol  string `json:"symbol"`
// 	} `json:"quoteToken"`

// 	PriceNative string      `json:"priceNative"`
// 	PriceUSD    json.Number `json:"priceUsd"`

// 	Volume struct {
// 		M5  json.Number `json:"m5"`  // This is volume over last 5min
// 		H1  json.Number `json:"h1"`  // This is volume over last hour
// 		H6  json.Number `json:"h6"`  // This is volume over last 6hrs
// 		H24 json.Number `json:"h24"` // This is volume over last 1day
// 	} `json:"volume"`

// 	Txns struct { // Number of transactions over last x time period (transaction as in buy or sell)
// 		M5 struct {
// 			Buys  int64 `json:"buys"`
// 			Sells int64 `json:"sells"`
// 		} `json:"m5"`
// 		H1 struct {
// 			Buys  int64 `json:"buys"`
// 			Sells int64 `json:"sells"`
// 		} `json:"h1"`
// 		H6 struct {
// 			Buys  int64 `json:"buys"`
// 			Sells int64 `json:"sells"`
// 		} `json:"h6"`
// 		H24 struct {
// 			Buys  int64 `json:"buys"`
// 			Sells int64 `json:"sells"`
// 		} `json:"h24"`
// 	} `json:"txns"`

// 	PriceChange struct {
// 		M5  json.Number `json:"m5"`
// 		H1  json.Number `json:"h1"`
// 		H6  json.Number `json:"h6"`
// 		H24 json.Number `json:"h24"`
// 	} `json:"priceChange"`

// 	Liquidity struct {
// 		USD   json.Number `json:"usd"`
// 		Base  json.Number `json:"base"`
// 		Quote json.Number `json:"quote"`
// 	} `json:"liquidity"`

// 	PairCreatedAt int64 `json:"pairCreatedAt"`
// }

type Token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type Volume struct {
	M5  json.Number `json:"m5"`
	H1  json.Number `json:"h1"`
	H6  json.Number `json:"h6"`
	H24 json.Number `json:"h24"`
}

type TxnCounts struct {
	Buys  int64 `json:"buys"`
	Sells int64 `json:"sells"`
}

type Txns struct {
	M5  TxnCounts `json:"m5"`
	H1  TxnCounts `json:"h1"`
	H6  TxnCounts `json:"h6"`
	H24 TxnCounts `json:"h24"`
}

type Liquidity struct {
	USD   json.Number `json:"usd"`
	Base  json.Number `json:"base"`
	Quote json.Number `json:"quote"`
}

type DexPair struct {
	ChainID       string      `json:"chainId"`
	DexID         string      `json:"dexId"`
	URL           string      `json:"url"`
	PairAddr      string      `json:"pairAddress"`
	Labels        []string    `json:"labels"`
	BaseToken     Token       `json:"baseToken"`
	QuoteToken    Token       `json:"quoteToken"`
	PriceNative   string      `json:"priceNative"`
	PriceUSD      json.Number `json:"priceUsd"`
	Volume        Volume      `json:"volume"`
	Txns          Txns        `json:"txns"`
	PriceChange   Volume      `json:"priceChange"` // reusing Volume shape for convenience
	Liquidity     Liquidity   `json:"liquidity"`
	PairCreatedAt int64       `json:"pairCreatedAt"`
}

type DexResponse struct {
	SchemaVersion string    `json:"schemaVersion"`
	Pairs         []DexPair `json:"pairs"`
}

// Will need a type for our aggregated data snapshots
type AggregateSnapshot struct {
	PriceUSD   float64
	VolumeM5   float64
	TxnsBuyM5  int64
	TxnsSellM5 int64
}

// When working with the WebSocket, the format is a little different so need to define two other types
// First we define a wrapper type so that we only need to unmarshall part of the output
type BinanceKlineWrapper struct {
	EventType    string       `json:"e"`
	EventTime    int64        `json:"E"`
	PairSymbol   string       `json:"s"`
	ContractType string       `json:"ps"`
	Kline        BinanceKline `json:"k"`
}

// Now can declare the Kline with all the corresponding Candle data
type BinanceKline struct {
	OpenTime      int64  `json:"t"`
	CloseTime     int64  `json:"T"`
	Symbol        string `json:"S"`
	Interval      string `json:"i"`
	FirstTradeID  int64  `json:"F"`
	LastTradeID   int64  `json:"L"`
	Open          string `json:"o"`
	Close         string `json:"c"`
	High          string `json:"h"`
	Low           string `json:"l"`
	Volume        string `json:"v"`
	NumOfTrades   int64  `json:"n"`
	IsFinal       bool   `json:"x"`
	QuoteVolume   string `json:"q"`
	TakerBuyVol   string `json:"V"`
	TakerBuyQuote string `json:"Q"`
	Ignore        string `json:"B"`
}

// Need the type which includes our standard candlestick data but also the ADX value
type EnrichedCandle struct {
	OpenTime int64
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
	ADX      float64
	PlusDI   float64
	MinusDI  float64
}

// Need the type for the data which will be inserted to MySQL DB to develop the ML component
type DevData struct {
	OpenTime int64
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
	ADX      float64
	Idx      int
	SigEntry int
	SigExit  int
}

// Need the type for our trendlines
type Trendline struct {
	Gradient  float64 // Need the gradient and intercept for linear line: y = mx + c
	Intercept float64
	A1Time    int64 // Also need the Open time of the 2 anchor candles
	A2Time    int64
	A1Price   float64 // Also need the 2 anchor candles respective close price
	A2Price   float64
}
