CREATE TABLE IF NOT EXISTS fundamentals (
  ticker TEXT PRIMARY KEY,
  pe_ratio REAL,
  roe REAL,
  yoy_profit JSON,
  yoy_turnover JSON,
  earnings_outlook TEXT,
  updated_at TEXT
);

CREATE TABLE IF NOT EXISTS prices (
  ticker TEXT,
  date TEXT,
  close REAL,
  sma50 REAL,
  sma200 REAL,
  PRIMARY KEY (ticker, date)
);

ALTER TABLE fundamentals ADD COLUMN dividend_yield REAL;
ALTER TABLE fundamentals ADD COLUMN dividend_growth_5y REAL;
ALTER TABLE fundamentals ADD COLUMN intrinsic_value REAL;
ALTER TABLE fundamentals ADD COLUMN margin_of_safety REAL;
