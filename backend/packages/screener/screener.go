package screener

import (
	"database/sql"
)

type Filter struct {
	MaxPE         float64
	MinROE        float64
	PriceBelowSMA string // "sma50" | "sma200"
}

type Stock struct {
	Ticker          string
	PE              float64
	ROE             float64
	Close           float64
	SMA50           float64
	SMA200          float64
	EarningsOutlook string
}

func Screen(db *sql.DB, filter Filter) ([]Stock, error) {
	query := `
		SELECT f.ticker, f.pe_ratio, f.roe, p.close, p.sma50, p.sma200, f.earnings_outlook
		FROM fundamentals f
		JOIN prices p ON f.ticker = p.ticker
		WHERE f.pe_ratio <= ?
		  AND f.roe >= ?
		  AND p.date = (SELECT MAX(date) FROM prices WHERE ticker = f.ticker)
	`
	args := []any{filter.MaxPE, filter.MinROE}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Stock
	for rows.Next() {
		var s Stock
		if err := rows.Scan(&s.Ticker, &s.PE, &s.ROE, &s.Close, &s.SMA50, &s.SMA200, &s.EarningsOutlook); err != nil {
			return nil, err
		}

		if filter.PriceBelowSMA == "sma50" && s.Close >= s.SMA50 {
			continue
		}
		if filter.PriceBelowSMA == "sma200" && s.Close >= s.SMA200 {
			continue
		}

		results = append(results, s)
	}
	return results, nil
}
