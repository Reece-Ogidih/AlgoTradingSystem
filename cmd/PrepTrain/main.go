package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	models "github.com/Reece-Ogidih/CT-Bot/Models"
	bot "github.com/Reece-Ogidih/CT-Bot/bot"
	_ "github.com/go-sql-driver/mysql" // Need to save the data to the MySQL DB & Pull from Hist DB
	"github.com/joho/godotenv"         // Need to load secret info
)

// So this will be an original running version of the bot
// The point of this Bot is not to actually make trades, but to instead signal when trades wouldve been entered and allow for diagnostics to be done
// We will slightly adjust the logic to allow for the bot to feed in the historical data, as it wouldve with the live stream
// Trading logic will be used equivalently to plan for main bot, so trendline detection + ADX
// The bot will take in this historical data and will also attach a variable for buy or sell orders (ie 1 - buy, 0 - no action, -1 - sell)
// The original hist data + these flags for when trades were entered or exited will be entered into a new table within MySQL DB
// Can then use this dataset to develop the ML model.

// Load the .env info
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

// Will need a helper function to get the data from the MySQL database
func getCandles(db *sql.DB) ([]models.CandleStick, error) {
	// Since this function is getting run in main, will not need to reopen the db, but will instead pass it as an argument.
	// Want each of the rows of the database sicne each row represents one candle
	rows, err := db.Query("SELECT open_times_ms, open, close, high, low, volume FROM hist_candles_1m ORDER BY open_times_ms ASC")
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	defer rows.Close()

	// Now can declare the slice of candles, and each individual candle within it
	var candles []models.CandleStick
	var candle models.CandleStick

	// Can now do a loop "rows.Next()" will return true if there is a next row
	for rows.Next() {
		if err := rows.Scan(&candle.OpenTime, &candle.Open, &candle.High, &candle.Low, &candle.Close, &candle.Volume); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// If the scan works (meaning all the fields are filled) we can append this to candle set
		candles = append(candles, candle)
	}

	// Final Error check
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error: %w", err)
	}

	return candles, nil
}

// The next Helper function we need is one that will calculate the ADX on all of these data points and then remove the first N candles (without ADX value)
func getADXCandles(candles []models.CandleStick) (fullCandles []models.EnrichedCandle) {
	// First we get the set of ADX values (am using standard period of 14 but may change this if too sensitive)
	ADXs, _, _, _, _, _ := bot.CalculateADX(candles, 14)
	var enriched []models.EnrichedCandle

	// Will first trim the slices to remove the first 14 candles (no ADX values)
	candles = candles[14:]
	ADXs = ADXs[14:]

	// Can now combine the original candle data with the ADX values to get the enriched candle
	for i := 0; i < len(candles); i++ {
		enriched = append(enriched, models.EnrichedCandle{
			OpenTime: candles[i].OpenTime,
			Open:     candles[i].Open,
			High:     candles[i].High,
			Low:      candles[i].Low,
			Close:    candles[i].Close,
			Volume:   candles[i].Volume,
			ADX:      ADXs[i],
		})
	}

	return enriched
}

func main() {
	// Connect to the DB
	db, err := sql.Open("mysql", getDSN())
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	defer db.Close()

	// Get the candles
	candles, err := getCandles(db)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Transform them into the form that holds the ADX as well
	fullCandles := getADXCandles(candles)

	// The next step here is to now create the sliding window and then create the detection for when to lodge BUY vs Sell orders
	// Will immediately place the first 40 candles into the window
	window := bot.SlidingWindowTrain{
		Symbol:      "SOLUSDT",
		Interval:    "1m",
		Size:        40,
		Candles:     fullCandles[:40],
		SupLine:     models.Trendline{},
		ResLine:     models.Trendline{},
		Initialised: true,
	}

	// Now need to remove these first 40 candles from the remaining dataset
	fullCandles = fullCandles[40:]

	// Now will compute the support and resistance trendlines
	window.ComputeTrendlines()

	// The logic that will be implemented is that the bot will wait for a breakout or for certain number of candles to pass.
	// In the case of breakout, ADX will be evaluated, if the trend strength is high enough it will trigger a "trade"
	// No matter if a trade is triggered or not, once there is a breakout or n candles pass, New trendlines will be drawn.
	// Will first set a counter for candles since the last update
	candlesSinceUpdate := 0

	// Will also add a line to note if bot is in a trade
	inTrade := false
	signal := 0

	// To ensure the bot will recalibrate the trendlines after N candles I declare 2 variables
	// One variable is the max number of candles when the bot is not in a trade before it recalibrates
	// The other is specifically how long it can hold trades
	idleRecalibrate := 60    // This would be redraw the trendlines once an hour
	activeRecalibrate := 120 // This means the bot can hold a trade for at most 2 hours

	// Finally, the slice of candle data for Ml dev needs to be declared
	var finalData []models.DevData

	// Now can begin the loop
	for i := 0; i < len(fullCandles); i++ {
		newCandle := fullCandles[i] // The "current" candle as it would be in live stream

		// Now check for breakouts
		brokeRes := window.CheckTrenlines2(window.ResLine, newCandle, true)
		brokeSup := window.CheckTrenlines2(window.SupLine, newCandle, false)

		// Now need to flag if there was a breakout
		breakoutOccured := brokeRes || brokeSup // || is the "or" operator in Go, so this is saying either broke res line or broke sup line

		// Can now do conditional actions based on breakouot detection
		if breakoutOccured {
			// Will start with thethreshold for ADX being 30, found 25 to be quite weak.
			adx := newCandle.ADX
			if adx >= 30.0 {
				// This would mean that there is a strong trend, bot should attempt to trigger trade here
				inTrade = true
				if brokeRes {
					signal = 1
				} else if brokeSup {
					signal = -1
				}
			}

			// After a breakout, trendlines will be redrawn regardless of trend strength
			candlesSinceUpdate = 0
			window.NewWindowTrain(newCandle)
			window.ComputeTrendlines()
		} else {
			// After checking for breakout, increase the counter and then check to see if there have been too many candles that have passed
			candlesSinceUpdate++
			window.NewWindowTrain(newCandle) // Update the sliding window

			// Check for too long idle or holding a position
			maxCandles := idleRecalibrate
			if inTrade {
				maxCandles = activeRecalibrate
			}
			if candlesSinceUpdate >= maxCandles {
				// Exit any trades
				inTrade = false
				signal = 0

				// Reset the counter and recalibrate the trendlines
				candlesSinceUpdate = 0
				window.ComputeTrendlines()
			}
		}
		finalData = append(finalData, models.DevData{
			OpenTime: newCandle.OpenTime,
			Open:     newCandle.Open,
			High:     newCandle.High,
			Low:      newCandle.Low,
			Close:    newCandle.Close,
			Volume:   newCandle.Volume,
			ADX:      newCandle.ADX,
			Signal:   signal,
		})
	}

	// Insert the data into the DB
	stmt, err := db.Prepare(`
	INSERT INTO train_ml (open_times_ms, open, close, high, low, volume, adx, trades)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal("Preparation error:", err)
	}
	defer stmt.Close()

	for _, c := range finalData {
		_, err := stmt.Exec(c.OpenTime, c.Open, c.Close, c.High, c.Low, c.Volume, c.ADX, c.Signal)
		if err != nil {
			log.Println("Error inserting:", err)
		}
	}

	// Conclusive print (would expect 525600-14 since there are that many minutes in a year (first 13 have no adx))
	fmt.Printf("Total number of Candles	inserted: %d\n", len(fullCandles))
}
