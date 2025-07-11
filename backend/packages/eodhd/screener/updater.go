package screener

import (
	"database/sql"
	"log"
	"time"

	"github.com/finsights-ai/backend/packages/eodhd"
)

func ShouldUpdateNow(now time.Time) bool {
	weekday := now.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

func RunNightlyUpdate(db *sql.DB, client *eodhd.Client, tickers []string) {
	now := time.Now()
	weekday := now.Weekday()

	if weekday == time.Saturday || weekday == time.Sunday {
		log.Println("Skipping nightly update: weekend.")
		return
	}

	log.Println("Starting nightly update...")

	for _, ticker := range tickers {
		log.Printf("Updating: %s\n", ticker)

		err := ProcessTicker(db, client, ticker)
		if err != nil {
			log.Printf("Error updating %s: %v\n", ticker, err)
			continue
		}
	}

	log.Println("Nightly update complete.")
}
