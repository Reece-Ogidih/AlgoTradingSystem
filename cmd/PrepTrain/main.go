package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

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
	// main.go getCandles()
	rows, err := db.Query(`
	SELECT open_times_ms, open, high, low, close, volume
	FROM hist_candles_1m
	ORDER BY open_times_ms ASC
	`)

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
	// First we get the set of ADX values
	size, err := strconv.Atoi(os.Getenv("ADX_PERIOD"))
	if err != nil {
		log.Fatalf("Error parsing ADX_PERIOD: %v", err)
	}
	ADXs, plusDIvals, minusDIvals, _, _, _, _, _ := bot.CalculateADX(candles, size)
	var enriched []models.EnrichedCandle

	// Will first trim the slices to remove the first N candles (no ADX values)
	candles = candles[size:]
	ADXs = ADXs[size:]
	plusDIvals = plusDIvals[size:]
	minusDIvals = minusDIvals[size:]

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
			PlusDI:   plusDIvals[i],
			MinusDI:  minusDIvals[i],
		})
	}

	return enriched
}

func main() {
	// First pull some of the key info from .env
	adx_threshold, err := strconv.Atoi(os.Getenv("ADX_THRESHOLD")) // This is the required ADX for a trade to be placed
	if err != nil {
		log.Fatalf("Error parsing ADX_THRESHOLD: %v", err) // Note: These can be converted from strings to integers and then to float64 since there is no decimals
	}
	adx_min, err := strconv.Atoi(os.Getenv("ADX_MIN")) // This is the required ADX for a position to be held
	if err != nil {
		log.Fatalf("Error parsing ADX_MIN: %v", err)
	}

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
	// Will immediately place the first N candles into the window
	size, err := strconv.Atoi(os.Getenv("WINDOW_SIZE")) // Grab the window size from .env
	if err != nil {
		log.Fatalf("Error parsing WINDOW_SIZE: %v", err)
	}
	window := bot.SlidingWindowTrain{
		Symbol:      "SOLUSDT",
		Interval:    "1m",
		Size:        size,
		Candles:     fullCandles[:size],
		SupLine:     models.Trendline{},
		ResLine:     models.Trendline{},
		Initialised: true,
		Idxs:        make([]int, size),
	}

	// For loop to fill in the Indices
	for i := range window.Idxs {
		window.Idxs[i] = i
	}

	// Now need to remove these first N candles from the remaining dataset
	fullCandles = fullCandles[size:]

	// Now will compute the support and resistance trendlines
	window.ComputeTrendlines()

	// The logic that will be implemented is that the bot will wait for a breakout or for certain number of candles to pass.
	// In the case of breakout, ADX will be evaluated, if the trend strength is high enough it will trigger a "trade"
	// No matter if a trade is triggered or not, once there is a breakout or n candles pass, New trendlines will be drawn.
	// Will first set a counter for candles since the last update
	candlesSinceUpdate := 0

	// Will also add some lines to note if bot is in a trade as well as entry and exit positions
	entrySignal := 0
	exitSignal := 0
	currentPos := 0
	inTrade := currentPos != 0

	// Need the next var for some debugging prints within the loop
	countprint := 0

	// To ensure the bot will recalibrate the trendlines after N candles I declare 2 variables
	// One variable is the max number of candles when the bot is not in a trade before it recalibrates
	// The other is specifically how long it can hold trades
	idleRecalibrate, err := strconv.Atoi(os.Getenv("IDLE_LIMIT")) // This would be frequency of redrawing the trendlines when outside of trades
	if err != nil {
		log.Fatalf("Error parsing IDLE_LIMIT: %v", err)
	}
	activeRecalibrate, err := strconv.Atoi(os.Getenv("ACTIVE_LIMIT")) // This means the bot can hold a trade for at most this long
	if err != nil {
		log.Fatalf("Error parsing ACTIVE_LIMIT: %v", err)
	}
	// Finally, the slice of candle data for Ml dev needs to be declared
	var finalData []models.DevData

	// Now can begin the loop
	for i := 0; i < len(fullCandles); i++ {
		newCandle := fullCandles[i] // The "current" candle as it would be in live stream
		nextIdx := window.GetNextIdx()
		adx := newCandle.ADX
		plusDI := newCandle.PlusDI
		minusDI := newCandle.MinusDI

		// Must reset entry and exit signals for each candle
		entrySignal = 0
		exitSignal = 0

		// Now check for breakouts, will need to declare this variable
		var breakoutOccured bool
		var brokeSup bool
		var brokeRes bool

		// Now need to flag if there was a breakout
		// Must account for if we are in a trade
		if currentPos == 1 {
			brokeRes = false
			brokeSup = window.CheckTrendlines2(window.SupLine, newCandle, false, currentPos, nextIdx)
			breakoutOccured = brokeRes || brokeSup // || is the "or" operator in Go, so this is saying either broke res line or broke sup line
		} else if currentPos == -1 {
			brokeRes = window.CheckTrendlines2(window.ResLine, newCandle, true, currentPos, nextIdx)
			brokeSup = false
			breakoutOccured = brokeRes || brokeSup
		} else {
			brokeRes = window.CheckTrendlines2(window.ResLine, newCandle, true, currentPos, nextIdx)
			brokeSup = window.CheckTrendlines2(window.SupLine, newCandle, false, currentPos, nextIdx)
			breakoutOccured = brokeRes || brokeSup
		}

		// This is the debugging part added to print

		supY := window.SupLine.Gradient*float64(nextIdx) + window.SupLine.Intercept
		resY := window.ResLine.Gradient*float64(nextIdx) + window.ResLine.Intercept

		if countprint <= 100 {
			countprint++
			fmt.Printf("Pos=%d x=%d SupGrad=%.5f SupInt=%.5f SupY=%.5f ResGrad=%.5f ResInt=%.5f ResY=%.5f High=%.5f Low=%.5f\n",
				currentPos, nextIdx,
				window.SupLine.Gradient, window.SupLine.Intercept, supY,
				window.ResLine.Gradient, window.ResLine.Intercept, resY,
				newCandle.High, newCandle.Low)
		}

		// Add a checker for if we are in a trade and the adx drops (so little momentum in market), the bot should exit its position (anticipating reversal)
		if inTrade && adx < float64(adx_min) {
			if currentPos == 1 {
				exitSignal = 1
				currentPos = 0
				inTrade = false
			} else if currentPos == -1 {
				exitSignal = -1
				currentPos = 0
				inTrade = false
			} else {
				fmt.Println("Error, intrade and currentPos don't align, candle:", nextIdx)
			}
		}

		// Can now do conditional actions based on breakouot detection
		if breakoutOccured {
			// First need to check to see if there was an ongoing trade as the bot would exit here
			if brokeRes && currentPos == -1 {
				exitSignal = -1 // Would exit here
				currentPos = 0  // Reset our position to no trades
				inTrade = false
			}
			if brokeSup && currentPos == 1 {
				exitSignal = 1
				currentPos = 0
				inTrade = false
			}

			if adx >= float64(adx_threshold) {
				// This would mean that there is a strong trend, bot should attempt to trigger trade here
				// Note that the currentPos = 0 check would not be in the actual trading bot as it limits trades to only occuring when not in any positions
				if brokeRes && plusDI > minusDI {
					// Would enter into a long position here, so will update trackers accordingly
					// Must check if this is a new entry or if theres an error
					if currentPos == 0 {
						entrySignal = 1
						currentPos = 1
						inTrade = true
					} else if currentPos == -1 {
						exitSignal = -1
						entrySignal = 1
						currentPos = 1
						inTrade = true
					} else {
						fmt.Println("Error, proposing long whilst in long position:", newCandle)
					}
				} else if brokeSup && plusDI < minusDI {
					// Would enter into a short position
					// Must check if this is a new entry or if theres an error
					if currentPos == 0 {
						entrySignal = -1
						currentPos = -1
						inTrade = true
					} else if currentPos == 1 {
						exitSignal = 1
						entrySignal = -1
						currentPos = -1
						inTrade = true
					} else {
						fmt.Println("Error, proposing short whilst in short position", newCandle)
					}
				}
			}

			// After a breakout, trendlines will be redrawn regardless of trend strength
			candlesSinceUpdate = 0
			window.NewWindowTrain(newCandle)
			window.UpdateIdx(nextIdx)
			window.ComputeTrendlines()

		} else {
			// After checking for breakout, increase the counter and then check to see if there have been too many candles that have passed
			candlesSinceUpdate++
			window.NewWindowTrain(newCandle) // Update the sliding window
			window.UpdateIdx(nextIdx)

			// Check for too long idle or holding a position
			maxCandles := idleRecalibrate
			if inTrade {
				maxCandles = activeRecalibrate
			}
			if candlesSinceUpdate >= maxCandles {
				// Exit any trades
				if currentPos != 0 {
					exitSignal = currentPos
					currentPos = 0
					inTrade = false
				}

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
			Idx:      nextIdx,
			SigEntry: entrySignal,
			SigExit:  exitSignal,
		})
	}

	// Insert the data into the DB
	stmt, err := db.Prepare(`
	INSERT INTO train_ml (open_times_ms, open, close, high, low, volume, adx, idx, sig_entry, sig_exit)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal("Preparation error:", err)
	}
	defer stmt.Close()

	for _, c := range finalData {
		_, err := stmt.Exec(c.OpenTime, c.Open, c.Close, c.High, c.Low, c.Volume, c.ADX, c.Idx, c.SigEntry, c.SigExit)
		if err != nil {
			log.Println("Error inserting:", err)
		}
	}

	// Conclusive print (would expect 525600-14 since there are that many minutes in a year (first 14 have no adx))
	fmt.Printf("Total number of Candles	inserted: %d\n", len(fullCandles))
}
