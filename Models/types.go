package models

// This file is for storing all type declarations which are not local to a single file
// For a different file to use one of these types it will need to preface it with models. (For example models.Candlestick)

// First I define type of candlestick which then has the data I pull from each candlestick through Binance REST API
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

	// Will need to add the type in for the technical indicators which will be appended after theyre calculated from the raw data
	// For now leave out, may create a new type for this
	// EMA9       float64
	// EMA21      float64
	// StochRSI   float64
	// MACD       float64
	// MACDSignal float64
	// MACDHist   float64
}

type Dataset struct {
	Candles []CandleStick
}

// When working with the WebSocket, the format is a little different so need to define two other types
// First we define a wrapper type so that we only need to unmarshall part of the output
type BinanceKlineWrapper struct {
	EventType    string       `json:"e"` // These are struct tags that help the encoding/json package and help map json fieldnames to Go ones
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
