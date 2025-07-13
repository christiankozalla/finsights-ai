# HTTP Screener API

This package provides HTTP endpoints for the stock screener functionality using a custom SQLite-based screener instead of the EODHD screener API.

## Overview

The screener API has been refactored to use a custom database-backed implementation that provides more flexibility and control over stock screening criteria. The API maintains backward compatibility with the JSON array filter format while adding support for new financial metrics.

## Endpoints

### GET /api/screener

Returns paginated stock screening results based on specified filters and sorting criteria.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (1-based) |
| `limit` | integer | 50 | Number of results per page (1-1000) |
| `filters` | string | `[["pe_ratio","<",20]]` | Filter criteria (JSON array format) |
| `sort` | string | `pe_ratio.asc` | Sort criteria |

#### Available Fields for Filtering

**Fundamental Metrics:**
- `pe_ratio` - Price-to-earnings ratio
- `roe` - Return on equity
- `dividend_yield` - Dividend yield percentage
- `dividend_growth_5y` - 5-year dividend growth rate
- `intrinsic_value` - Calculated intrinsic value
- `margin_of_safety` - Margin of safety percentage
- `earnings_outlook` - Earnings outlook (positive, negative, neutral, stable)
- `ticker` - Stock ticker symbol

**Price Metrics:**
- `close` - Current closing price
- `sma50` - 50-day simple moving average
- `sma200` - 200-day simple moving average

**Special Computed Fields:**
- `price_vs_sma50` - Price relative to SMA50 (use `< 1.0` for below SMA)
- `price_vs_sma200` - Price relative to SMA200 (use `< 1.0` for below SMA)
- `intrinsic_vs_price` - Intrinsic value relative to price (use `> 1.0` for undervalued)

#### Filter Examples

**Basic Filters:**
```json
[["pe_ratio","<",15]]
[["roe",">",0.15]]
[["dividend_yield",">",0.03]]
[["margin_of_safety",">",0.20]]
```

**Complex Filters:**
```json
[["pe_ratio","<",15],["roe",">",0.15]]
[["dividend_yield",">",0.03],["dividend_growth_5y",">",0.05]]
[["price_vs_sma200","<",1.0],["pe_ratio","<",12]]
```

**Earnings Outlook Filter:**
```json
[["earnings_outlook","=","positive"]]
```

#### Sort Options

- `pe_ratio.asc` / `pe_ratio.desc` - Sort by P/E ratio
- `roe.asc` / `roe.desc` - Sort by return on equity
- `close.asc` / `close.desc` - Sort by closing price
- `dividend_yield.asc` / `dividend_yield.desc` - Sort by dividend yield
- `margin_of_safety.asc` / `margin_of_safety.desc` - Sort by margin of safety
- `ticker.asc` / `ticker.desc` - Sort by ticker symbol

#### Response Format

```json
{
  "data": [
    {
      "ticker": "AAPL",
      "pe_ratio": 14.5,
      "roe": 0.25,
      "close": 150.25,
      "sma50": 145.80,
      "sma200": 140.30,
      "earnings_outlook": "positive",
      "dividend_yield": 0.005,
      "dividend_growth_5y": 0.08,
      "intrinsic_value": 180.50,
      "margin_of_safety": 0.25
    }
  ],
  "page": 1,
  "limit": 50,
  "total_count": 1,
  "has_more": false
}
```

#### Error Response

```json
{
  "error": "ERROR_CODE",
  "message": "Human readable error message"
}
```

#### Example Requests

```bash
# Get first page with default settings
curl "http://localhost:8080/api/screener"

# Value stocks: Low P/E and high ROE
curl "http://localhost:8080/api/screener?filters=%5B%5B%22pe_ratio%22%2C%22%3C%22%2C15%5D%2C%5B%22roe%22%2C%22%3E%22%2C0.15%5D%5D"

# Dividend stocks: Good yield and growth
curl "http://localhost:8080/api/screener?filters=%5B%5B%22dividend_yield%22%2C%22%3E%22%2C0.03%5D%2C%5B%22dividend_growth_5y%22%2C%22%3E%22%2C0.05%5D%5D"

# Undervalued stocks: High margin of safety
curl "http://localhost:8080/api/screener?filters=%5B%5B%22margin_of_safety%22%2C%22%3E%22%2C0.20%5D%5D"

# Bargain stocks: Low P/E and below 200-day moving average
curl "http://localhost:8080/api/screener?filters=%5B%5B%22pe_ratio%22%2C%22%3C%22%2C10%5D%2C%5B%22price_vs_sma200%22%2C%22%3C%22%2C1%5D%5D"

# Growth stocks: High ROE and positive outlook
curl "http://localhost:8080/api/screener?filters=%5B%5B%22roe%22%2C%22%3E%22%2C0.20%5D%2C%5B%22earnings_outlook%22%2C%22%3D%22%2C%22positive%22%5D%5D"

# Sort by margin of safety descending
curl "http://localhost:8080/api/screener?sort=margin_of_safety.desc"
```

