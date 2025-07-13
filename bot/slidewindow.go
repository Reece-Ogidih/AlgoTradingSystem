package bot

import (
	"log"

	histdata "github.com/Reece-Ogidih/CT-Bot/HistoricalData"
	models "github.com/Reece-Ogidih/CT-Bot/Models"
)

// Need the type for the SlidingWindow so that everything to do with the slidingwindow can be encapsulated
// Did this here since methods can only be defined on local types
type SlidingWindow struct {
	Symbol      string // Will add this for scalability later when expanding to multiple coins
	Interval    string
	Size        int
	Candles     []models.CandleStick
	Initialised bool
}

func (sw *SlidingWindow) Init() error {
	// Add the case when the window has already been initialised
	if sw.Initialised {
		return nil
	}

	// Now make the call to fetch recent candles to fill the initial window
	candles, err := histdata.RecentCandles(sw.Symbol, sw.Interval, sw.Size)
	if err != nil {
		log.Fatal(err)
	}

	// Adjust the states of the struct
	sw.Candles = candles
	sw.Initialised = true
	return nil
}

// Next we define the method to update the sliding window live as new candles come in
func (sw *SlidingWindow) NewWindow(candle models.CandleStick) {

	sw.Candles = append(sw.Candles, candle)

	// To keep the window at a fixed size, remove the oldest candle when a new one arrives
	if len(sw.Candles) > sw.Size {
		sw.Candles = sw.Candles[1:]
	}
}
