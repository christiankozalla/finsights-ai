package screener

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// ScreenerResult represents a stock result from screening
type ScreenerResult struct {
	Ticker           string  `json:"ticker"`
	PE               float64 `json:"pe_ratio"`
	ROE              float64 `json:"roe"`
	Close            float64 `json:"close"`
	SMA50            float64 `json:"sma50"`
	SMA200           float64 `json:"sma200"`
	EarningsOutlook  string  `json:"earnings_outlook"`
	DividendYield    float64 `json:"dividend_yield"`
	DividendGrowth5Y float64 `json:"dividend_growth_5y"`
	IntrinsicValue   float64 `json:"intrinsic_value"`
	MarginOfSafety   float64 `json:"margin_of_safety"`
}

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// ScreenerFilter contains filtering, sorting and pagination parameters
type ScreenerFilter struct {
	Conditions []FilterCondition `json:"conditions"`
	Sort       string            `json:"sort"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// FilterBuilder provides an idiomatic way to build filters
type FilterBuilder struct {
	conditions []FilterCondition
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		conditions: make([]FilterCondition, 0),
	}
}

// AddCondition adds a generic filter condition
func (fb *FilterBuilder) AddCondition(field, operator string, value any) *FilterBuilder {
	fb.conditions = append(fb.conditions, FilterCondition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return fb
}

// PE ratio filters
func (fb *FilterBuilder) PELessThan(value float64) *FilterBuilder {
	return fb.AddCondition("pe_ratio", "<", value)
}

func (fb *FilterBuilder) PEGreaterThan(value float64) *FilterBuilder {
	return fb.AddCondition("pe_ratio", ">", value)
}

func (fb *FilterBuilder) PEBetween(min, max float64) *FilterBuilder {
	return fb.AddCondition("pe_ratio", ">=", min).AddCondition("pe_ratio", "<=", max)
}

// ROE filters
func (fb *FilterBuilder) ROEGreaterThan(value float64) *FilterBuilder {
	return fb.AddCondition("roe", ">", value)
}

func (fb *FilterBuilder) ROELessThan(value float64) *FilterBuilder {
	return fb.AddCondition("roe", "<", value)
}

func (fb *FilterBuilder) ROEBetween(min, max float64) *FilterBuilder {
	return fb.AddCondition("roe", ">=", min).AddCondition("roe", "<=", max)
}

// Price filters
func (fb *FilterBuilder) PriceBelowSMA50() *FilterBuilder {
	return fb.AddCondition("price_vs_sma50", "<", 1.0)
}

func (fb *FilterBuilder) PriceBelowSMA200() *FilterBuilder {
	return fb.AddCondition("price_vs_sma200", "<", 1.0)
}

func (fb *FilterBuilder) PriceAboveSMA50() *FilterBuilder {
	return fb.AddCondition("price_vs_sma50", ">", 1.0)
}

func (fb *FilterBuilder) PriceAboveSMA200() *FilterBuilder {
	return fb.AddCondition("price_vs_sma200", ">", 1.0)
}

func (fb *FilterBuilder) PriceBetween(min, max float64) *FilterBuilder {
	return fb.AddCondition("close", ">=", min).AddCondition("close", "<=", max)
}

// Dividend filters
func (fb *FilterBuilder) DividendYieldGreaterThan(value float64) *FilterBuilder {
	return fb.AddCondition("dividend_yield", ">", value)
}

func (fb *FilterBuilder) DividendGrowthGreaterThan(value float64) *FilterBuilder {
	return fb.AddCondition("dividend_growth_5y", ">", value)
}

// Valuation filters
func (fb *FilterBuilder) MarginOfSafetyGreaterThan(value float64) *FilterBuilder {
	return fb.AddCondition("margin_of_safety", ">", value)
}

func (fb *FilterBuilder) IntrinsicValueGreaterThan(currentPrice float64) *FilterBuilder {
	return fb.AddCondition("intrinsic_vs_price", ">", 1.0)
}

// Earnings outlook filter
func (fb *FilterBuilder) EarningsOutlook(outlook string) *FilterBuilder {
	return fb.AddCondition("earnings_outlook", "=", outlook)
}

// Ticker filter (for specific stocks)
func (fb *FilterBuilder) Ticker(ticker string) *FilterBuilder {
	return fb.AddCondition("ticker", "=", ticker)
}

func (fb *FilterBuilder) TickerIn(tickers []string) *FilterBuilder {
	return fb.AddCondition("ticker", "IN", tickers)
}

// Build creates the final filter
func (fb *FilterBuilder) Build() ScreenerFilter {
	return ScreenerFilter{
		Conditions: fb.conditions,
		Sort:       "pe_ratio ASC", // Default sort
		Limit:      50,             // Default limit
		Offset:     0,              // Default offset
	}
}

// BuildWithPagination creates the final filter with custom sort and pagination
func (fb *FilterBuilder) BuildWithPagination(sort string, limit, offset int) ScreenerFilter {
	return ScreenerFilter{
		Conditions: fb.conditions,
		Sort:       sort,
		Limit:      limit,
		Offset:     offset,
	}
}

// ParseFilterFromJSON parses a JSON filter string (compatible with EODHD format)
func ParseFilterFromJSON(filterJSON string) (ScreenerFilter, error) {
	if filterJSON == "" {
		return ScreenerFilter{
			Conditions: []FilterCondition{},
			Sort:       "pe_ratio ASC",
			Limit:      50,
			Offset:     0,
		}, nil
	}

	var rawConditions [][]any
	if err := json.Unmarshal([]byte(filterJSON), &rawConditions); err != nil {
		return ScreenerFilter{}, fmt.Errorf("invalid filter JSON: %w", err)
	}

	conditions := make([]FilterCondition, 0, len(rawConditions))
	for _, raw := range rawConditions {
		if len(raw) != 3 {
			return ScreenerFilter{}, fmt.Errorf("invalid condition format: expected [field, operator, value]")
		}

		field, ok := raw[0].(string)
		if !ok {
			return ScreenerFilter{}, fmt.Errorf("field must be a string")
		}

		operator, ok := raw[1].(string)
		if !ok {
			return ScreenerFilter{}, fmt.Errorf("operator must be a string")
		}

		// Map EODHD-style field names to our schema
		field = mapFieldName(field)

		conditions = append(conditions, FilterCondition{
			Field:    field,
			Operator: operator,
			Value:    raw[2],
		})
	}

	return ScreenerFilter{
		Conditions: conditions,
		Sort:       "pe_ratio ASC",
		Limit:      50,
		Offset:     0,
	}, nil
}

// mapFieldName maps external field names to internal database column names
func mapFieldName(field string) string {
	mapping := map[string]string{
		"market_capitalization": "market_cap",
		"dividend_yield":        "dividend_yield",
		"earnings_share":        "eps",
		"sector":                "sector",
		"industry":              "industry",
		"exchange":              "exchange",
		"refund_5d_p":           "return_5d",
		"avgvol_200d":           "avg_volume_200d",
		"type":                  "asset_type",
	}

	if mapped, exists := mapping[field]; exists {
		return mapped
	}
	return field
}

// ScreenStocks performs stock screening based on the provided filter
func ScreenStocks(db *sql.DB, filter ScreenerFilter) ([]ScreenerResult, error) {
	query, args := buildQuery(filter)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var results []ScreenerResult
	for rows.Next() {
		var result ScreenerResult
		err := rows.Scan(
			&result.Ticker,
			&result.PE,
			&result.ROE,
			&result.Close,
			&result.SMA50,
			&result.SMA200,
			&result.EarningsOutlook,
			&result.DividendYield,
			&result.DividendGrowth5Y,
			&result.IntrinsicValue,
			&result.MarginOfSafety,
		)
		if err != nil {
			return nil, fmt.Errorf("row scanning failed: %w", err)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return results, nil
}

// buildQuery constructs the SQL query based on the filter conditions
func buildQuery(filter ScreenerFilter) (string, []any) {
	baseQuery := `
		SELECT
			f.ticker,
			COALESCE(f.pe_ratio, 0) as pe_ratio,
			COALESCE(f.roe, 0) as roe,
			COALESCE(p.close, 0) as close,
			COALESCE(p.sma50, 0) as sma50,
			COALESCE(p.sma200, 0) as sma200,
			COALESCE(f.earnings_outlook, '') as earnings_outlook,
			COALESCE(f.dividend_yield, 0) as dividend_yield,
			COALESCE(f.dividend_growth_5y, 0) as dividend_growth_5y,
			COALESCE(f.intrinsic_value, 0) as intrinsic_value,
			COALESCE(f.margin_of_safety, 0) as margin_of_safety
		FROM fundamentals f
		LEFT JOIN (
			SELECT ticker, close, sma50, sma200
			FROM prices p1
			WHERE date = (SELECT MAX(date) FROM prices p2 WHERE p2.ticker = p1.ticker)
		) p ON f.ticker = p.ticker
	`

	var whereConditions []string
	var args []any

	// Build WHERE clause from filter conditions
	for _, condition := range filter.Conditions {
		sqlCondition, value := buildSQLCondition(condition)
		if sqlCondition != "" {
			whereConditions = append(whereConditions, sqlCondition)
			if value != nil {
				// Handle array values for IN operator
				if arr, ok := value.([]string); ok {
					for _, v := range arr {
						args = append(args, v)
					}
				} else {
					args = append(args, value)
				}
			}
		}
	}

	// Add WHERE clause if there are conditions
	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add ORDER BY clause
	if filter.Sort != "" {
		baseQuery += " ORDER BY " + sanitizeSort(filter.Sort)
	}

	// Add LIMIT and OFFSET
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	return baseQuery, args
}

// buildSQLCondition converts a FilterCondition to SQL
func buildSQLCondition(condition FilterCondition) (string, any) {
	field := condition.Field
	operator := condition.Operator
	value := condition.Value

	// Handle special computed fields
	switch field {
	case "price_vs_sma50":
		if operator == "<" && value == 1.0 {
			return "p.close < p.sma50", nil
		} else if operator == ">" && value == 1.0 {
			return "p.close > p.sma50", nil
		}
	case "price_vs_sma200":
		if operator == "<" && value == 1.0 {
			return "p.close < p.sma200", nil
		} else if operator == ">" && value == 1.0 {
			return "p.close > p.sma200", nil
		}
	case "intrinsic_vs_price":
		if operator == ">" && value == 1.0 {
			return "f.intrinsic_value > p.close", nil
		}
	}

	// Handle IN operator for arrays
	if operator == "IN" {
		if arr, ok := value.([]string); ok {
			placeholders := strings.Repeat("?,", len(arr)-1) + "?"
			if isFieldInFundamentals(field) {
				return fmt.Sprintf("f.%s IN (%s)", field, placeholders), value
			}
			return fmt.Sprintf("%s IN (%s)", field, placeholders), value
		}
	}

	// Map field to table alias
	if isFieldInFundamentals(field) {
		field = "f." + field
	} else if isFieldInPrices(field) {
		field = "p." + field
	} else {
		// For unknown fields, assume fundamentals table
		field = "f." + field
	}

	// Standard operators
	switch operator {
	case "=", ">", "<", ">=", "<=", "!=":
		return fmt.Sprintf("%s %s ?", field, operator), value
	case "LIKE":
		return fmt.Sprintf("%s LIKE ?", field), value
	}

	return "", nil
}

// isFieldInFundamentals checks if a field belongs to the fundamentals table
func isFieldInFundamentals(field string) bool {
	fundamentalsFields := []string{
		"ticker", "pe_ratio", "roe", "earnings_outlook",
		"dividend_yield", "dividend_growth_5y", "intrinsic_value", "margin_of_safety",
	}
	return slices.Contains(fundamentalsFields, field)
}

// isFieldInPrices checks if a field belongs to the prices table
func isFieldInPrices(field string) bool {
	pricesFields := []string{"close", "sma50", "sma200"}
	return slices.Contains(pricesFields, field)
}

// sanitizeSort ensures the sort parameter is safe for SQL
func sanitizeSort(sort string) string {
	// Allow only known fields and directions
	validSorts := map[string]string{
		"pe_ratio.asc":          "f.pe_ratio ASC",
		"pe_ratio.desc":         "f.pe_ratio DESC",
		"roe.asc":               "f.roe ASC",
		"roe.desc":              "f.roe DESC",
		"close.asc":             "p.close ASC",
		"close.desc":            "p.close DESC",
		"dividend_yield.asc":    "f.dividend_yield ASC",
		"dividend_yield.desc":   "f.dividend_yield DESC",
		"margin_of_safety.asc":  "f.margin_of_safety ASC",
		"margin_of_safety.desc": "f.margin_of_safety DESC",
		"ticker.asc":            "f.ticker ASC",
		"ticker.desc":           "f.ticker DESC",
	}

	if sanitized, exists := validSorts[sort]; exists {
		return sanitized
	}

	// Default sort
	return "f.pe_ratio ASC"
}

// Common filter presets for easy usage
var (
	// ValueStocks finds stocks with low PE and high ROE
	ValueStocks = NewFilterBuilder().
			PELessThan(15).
			ROEGreaterThan(0.15)

	// DividendStocks finds stocks with good dividend yield and growth
	DividendStocks = NewFilterBuilder().
			DividendYieldGreaterThan(0.03).
			DividendGrowthGreaterThan(0.05)

	// UndervaluedStocks finds stocks trading below intrinsic value
	UndervaluedStocks = NewFilterBuilder().
				MarginOfSafetyGreaterThan(0.20).
				IntrinsicValueGreaterThan(0)

	// GrowthStocks finds stocks with high ROE and positive outlook
	GrowthStocks = NewFilterBuilder().
			ROEGreaterThan(0.20).
			EarningsOutlook("positive")

	// BargainStocks finds cheap stocks below moving averages
	BargainStocks = NewFilterBuilder().
			PELessThan(10).
			PriceBelowSMA200()
)
