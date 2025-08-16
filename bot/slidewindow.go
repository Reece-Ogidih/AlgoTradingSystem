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
	SupLine     models.Trendline
	ResLine     models.Trendline
	Initialised bool
	// Will need to add the indices list that I am using in the simulation bot
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

// Also add a version for the bot run to prepare for ML training
type SlidingWindowTrain struct {
	Symbol      string // Will add this for scalability later when expanding to multiple coins
	Interval    string
	Size        int
	Candles     []models.EnrichedCandle
	SupLine     models.Trendline
	ResLine     models.Trendline
	Initialised bool
	Idxs        []int // Will need to add an index to each candle in order for trendline slopes to not be near zero
}

func (sw *SlidingWindowTrain) NewWindowTrain(candle models.EnrichedCandle) {
	sw.Candles = append(sw.Candles, candle)

	// To keep the window at a fixed size, remove the oldest candle when a new one arrives
	if len(sw.Candles) > sw.Size {
		sw.Candles = sw.Candles[1:]
	}
}

// Function to increase the index list when a new candle is considered. will separate this into two functions since I may want the index for next candle without updating window
func (sw *SlidingWindowTrain) GetNextIdx() (nextIdx int) {
	nextIdx = sw.Idxs[len(sw.Idxs)-1] + 1
	return nextIdx
}

func (sw *SlidingWindowTrain) UpdateIdx(nextIdx int) {
	sw.Idxs = append(sw.Idxs, nextIdx)

	if len(sw.Idxs) > sw.Size {
		sw.Idxs = sw.Idxs[1:]
	}
}