#### Unencoded Filter Examples (URL encode before use)
```
Value stocks: [["pe_ratio","<",15],["roe",">",0.15]]
Dividend stocks: [["dividend_yield",">",0.03],["dividend_growth_5y",">",0.05]]
Undervalued stocks: [["margin_of_safety",">",0.20]]
Growth stocks: [["roe",">",0.20],["earnings_outlook","=","positive"]]
Bargain stocks: [["pe_ratio","<",10],["price_vs_sma200","<",1.0]]
```

#### HTTP Status Codes

- `200 OK` - Success
- `400 Bad Request` - Invalid parameters or filter format
- `405 Method Not Allowed` - Non-GET request
- `500 Internal Server Error` - Database or server error

## FilterBuilder Usage (Go)

The package includes a `FilterBuilder` helper for constructing filter queries programmatically:

```go
package main

import (
    "fmt"
    "github.com/finsights-ai/backend/packages/http"
)

func main() {
    // Build a filter for value stocks
    builder := http.NewFilterBuilder()
    filterStr, err := builder.
        PELessThan(15).
        ROEGreaterThan(0.15).
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(filterStr)
    // Output: [["pe_ratio","<",15],["roe",">",0.15]]
}
```

**Available FilterBuilder Methods:**
- `PELessThan(value float64)` - P/E ratio less than value
- `PEGreaterThan(value float64)` - P/E ratio greater than value
- `ROEGreaterThan(value float64)` - ROE greater than value
- `ROELessThan(value float64)` - ROE less than value
- `DividendYieldGreaterThan(value float64)` - Dividend yield greater than
- `MarginOfSafetyGreaterThan(value float64)` - Margin of safety greater than
- `PriceBelowSMA50()` - Price below 50-day moving average
- `PriceBelowSMA200()` - Price below 200-day moving average
- `EarningsOutlook(outlook string)` - Filter by earnings outlook
- `Ticker(ticker string)` - Filter by specific ticker
- `AddFilter(field, operator string, value interface{})` - Add custom filter

## Screener Package Integration

The HTTP handler integrates with the `screener` package which provides:

### Database Schema

```sql
CREATE TABLE fundamentals (
  ticker TEXT PRIMARY KEY,
  pe_ratio REAL,
  roe REAL,
  earnings_outlook TEXT,
  dividend_yield REAL,
  dividend_growth_5y REAL,
  intrinsic_value REAL,
  margin_of_safety REAL
);

CREATE TABLE prices (
  ticker TEXT,
  date TEXT,
  close REAL,
  sma50 REAL,
  sma200 REAL,
  PRIMARY KEY (ticker, date)
);
```

### Screener FilterBuilder (Go)

```go
import "github.com/finsights-ai/backend/packages/screener"

// Create complex filters using the screener package
builder := screener.NewFilterBuilder()
filter := builder.
    PELessThan(15).
    ROEGreaterThan(0.15).
    DividendYieldGreaterThan(0.02).
    MarginOfSafetyGreaterThan(0.20).
    BuildWithPagination("pe_ratio.asc", 25, 0)

results, err := screener.ScreenStocks(db, filter)
```

### Preset Filters

The screener package includes predefined filter presets:

- **ValueStocks** - Low P/E, high ROE stocks
- **DividendStocks** - High dividend yield and growth
- **UndervaluedStocks** - High margin of safety
- **GrowthStocks** - High ROE with positive outlook
- **BargainStocks** - Low P/E stocks below 200-day MA

## Setup and Usage

```go
package main

import (
    "database/sql"
    "github.com/finsights-ai/backend/packages/http"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    // Open database
    db, err := sql.Open("sqlite3", "./screener.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create screener client
    screenerClient := http.NewDatabaseScreenerClient(db)
    
    // Create handler
    handler := http.NewScreenerHandler(screenerClient)
    
    // Setup routes
    http.HandleFunc("/api/screener", handler.GetScreenerData)
    
    // Start server
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Environment Variables

- `DATABASE_PATH` - Path to SQLite database file (default: "./screener.db")

## Testing

Run the tests with:

```bash
go test ./packages/http/...
go test ./packages/screener/...
```

The test suite includes:
- Parameter validation
- Pagination logic
- Filter parsing and validation
- SQL query generation
- Error handling
- Method restrictions
- Response format validation
- Database integration tests

## Migration from EODHD Screener

The new implementation maintains API compatibility while replacing the external EODHD screener with a custom database solution:

**Key Changes:**
- Filters now operate on local database fields
- New financial metrics available (intrinsic value, margin of safety)
- Better performance with local data
- More flexible filtering capabilities
- Custom computed fields for technical analysis

**Backward Compatibility:**
- Same JSON array filter format
- Same pagination parameters
- Same response structure
- Same HTTP endpoints