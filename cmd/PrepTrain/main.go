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
	// For now will skip and directly insert the enriched candles into the DB to check indexing and any errors so far.

	// Insert the data into the DB
	stmt, err := db.Prepare(`
	INSERT INTO train_ml (open_times_ms, open, close, high, low, volume, adx)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal("Preparation error:", err)
	}
	defer stmt.Close()

	for _, c := range fullCandles {
		_, err := stmt.Exec(c.OpenTime, c.Open, c.Close, c.High, c.Low, c.Volume, c.ADX)
		if err != nil {
			log.Println("Error inserting:", err)
		}
	}

	// Conclusive print (would expect 525600-14 since there are that many minutes in a year (first 13 have no adx))
	fmt.Printf("Total number of Candles	inserted: %d\n", len(fullCandles))
}
