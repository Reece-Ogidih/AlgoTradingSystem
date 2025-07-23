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
func getCandles() ([]models.CandleStick, error) {
	// Open the connection to the DB
	db, err := sql.Open("mysql", getDSN())
	if err != nil {
		log.Fatal("DB connection error:", err)
	}

	// Will defer the close until end of function
	defer db.Close()

	// Want each of the rows of the database sicne each row represents one candle
	rows, err := db.Query("SELECT open_times_ms, open, close, high, low, volume FROM hist_candles_1m ORDER BY open_times_ms ASC")
	if err != nil {
		return nil, fmt.Errorf("Query Error: %w", err)
	}

	defer rows.Close()

	// Now can declare the slice of candles, and each individual candle within it
	var candles []models.CandleStick
	var candle models.CandleStick

	// Can now do a loop "rows.Next()" will return true if there is a next row
	for rows.Next() {
		if err := rows.Scan(&candle.OpenTime, &candle.Open, &candle.High, &candle.Low, &candle.Close, &candle.Volume); err != nil {
			return nil, fmt.Errorf("Scan Error: %w", err)
		}

		// If the scan works (meaning all the fields are filled) we can append this to candle set
		candles = append(candles, candle)
	}

	// Final Error check
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Iteration error: %w", err)
	}

	return candles, nil
}

// The next Helper function we need is one that will calculate the ADX on all of these data points and then remove the first N candles (without ADX value)
func getADXCandles(candles []models.CandleStick) (fullCandles []models.EnrichedCandle) {
	// First we get the set of ADX values (am using standard period of 14 but may change this if too sensitive)
	ADXs, _, _, _, _, _ := bot.CalculateADX(candles, 14)
	var enriched []models.EnrichedCandle

	// Will first append them without adjusting for first 14 having no ADX value, then can check the output data to see exactly how they align
	for i := 0; i < len(candles); i++ {
		enriched = append(enriched, models.EnrichedCandle{
			Candle: candles[i],
			ADX:    ADXs[i],
		})
	}

	return enriched
}

func main() {

}
