package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/finsights-ai/backend/packages/eodhd"
)

// MockClient implements a mock EODHD client for testing
type MockClient struct {
	screenStocksFunc func(filter eodhd.ScreenerFilter) ([]eodhd.ScreenerResult, error)
}

func (m *MockClient) ScreenStocks(filter eodhd.ScreenerFilter) ([]eodhd.ScreenerResult, error) {
	if m.screenStocksFunc != nil {
		return m.screenStocksFunc(filter)
	}
	return nil, nil
}

func TestGetScreenerData(t *testing.T) {
	// Mock data
	mockResults := []eodhd.ScreenerResult{
		{
			Code:               "AAPL",
			Name:               "Apple Inc.",
			Exchange:           "NASDAQ",
			MarketCap:          3000000000000,
			DividendYield:      0.005,
			EarningsShare:      6.05,
			Sector:             "Technology",
			Industry:           "Consumer Electronics",
			AdjustedClosePrice: 150.25,
		},
		{
			Code:               "GOOGL",
			Name:               "Alphabet Inc.",
			Exchange:           "NASDAQ",
			MarketCap:          1800000000000,
			DividendYield:      0.0,
			EarningsShare:      5.61,
			Sector:             "Technology",
			Industry:           "Internet Content & Information",
			AdjustedClosePrice: 2750.80,
		},
	}

	tests := []struct {
		name            string
		queryParams     url.Values
		mockFunc        func(filter eodhd.ScreenerFilter) ([]eodhd.ScreenerResult, error)
		expectedStatus  int
		expectedPage    int
		expectedLimit   int
		expectedHasMore bool
	}{
		{
			name:        "successful request with default params",
			queryParams: url.Values{},
			mockFunc: func(filter eodhd.ScreenerFilter) ([]eodhd.ScreenerResult, error) {
				return mockResults, nil
			},
			expectedStatus:  http.StatusOK,
			expectedPage:    1,
			expectedLimit:   50,
			expectedHasMore: false,
		},
		{
			name: "successful request with custom pagination",
			queryParams: url.Values{
				"page":  {"2"},
				"limit": {"1"},
			},
			mockFunc: func(filter eodhd.ScreenerFilter) ([]eodhd.ScreenerResult, error) {
				// Return 2 results to test hasMore logic
				return append(mockResults, eodhd.ScreenerResult{
					Code:               "MSFT",
					Name:               "Microsoft Corporation",
					Exchange:           "NASDAQ",
					MarketCap:          2800000000000,
					DividendYield:      0.007,
					EarningsShare:      8.05,
					Sector:             "Technology",
					Industry:           "Software",
					AdjustedClosePrice: 330.59,
				}), nil
			},
			expectedStatus:  http.StatusOK,
			expectedPage:    2,
			expectedLimit:   1,
			expectedHasMore: true,
		},
		{
			name: "successful request with custom filters and sort",
			queryParams: url.Values{
				"filters": {`[["market_capitalization",">",1000000000]]`},
				"sort":    {"market_capitalization.asc"},
				"limit":   {"10"},
			},
			mockFunc: func(filter eodhd.ScreenerFilter) ([]eodhd.ScreenerResult, error) {
				// Verify filter parameters
				if filter.Filters != `[["market_capitalization",">",1000000000]]` {
					t.Errorf("Expected filters to be %q, got %s", `[["market_capitalization",">",1000000000]]`, filter.Filters)
				}
				if filter.Sort != "market_capitalization.asc" {
					t.Errorf("Expected sort to be 'market_capitalization.asc', got %s", filter.Sort)
				}
				if filter.Limit != 11 { // +1 for hasMore check
					t.Errorf("Expected limit to be 11, got %d", filter.Limit)
				}
				return mockResults, nil
			},
			expectedStatus:  http.StatusOK,
			expectedPage:    1,
			expectedLimit:   10,
			expectedHasMore: false,
		},
		{
			name: "invalid page parameter",
			queryParams: url.Values{
				"page": {"0"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid limit parameter",
			queryParams: url.Values{
				"limit": {"1001"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-numeric page parameter",
			queryParams: url.Values{
				"page": {"abc"},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock client
			mockClient := &MockClient{
				screenStocksFunc: tt.mockFunc,
			}

			// Create handler
			handler := NewScreenerHandler(mockClient)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/screener", nil)
			req.URL.RawQuery = tt.queryParams.Encode()

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.GetScreenerData(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			// For successful requests, check response body
			if tt.expectedStatus == http.StatusOK {
				var response ScreenerResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Page != tt.expectedPage {
					t.Errorf("Expected page %d, got %d", tt.expectedPage, response.Page)
				}

				if response.Limit != tt.expectedLimit {
					t.Errorf("Expected limit %d, got %d", tt.expectedLimit, response.Limit)
				}

				if response.HasMore != tt.expectedHasMore {
					t.Errorf("Expected hasMore %v, got %v", tt.expectedHasMore, response.HasMore)
				}

				// Check content type
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
				}
			}

			// For error requests, check error response
			if tt.expectedStatus != http.StatusOK {
				var errorResponse ErrorResponse
				if err := json.NewDecoder(rr.Body).Decode(&errorResponse); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if errorResponse.Error == "" {
					t.Error("Expected error code in response")
				}

				if errorResponse.Message == "" {
					t.Error("Expected error message in response")
				}
			}
		})
	}
}

func TestGetScreenerDataMethodNotAllowed(t *testing.T) {
	mockClient := &MockClient{}
	handler := NewScreenerHandler(mockClient)

	req := httptest.NewRequest(http.MethodPost, "/api/screener", nil)
	rr := httptest.NewRecorder()

	handler.GetScreenerData(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestParseIntParam(t *testing.T) {
	tests := []struct {
		name         string
		param        string
		defaultValue int
		expected     int
		expectError  bool
	}{
		{
			name:         "empty param uses default",
			param:        "",
			defaultValue: 10,
			expected:     10,
			expectError:  false,
		},
		{
			name:         "valid integer",
			param:        "42",
			defaultValue: 10,
			expected:     42,
			expectError:  false,
		},
		{
			name:         "invalid integer",
			param:        "abc",
			defaultValue: 10,
			expected:     0,
			expectError:  true,
		},
		{
			name:         "negative integer",
			param:        "-5",
			defaultValue: 10,
			expected:     -5,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIntParam(tt.param, tt.defaultValue)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFilterBuilder(t *testing.T) {
	tests := []struct {
		name     string
		buildFn  func(*FilterBuilder) *FilterBuilder
		expected string
	}{
		{
			name: "single filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.Exchange("US")
			},
			expected: `[["exchange","=","US"]]`,
		},
		{
			name: "multiple filters",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.Exchange("US").Sector("Technology").MarketCapGreaterThan(1000000000)
			},
			expected: `[["exchange","=","US"],["sector","=","Technology"],["market_capitalization",">",1000000000]]`,
		},
		{
			name: "dividend yield filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.DividendYieldGreaterThan(0.02)
			},
			expected: `[["dividend_yield",">",0.02]]`,
		},
		{
			name: "earnings share filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.EarningsShareGreaterThan(5.0)
			},
			expected: `[["earnings_share",">",5]]`,
		},
		{
			name: "ETF type filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.Type("etf").AvgVolume200DGreaterThan(50000)
			},
			expected: `[["type","=","etf"],["avgvol_200d",">",50000]]`,
		},
		{
			name: "custom filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.AddFilter("custom_field", ">=", 100)
			},
			expected: `[["custom_field",">=",100]]`,
		},
		{
			name: "complex filter combination",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.Exchange("XETRA").
					MarketCapGreaterThan(10000000000).
					DividendYieldGreaterThan(0.01).
					Return5DGreaterThan(5.0)
			},
			expected: `[["exchange","=","XETRA"],["market_capitalization",">",10000000000],["dividend_yield",">",0.01],["refund_5d_p",">",5]]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilterBuilder()
			result, err := tt.buildFn(builder).Build()

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected: %s, got: %s", tt.expected, result)
			}
		})
	}
}

func TestFilterBuilderEmpty(t *testing.T) {
	builder := NewFilterBuilder()
	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty string, got: %s", result)
	}
}

func TestExampleFilters(t *testing.T) {
	tests := []struct {
		name   string
		filter string
	}{
		{"XETRAHighCap", ExampleFilters.XETRAHighCap},
		{"USTechDividend", ExampleFilters.USTechDividend},
		{"EnergyEarnings", ExampleFilters.EnergyEarnings},
		{"SmallCapReturns", ExampleFilters.SmallCapReturns},
		{"ETFHighVolume", ExampleFilters.ETFHighVolume},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the filter is valid JSON
			var filters [][]any
			err := json.Unmarshal([]byte(tt.filter), &filters)
			if err != nil {
				t.Errorf("Filter %s is not valid JSON: %v", tt.name, err)
			}

			// Test that each filter has the correct structure
			for i, filter := range filters {
				if len(filter) != 3 {
					t.Errorf("Filter %s[%d] should have 3 elements, got %d", tt.name, i, len(filter))
				}
			}
		})
	}
}
