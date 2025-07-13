package screener

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	schema := `
		CREATE TABLE IF NOT EXISTS fundamentals (
			ticker TEXT PRIMARY KEY,
			pe_ratio REAL,
			roe REAL,
			yoy_profit JSON,
			yoy_turnover JSON,
			earnings_outlook TEXT,
			updated_at TEXT,
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
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Insert test data
	testData := `
		INSERT INTO fundamentals (ticker, pe_ratio, roe, earnings_outlook, dividend_yield, dividend_growth_5y, intrinsic_value, margin_of_safety) VALUES
		('AAPL', 14.5, 0.25, 'positive', 0.005, 0.08, 180.50, 0.25),
		('GOOGL', 13.1, 0.18, 'positive', 0.0, 0.0, 3100.0, 0.15),
		('MSFT', 12.5, 0.22, 'positive', 0.035, 0.12, 375.0, 0.22),
		('TSLA', 45.2, 0.15, 'neutral', 0.0, 0.0, 800.0, -0.05),
		('IBM', 8.3, 0.08, 'negative', 0.045, 0.08, 120.0, 0.35),
		('KO', 9.7, 0.16, 'positive', 0.045, 0.08, 65.0, 0.25),
		('JNJ', 11.2, 0.18, 'positive', 0.038, 0.06, 170.0, 0.18),
		('PFE', 7.8, 0.12, 'positive', 0.055, 0.10, 55.0, 0.30);

		INSERT INTO prices (ticker, date, close, sma50, sma200) VALUES
		('AAPL', '2024-01-15', 150.25, 145.80, 140.30),
		('GOOGL', '2024-01-15', 2750.80, 2720.50, 2680.20),
		('MSFT', '2024-01-15', 330.59, 325.20, 315.80),
		('TSLA', '2024-01-15', 220.45, 235.60, 245.90),
		('IBM', '2024-01-15', 78.20, 82.40, 85.10),
		('KO', '2024-01-15', 48.75, 52.20, 55.50),
		('JNJ', '2024-01-15', 158.30, 162.10, 165.80),
		('PFE', '2024-01-15', 42.15, 45.20, 48.90);
	`

	if _, err := db.Exec(testData); err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	return db
}

func TestScreenStocks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name          string
		filter        ScreenerFilter
		expectedCount int
		expectedFirst string
	}{
		{
			name: "no filters",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{},
				Sort:       "pe_ratio ASC",
				Limit:      10,
				Offset:     0,
			},
			expectedCount: 8,
			expectedFirst: "PFE",
		},
		{
			name: "PE ratio filter",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{
					{Field: "pe_ratio", Operator: "<", Value: 20.0},
				},
				Sort:   "pe_ratio ASC",
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 7,
			expectedFirst: "PFE",
		},
		{
			name: "ROE filter",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{
					{Field: "roe", Operator: ">", Value: 0.15},
				},
				Sort:   "roe DESC",
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 5,
			expectedFirst: "KO",
		},
		{
			name: "multiple conditions",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{
					{Field: "pe_ratio", Operator: "<", Value: 30.0},
					{Field: "roe", Operator: ">", Value: 0.15},
				},
				Sort:   "pe_ratio ASC",
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 5,
			expectedFirst: "KO",
		},
		{
			name: "dividend yield filter",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{
					{Field: "dividend_yield", Operator: ">", Value: 0.01},
				},
				Sort:   "dividend_yield DESC",
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 5,
			expectedFirst: "PFE",
		},
		{
			name: "earnings outlook filter",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{
					{Field: "earnings_outlook", Operator: "=", Value: "positive"},
				},
				Sort:   "ticker ASC",
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 6,
			expectedFirst: "PFE",
		},
		{
			name: "margin of safety filter",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{
					{Field: "margin_of_safety", Operator: ">", Value: 0.15},
				},
				Sort:   "margin_of_safety DESC",
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 6,
			expectedFirst: "PFE",
		},
		{
			name: "price below SMA50",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{
					{Field: "price_vs_sma50", Operator: "<", Value: 1.0},
				},
				Sort:   "ticker ASC",
				Limit:  10,
				Offset: 0,
			},
			expectedCount: 5,
			expectedFirst: "PFE",
		},
		{
			name: "pagination test",
			filter: ScreenerFilter{
				Conditions: []FilterCondition{},
				Sort:       "ticker ASC",
				Limit:      2,
				Offset:     2,
			},
			expectedCount: 2,
			expectedFirst: "KO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ScreenStocks(db, tt.filter)
			if err != nil {
				t.Fatalf("ScreenStocks failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			if len(results) > 0 && results[0].Ticker != tt.expectedFirst {
				t.Errorf("Expected first result to be %s, got %s", tt.expectedFirst, results[0].Ticker)
			}
		})
	}
}

func TestFilterBuilder(t *testing.T) {
	tests := []struct {
		name               string
		buildFunc          func(*FilterBuilder) *FilterBuilder
		expectedConditions int
	}{
		{
			name: "PE less than",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PELessThan(20.0)
			},
			expectedConditions: 1,
		},
		{
			name: "PE between",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PEBetween(10.0, 25.0)
			},
			expectedConditions: 2,
		},
		{
			name: "ROE greater than",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.ROEGreaterThan(0.15)
			},
			expectedConditions: 1,
		},
		{
			name: "ROE between",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.ROEBetween(0.10, 0.25)
			},
			expectedConditions: 2,
		},
		{
			name: "price below SMA",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PriceBelowSMA50().PriceBelowSMA200()
			},
			expectedConditions: 2,
		},
		{
			name: "dividend filters",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.DividendYieldGreaterThan(0.02).DividendGrowthGreaterThan(0.05)
			},
			expectedConditions: 2,
		},
		{
			name: "valuation filters",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.MarginOfSafetyGreaterThan(0.15).IntrinsicValueGreaterThan(100.0)
			},
			expectedConditions: 2,
		},
		{
			name: "earnings outlook",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.EarningsOutlook("positive")
			},
			expectedConditions: 1,
		},
		{
			name: "ticker filters",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.Ticker("AAPL").TickerIn([]string{"AAPL", "GOOGL", "MSFT"})
			},
			expectedConditions: 2,
		},
		{
			name: "complex combination",
			buildFunc: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PELessThan(25.0).
					ROEGreaterThan(0.15).
					DividendYieldGreaterThan(0.0).
					MarginOfSafetyGreaterThan(0.10)
			},
			expectedConditions: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilterBuilder()
			filter := tt.buildFunc(builder).Build()

			if len(filter.Conditions) != tt.expectedConditions {
				t.Errorf("Expected %d conditions, got %d", tt.expectedConditions, len(filter.Conditions))
			}
		})
	}
}

func TestFilterBuilderWithPagination(t *testing.T) {
	builder := NewFilterBuilder()
	filter := builder.PELessThan(20.0).BuildWithPagination("roe DESC", 25, 50)

	if filter.Sort != "roe DESC" {
		t.Errorf("Expected sort 'roe DESC', got '%s'", filter.Sort)
	}

	if filter.Limit != 25 {
		t.Errorf("Expected limit 25, got %d", filter.Limit)
	}

	if filter.Offset != 50 {
		t.Errorf("Expected offset 50, got %d", filter.Offset)
	}
}

func TestParseFilterFromJSON(t *testing.T) {
	tests := []struct {
		name           string
		filterJSON     string
		expectedLength int
		expectError    bool
	}{
		{
			name:           "empty filter",
			filterJSON:     "",
			expectedLength: 0,
			expectError:    false,
		},
		{
			name:           "single condition",
			filterJSON:     `[["pe_ratio","<",20]]`,
			expectedLength: 1,
			expectError:    false,
		},
		{
			name:           "multiple conditions",
			filterJSON:     `[["pe_ratio","<",20],["roe",">",0.15]]`,
			expectedLength: 2,
			expectError:    false,
		},
		{
			name:           "field mapping",
			filterJSON:     `[["market_capitalization",">",1000000]]`,
			expectedLength: 1,
			expectError:    false,
		},
		{
			name:        "invalid JSON",
			filterJSON:  `invalid json`,
			expectError: true,
		},
		{
			name:        "invalid condition format",
			filterJSON:  `[["pe_ratio","<"]]`,
			expectError: true,
		},
		{
			name:        "non-string field",
			filterJSON:  `[[123,"<",20]]`,
			expectError: true,
		},
		{
			name:        "non-string operator",
			filterJSON:  `[["pe_ratio",123,20]]`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := ParseFilterFromJSON(tt.filterJSON)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if len(filter.Conditions) != tt.expectedLength {
				t.Errorf("Expected %d conditions, got %d", tt.expectedLength, len(filter.Conditions))
			}
		})
	}
}

func TestMapFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"market_capitalization", "market_cap"},
		{"dividend_yield", "dividend_yield"},
		{"earnings_share", "eps"},
		{"pe_ratio", "pe_ratio"},           // no mapping
		{"unknown_field", "unknown_field"}, // no mapping
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("mapFieldName(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPresetFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	presets := []struct {
		name    string
		builder *FilterBuilder
	}{
		{"ValueStocks", ValueStocks},
		{"DividendStocks", DividendStocks},
		{"UndervaluedStocks", UndervaluedStocks},
		{"GrowthStocks", GrowthStocks},
		{"BargainStocks", BargainStocks},
	}

	for _, preset := range presets {
		t.Run(preset.name, func(t *testing.T) {
			filter := preset.builder.Build()
			results, err := ScreenStocks(db, filter)
			if err != nil {
				t.Errorf("Preset filter %s failed: %v", preset.name, err)
			}

			// Just verify it doesn't crash and returns valid structure
			if results == nil {
				t.Errorf("Preset filter %s returned nil results", preset.name)
			}

			// Log results for debugging (can be removed later)
			t.Logf("Preset %s returned %d results", preset.name, len(results))

			// Verify each result has valid structure
			for _, result := range results {
				if result.Ticker == "" {
					t.Errorf("Found result with empty ticker in preset %s", preset.name)
				}
			}
		})
	}
}

func TestBuildSQLCondition(t *testing.T) {
	tests := []struct {
		name             string
		condition        FilterCondition
		expectedSQL      string
		expectedHasValue bool
	}{
		{
			name:             "price below SMA50",
			condition:        FilterCondition{Field: "price_vs_sma50", Operator: "<", Value: 1.0},
			expectedSQL:      "p.close < p.sma50",
			expectedHasValue: false,
		},
		{
			name:             "price above SMA200",
			condition:        FilterCondition{Field: "price_vs_sma200", Operator: ">", Value: 1.0},
			expectedSQL:      "p.close > p.sma200",
			expectedHasValue: false,
		},
		{
			name:             "intrinsic value greater than price",
			condition:        FilterCondition{Field: "intrinsic_vs_price", Operator: ">", Value: 1.0},
			expectedSQL:      "f.intrinsic_value > p.close",
			expectedHasValue: false,
		},
		{
			name:             "standard fundamentals field",
			condition:        FilterCondition{Field: "pe_ratio", Operator: "<", Value: 20.0},
			expectedSQL:      "f.pe_ratio < ?",
			expectedHasValue: true,
		},
		{
			name:             "standard prices field",
			condition:        FilterCondition{Field: "close", Operator: ">", Value: 100.0},
			expectedSQL:      "p.close > ?",
			expectedHasValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlCondition, value := buildSQLCondition(tt.condition)

			if sqlCondition != tt.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tt.expectedSQL, sqlCondition)
			}

			hasValue := value != nil
			if hasValue != tt.expectedHasValue {
				t.Errorf("Expected hasValue %v, got %v", tt.expectedHasValue, hasValue)
			}
		})
	}
}

func TestSanitizeSort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"pe_ratio.asc", "f.pe_ratio ASC"},
		{"pe_ratio.desc", "f.pe_ratio DESC"},
		{"roe.asc", "f.roe ASC"},
		{"close.desc", "p.close DESC"},
		{"invalid_sort", "f.pe_ratio ASC"}, // default
		{"", "f.pe_ratio ASC"},             // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeSort(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeSort(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsFieldInTables(t *testing.T) {
	fundamentalsTests := []struct {
		field    string
		expected bool
	}{
		{"ticker", true},
		{"pe_ratio", true},
		{"roe", true},
		{"dividend_yield", true},
		{"close", false},
		{"sma50", false},
		{"unknown", false},
	}

	for _, tt := range fundamentalsTests {
		t.Run("fundamentals_"+tt.field, func(t *testing.T) {
			result := isFieldInFundamentals(tt.field)
			if result != tt.expected {
				t.Errorf("isFieldInFundamentals(%s) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}

	pricesTests := []struct {
		field    string
		expected bool
	}{
		{"close", true},
		{"sma50", true},
		{"sma200", true},
		{"ticker", false},
		{"pe_ratio", false},
		{"unknown", false},
	}

	for _, tt := range pricesTests {
		t.Run("prices_"+tt.field, func(t *testing.T) {
			result := isFieldInPrices(tt.field)
			if result != tt.expected {
				t.Errorf("isFieldInPrices(%s) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkScreenStocks(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	filter := ScreenerFilter{
		Conditions: []FilterCondition{
			{Field: "pe_ratio", Operator: "<", Value: 25.0},
			{Field: "roe", Operator: ">", Value: 0.15},
		},
		Sort:   "pe_ratio ASC",
		Limit:  10,
		Offset: 0,
	}

	for b.Loop() {
		_, err := ScreenStocks(db, filter)
		if err != nil {
			b.Fatalf("ScreenStocks failed: %v", err)
		}
	}
}

func BenchmarkFilterBuilder(b *testing.B) {
	for b.Loop() {
		builder := NewFilterBuilder()
		_ = builder.PELessThan(20.0).
			ROEGreaterThan(0.15).
			DividendYieldGreaterThan(0.02).
			MarginOfSafetyGreaterThan(0.10).
			Build()
	}
}
