# HTTP Screener API

This package provides HTTP endpoints for the stock screener functionality using the EODHD API.

## Endpoints

### GET /api/screener

Returns paginated stock screener results based on specified filters and sorting criteria.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (1-based) |
| `limit` | integer | 50 | Number of results per page (1-1000) |
| `filters` | string | `[["market_capitalization",">",1000000]]` | Filter criteria (JSON array format) |
| `sort` | string | `market_capitalization.desc` | Sort criteria |

#### Filter Examples

1. 🇩🇪 Large Cap German Stocks on XETRA
```go
eodhd.ScreenerFilter{
	Filters: `[["exchange","=","XETRA"],["market_capitalization",">",20000000000]]`,
	Sort:    "market_capitalization.desc",
	Limit:   20,
	Offset:  0,
}
```
2. 🇺🇸 Tech Stocks in US with Dividend Yield
```go
eodhd.ScreenerFilter{
	Filters: `[["exchange","=","US"],["sector","=","Technology"],["dividend_yield",">",0.01]]`,
	Sort:    "dividend_yield.desc",
	Limit:   25,
	Offset:  0,
}
3. 🇪🇺 European Energy Companies with Positive EPS
```go
eodhd.ScreenerFilter{
	Filters: `[["sector","=","Energy"],["earnings_share",">",0],["exchange","=","F"]]`,
	Sort:    "earnings_share.desc",
	Limit:   15,
	Offset:  0,
}
```

4. 📉 Low Cap Stocks with High 5-Day Returns
```go
eodhd.ScreenerFilter{
	Filters: `[["market_capitalization","<",100000000],["refund_5d_p",">",10]]`,
	Sort:    "refund_5d_p.desc",
	Limit:   10,
	Offset:  0,
}
```
5. 🟢 High Volume ETFs in Europe
```go
eodhd.ScreenerFilter{
	Filters: `[["exchange","=","F"],["type","=","etf"],["avgvol_200d",">",50000]]`,
	Sort:    "avgvol_200d.desc",
	Limit:   10,
	Offset:  0,
}
```
6. 🏦 Financial Sector Stocks in Germany
```go
eodhd.ScreenerFilter{
	Filters: `[["exchange","=","XETRA"],["sector","=","Financial Services"]]`,
	Sort:    "market_capitalization.desc",
	Limit:   10,
	Offset:  0,
}
```
7. 🔎 Stocks with Positive Book Value Signal
```go
eodhd.ScreenerFilter{
	Filters: `[["exchange","=","US"]]`,
	Sort:    "market_capitalization.desc",
	Limit:   20,
	Offset:  0,
}
```

Add signals=bookvalue_pos as a URL parameter in future if needed — that can be an extra field in ScreenerFilter.

#### Sort Examples

- `market_capitalization.desc` - Sort by market cap descending
- `dividend_yield.asc` - Sort by dividend yield ascending
- `earnings_share.desc` - Sort by earnings per share descending
- `avgvol_200d.desc` - Sort by 200-day average volume descending
- `refund_5d_p.desc` - Sort by 5-day return percentage descending

#### Response Format

```json
{
  "data": [
    {
      "code": "AAPL",
      "name": "Apple Inc.",
      "exchange": "NASDAQ",
      "market_capitalization": 3000000000000,
      "dividend_yield": 0.005,
      "earnings_share": 6.05,
      "sector": "Technology",
      "industry": "Consumer Electronics",
      "adjusted_close": 150.25
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

# Get second page with 10 results per page
curl "http://localhost:8080/api/screener?page=2&limit=10"

# Filter by XETRA exchange with market cap > 10B (URL encoded)
curl "http://localhost:8080/api/screener?filters=%5B%5B%22exchange%22%2C%22%3D%22%2C%22XETRA%22%5D%2C%5B%22market_capitalization%22%2C%22%3E%22%2C10000000000%5D%5D"

# Filter by US Technology sector with dividend yield > 1% (URL encoded)
curl "http://localhost:8080/api/screener?filters=%5B%5B%22exchange%22%2C%22%3D%22%2C%22US%22%5D%2C%5B%22sector%22%2C%22%3D%22%2C%22Technology%22%5D%2C%5B%22dividend_yield%22%2C%22%3E%22%2C0.01%5D%5D"

# Filter by Energy sector with positive earnings (URL encoded)
curl "http://localhost:8080/api/screener?filters=%5B%5B%22sector%22%2C%22%3D%22%2C%22Energy%22%5D%2C%5B%22earnings_share%22%2C%22%3E%22%2C0%5D%5D"

# Filter ETFs with high trading volume (URL encoded)
curl "http://localhost:8080/api/screener?filters=%5B%5B%22type%22%2C%22%3D%22%2C%22etf%22%5D%2C%5B%22avgvol_200d%22%2C%22%3E%22%2C50000%5D%5D"

# Simple examples (unencoded for readability - encode before use)
# XETRA high cap: [["exchange","=","XETRA"],["market_capitalization",">",10000000000]]
# US Tech dividend: [["exchange","=","US"],["sector","=","Technology"],["dividend_yield",">",0.01]]
# Energy earnings: [["sector","=","Energy"],["earnings_share",">",0]]
# Small cap returns: [["market_capitalization","<",100000000],["refund_5d_p",">",10]]
# ETF high volume: [["type","=","etf"],["avgvol_200d",">",50000]]

# Sort by dividend yield ascending
curl "http://localhost:8080/api/screener?sort=dividend_yield.asc"
```

#### HTTP Status Codes

- `200 OK` - Success
- `400 Bad Request` - Invalid parameters
- `405 Method Not Allowed` - Non-GET request
- `500 Internal Server Error` - Server or API error

## FilterBuilder Usage

The package includes a `FilterBuilder` helper for constructing filter queries programmatically:

```go
package main

import (
    "fmt"
    "github.com/finsights-ai/backend/packages/http"
)

func main() {
    // Build a filter for US Technology stocks with dividend yield > 1%
    builder := http.NewFilterBuilder()
    filterStr, err := builder.
        Exchange("US").
        Sector("Technology").
        DividendYieldGreaterThan(0.01).
        Build()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(filterStr)
    // Output: [["exchange","=","US"],["sector","=","Technology"],["dividend_yield",">",0.01]]
}
```

**Available FilterBuilder Methods:**
- `Exchange(exchange string)` - Filter by exchange
- `MarketCapGreaterThan(value float64)` - Market cap greater than
- `MarketCapLessThan(value float64)` - Market cap less than
- `Sector(sector string)` - Filter by sector
- `DividendYieldGreaterThan(value float64)` - Dividend yield greater than
- `EarningsShareGreaterThan(value float64)` - Earnings per share greater than
- `Type(assetType string)` - Filter by asset type
- `AvgVolume200DGreaterThan(value float64)` - 200-day average volume greater than
- `Return5DGreaterThan(value float64)` - 5-day return percentage greater than
- `AddFilter(field, operator string, value any)` - Add custom filter

## Usage

```go
package main

import (
    "github.com/finsights-ai/backend/packages/eodhd"
    "github.com/finsights-ai/backend/packages/http"
)

func main() {
    client, err := eodhd.NewClient("your-api-key", "./cache")
    if err != nil {
        log.Fatal(err)
    }

    handler := http.NewScreenerHandler(client)
    http.HandleFunc("/api/screener", handler.GetScreenerData)

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Testing

Run the tests with:

```bash
go test ./packages/http/...
```

The test suite includes:
- Parameter validation
- Pagination logic
- Error handling
- Method restrictions
- Response format validation
