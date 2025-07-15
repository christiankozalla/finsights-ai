-- Create base tables
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

-- Indexes
CREATE INDEX IF NOT EXISTS idx_fundamentals_pe_ratio ON fundamentals(pe_ratio);
CREATE INDEX IF NOT EXISTS idx_fundamentals_roe ON fundamentals(roe);
CREATE INDEX IF NOT EXISTS idx_fundamentals_dividend_yield ON fundamentals(dividend_yield);
CREATE INDEX IF NOT EXISTS idx_fundamentals_margin_of_safety ON fundamentals(margin_of_safety);
CREATE INDEX IF NOT EXISTS idx_fundamentals_earnings_outlook ON fundamentals(earnings_outlook);
CREATE INDEX IF NOT EXISTS idx_prices_ticker_date ON prices(ticker, date);
CREATE INDEX IF NOT EXISTS idx_prices_close ON prices(close);
