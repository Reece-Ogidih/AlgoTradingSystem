package models

// This file is for storing all type declarations which are not local to a single file

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
