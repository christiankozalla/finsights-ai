package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var (
		dbPath     = flag.String("db", "./screener.db", "Path to SQLite database file")
		withSample = flag.Bool("sample", false, "Insert sample data for testing")
		force      = flag.Bool("force", false, "Force recreate database (drops existing data)")
	)
	flag.Parse()

	fmt.Printf("Initializing database at: %s\n", *dbPath)

	// Remove existing database if force flag is set
	if *force {
		if err := os.Remove(*dbPath); err != nil && !os.IsNotExist(err) {
			log.Fatal("Failed to remove existing database:", err)
		}
		fmt.Println("Removed existing database")
	}

	// Open database connection
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create schema
	if err := createSchema(db); err != nil {
		log.Fatal("Failed to create schema:", err)
	}
	fmt.Println("Created database schema")

	// Insert sample data if requested
	if *withSample {
		if err := insertSampleData(db); err != nil {
			log.Fatal("Failed to insert sample data:", err)
		}
		fmt.Println("Inserted sample data")
	}

	fmt.Println("Database initialization completed successfully!")
}

func createSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS fundamentals (
			ticker TEXT PRIMARY KEY,
			pe_ratio REAL,
			roe REAL,
			yoy_profit JSON,
			yoy_turnover JSON,
			earnings_outlook TEXT,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
			dividend_yield REAL,
			dividend_growth_5y REAL,
			intrinsic_value REAL,
			margin_of_safety REAL
		);

		CREATE TABLE IF NOT EXISTS prices (
			ticker TEXT,
			date TEXT,
			close REAL,
			sma50 REAL,
			sma200 REAL,
			PRIMARY KEY (ticker, date)
		);

		-- Create indexes for better query performance
		CREATE INDEX IF NOT EXISTS idx_fundamentals_pe_ratio ON fundamentals(pe_ratio);
		CREATE INDEX IF NOT EXISTS idx_fundamentals_roe ON fundamentals(roe);
		CREATE INDEX IF NOT EXISTS idx_fundamentals_dividend_yield ON fundamentals(dividend_yield);
		CREATE INDEX IF NOT EXISTS idx_fundamentals_margin_of_safety ON fundamentals(margin_of_safety);
		CREATE INDEX IF NOT EXISTS idx_fundamentals_earnings_outlook ON fundamentals(earnings_outlook);
		CREATE INDEX IF NOT EXISTS idx_prices_ticker_date ON prices(ticker, date);
		CREATE INDEX IF NOT EXISTS idx_prices_close ON prices(close);
	`

	_, err := db.Exec(schema)
	return err
}

func insertSampleData(db *sql.DB) error {
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
		return fmt.Errorf("failed to insert fundamentals data: %w", err)
	}

	// Sample prices data (latest date for each ticker)
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
		return fmt.Errorf("failed to insert prices data: %w", err)
	}

	// Insert some historical price data for a few tickers
	historicalData := `
		INSERT OR REPLACE INTO prices
		(ticker, date, close, sma50, sma200)
		VALUES
		('AAPL', '2024-01-14', 148.50, 145.20, 140.10),
		('AAPL', '2024-01-13', 147.75, 144.80, 139.90),
		('AAPL', '2024-01-12', 149.20, 144.50, 139.70),
		('GOOGL', '2024-01-14', 2730.20, 2715.30, 2675.80),
		('GOOGL', '2024-01-13', 2742.15, 2710.50, 2670.40),
		('GOOGL', '2024-01-12', 2755.80, 2705.20, 2665.20),
		('MSFT', '2024-01-14', 328.75, 324.80, 315.40),
		('MSFT', '2024-01-13', 332.20, 324.50, 315.10),
		('MSFT', '2024-01-12', 329.85, 324.20, 314.80);
	`

	if _, err := db.Exec(historicalData); err != nil {
		return fmt.Errorf("failed to insert historical data: %w", err)
	}

	return nil
}
