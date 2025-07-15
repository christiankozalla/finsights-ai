package db

import (
	"database/sql"
	"log"
)

func InsertSampleData(db *sql.DB) error {
	log.Println("Checking for sample data...")

	// Check if data already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM fundamentals").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Found %d records in fundamentals table, skipping sample data insertion", count)
		return nil
	}

	log.Println("Inserting sample data...")

	// Sample fundamentals data
	fundamentalsData := `
		INSERT OR REPLACE INTO fundamentals
		(ticker, pe_ratio, roe, earnings_outlook, dividend_yield, dividend_growth_5y, intrinsic_value, margin_of_safety)
		VALUES
		('AAPL', 14.5, 0.25, 'positive', 0.005, 0.08, 180.50, 0.25),
		('GOOGL', 13.1, 0.18, 'positive', 0.0, 0.0, 3100.0, 0.15),
		('MSFT', 12.5, 0.22, 'positive', 0.035, 0.12, 375.0, 0.22),
		('TSLA', 45.2, 0.15, 'neutral', 0.0, 0.0, 800.0, -0.05),
		('IBM', 8.3, 0.08, 'negative', 0.045, 0.08, 120.0, 0.35),
		('KO', 9.7, 0.16, 'positive', 0.045, 0.08, 65.0, 0.25),
		('JNJ', 11.2, 0.18, 'positive', 0.038, 0.06, 170.0, 0.18),
		('PFE', 7.8, 0.12, 'positive', 0.055, 0.10, 55.0, 0.30),
		('WMT', 26.5, 0.19, 'stable', 0.016, 0.04, 145.0, 0.05),
		('XOM', 13.8, 0.14, 'neutral', 0.058, 0.03, 95.0, 0.12),
		('JPM', 10.2, 0.16, 'positive', 0.025, 0.05, 155.0, 0.18),
		('DIS', 22.1, 0.08, 'neutral', 0.0, 0.0, 110.0, 0.08),
		('NVDA', 65.3, 0.35, 'positive', 0.003, 0.15, 420.0, -0.12),
		('AMZN', 48.7, 0.12, 'positive', 0.0, 0.0, 3200.0, 0.02),
		('META', 18.9, 0.24, 'positive', 0.0, 0.0, 285.0, 0.15);
	`

	if _, err := db.Exec(fundamentalsData); err != nil {
		return err
	}

	// Sample prices data
	pricesData := `
		INSERT OR REPLACE INTO prices
		(ticker, date, close, sma50, sma200)
		VALUES
		('AAPL', '2024-01-15', 150.25, 145.80, 140.30),
		('GOOGL', '2024-01-15', 2750.80, 2720.50, 2680.20),
		('MSFT', '2024-01-15', 330.59, 325.20, 315.80),
		('TSLA', '2024-01-15', 220.45, 235.60, 245.90),
		('IBM', '2024-01-15', 78.20, 82.40, 85.10),
		('KO', '2024-01-15', 48.75, 52.20, 55.50),
		('JNJ', '2024-01-15', 158.30, 162.10, 165.80),
		('PFE', '2024-01-15', 42.15, 45.20, 48.90),
		('WMT', '2024-01-15', 162.85, 158.40, 155.20),
		('XOM', '2024-01-15', 104.25, 98.70, 95.30),
		('JPM', '2024-01-15', 168.90, 165.20, 160.50),
		('DIS', '2024-01-15', 98.75, 102.30, 105.80),
		('NVDA', '2024-01-15', 875.28, 820.50, 750.20),
		('AMZN', '2024-01-15', 3087.50, 3120.80, 3200.40),
		('META', '2024-01-15', 378.42, 365.20, 350.10);
	`

	if _, err := db.Exec(pricesData); err != nil {
		return err
	}

	log.Println("Sample data inserted successfully")
	return nil
}
