package screener

import (
	"database/sql"
	"errors"
	"sort"
)

type EOD struct {
	Date  string
	Close float64
}

// CalculateSMA computes SMA over given N periods
func CalculateSMA(data []EOD, days int) (float64, error) {
	if len(data) < days {
		return 0, errors.New("not enough data for SMA")
	}

	// Sort newest -> oldest
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date > data[j].Date
	})

	sum := 0.0
	for i := range days {
		sum += data[i].Close
	}
	return sum / float64(days), nil
}

func SaveSMA(db *sql.DB, ticker, date string, close, sma50, sma200 float64) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO prices (ticker, date, close, sma50, sma200)
		VALUES (?, ?, ?, ?, ?)`,
		ticker, date, close, sma50, sma200,
	)
	return err
}

func CalculateROE(netIncome, equity float64) (float64, error) {
	if equity == 0 {
		return 0, errors.New("equity cannot be zero")
	}
	return netIncome / equity, nil
}

func SaveROE(db *sql.DB, ticker string, roe, pe float64, outlook string) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO fundamentals (ticker, roe, pe_ratio, earnings_outlook, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'))`,
		ticker, roe, pe, outlook,
	)
	return err
}
