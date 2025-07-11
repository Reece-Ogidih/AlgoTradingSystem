package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	histdata "github.com/Reece-Ogidih/CT-Bot/HistoricalData"
	_ "github.com/go-sql-driver/mysql" // Need to save the data to the MySQL DB
	"github.com/joho/godotenv"         // Need to load secret info
)

// Load the .env info
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func getDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
}

// Time-wise we start with 1 year of data since it is short term day trading bot so should be sufficient, if model doesnt seem robust can extend to 2 years.

func main() {
	from := time.Date(2024, 7, 11, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 11, 0, 0, 0, 0, time.UTC)

	// Connect to the DB
	db, err := sql.Open("mysql", getDSN())
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	defer db.Close()

	// Now fetch the candle data
	data, err := histdata.FetchCandles(from, to)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Insert the data into the DB
	stmt, err := db.Prepare(`
	INSERT INTO hist_candles_1m (open_times_ms, open, close, high, low, volume, is_final)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal("Preparation error:", err)
	}
	defer stmt.Close()

	for _, c := range data.Candles {
		_, err := stmt.Exec(c.OpenTime, c.Open, c.Close, c.High, c.Low, c.Volume, c.IsFinal)
		if err != nil {
			log.Println("Error inserting:", err)
		}
	}

	// Conclusive print (would expect 525600 since there are that many minutes in a year)
	fmt.Printf("Total number of Candles	inserted: %d\n", len(data.Candles))
}
