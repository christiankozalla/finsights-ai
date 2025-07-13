package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/finsights-ai/backend/packages/screener"
)

// MockScreenerClient implements a mock screener client for testing
type MockScreenerClient struct {
	screenStocksFunc func(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error)
}

func (m *MockScreenerClient) ScreenStocks(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error) {
	if m.screenStocksFunc != nil {
		return m.screenStocksFunc(filter)
	}
	return nil, nil
}

func TestGetScreenerData(t *testing.T) {
	// Mock data
	mockResults := []screener.ScreenerResult{
		{
			Ticker:           "AAPL",
			PE:               25.5,
			ROE:              0.25,
			Close:            150.25,
			SMA50:            145.80,
			SMA200:           140.30,
			EarningsOutlook:  "positive",
			DividendYield:    0.005,
			DividendGrowth5Y: 0.08,
			IntrinsicValue:   180.50,
			MarginOfSafety:   0.167,
		},
		{
			Ticker:           "GOOGL",
			PE:               22.1,
			ROE:              0.18,
			Close:            2750.80,
			SMA50:            2720.50,
			SMA200:           2680.20,
			EarningsOutlook:  "positive",
			DividendYield:    0.0,
			DividendGrowth5Y: 0.0,
			IntrinsicValue:   3100.00,
			MarginOfSafety:   0.113,
		},
	}

	tests := []struct {
		name            string
		queryParams     url.Values
		mockFunc        func(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error)
		expectedStatus  int
		expectedPage    int
		expectedLimit   int
		expectedHasMore bool
	}{
		{
			name:        "successful request with default params",
			queryParams: url.Values{},
			mockFunc: func(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error) {
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
			mockFunc: func(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error) {
				// Return 2 results to test hasMore logic
				return append(mockResults, screener.ScreenerResult{
					Ticker:           "MSFT",
					PE:               28.5,
					ROE:              0.22,
					Close:            330.59,
					SMA50:            325.20,
					SMA200:           315.80,
					EarningsOutlook:  "positive",
					DividendYield:    0.007,
					DividendGrowth5Y: 0.12,
					IntrinsicValue:   375.00,
					MarginOfSafety:   0.134,
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
				"filters": {`[["pe_ratio","<",30],["roe",">",0.15]]`},
				"sort":    {"pe_ratio.asc"},
				"limit":   {"10"},
			},
			mockFunc: func(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error) {
				// Verify filter parameters
				if len(filter.Conditions) != 2 {
					t.Errorf("Expected 2 filter conditions, got %d", len(filter.Conditions))
				}
				if filter.Sort != "pe_ratio.asc" {
					t.Errorf("Expected sort to be 'pe_ratio.asc', got %s", filter.Sort)
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
		{
			name: "invalid filter JSON",
			queryParams: url.Values{
				"filters": {"invalid json"},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock client
			mockClient := &MockScreenerClient{
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
	mockClient := &MockScreenerClient{}
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
			name: "single PE filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PELessThan(20)
			},
			expected: `[["pe_ratio","<",20]]`,
		},
		{
			name: "multiple filters",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PELessThan(15).ROEGreaterThan(0.15)
			},
			expected: `[["pe_ratio","<",15],["roe",">",0.15]]`,
		},
		{
			name: "dividend yield filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.DividendYieldGreaterThan(0.03)
			},
			expected: `[["dividend_yield",">",0.03]]`,
		},
		{
			name: "margin of safety filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.MarginOfSafetyGreaterThan(0.20)
			},
			expected: `[["margin_of_safety",">",0.2]]`,
		},
		{
			name: "price below SMA filters",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PriceBelowSMA50().PriceBelowSMA200()
			},
			expected: `[["price_vs_sma50","<",1],["price_vs_sma200","<",1]]`,
		},
		{
			name: "earnings outlook filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.EarningsOutlook("positive")
			},
			expected: `[["earnings_outlook","=","positive"]]`,
		},
		{
			name: "ticker filter",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.Ticker("AAPL")
			},
			expected: `[["ticker","=","AAPL"]]`,
		},
		{
			name: "complex filter combination",
			buildFn: func(fb *FilterBuilder) *FilterBuilder {
				return fb.PELessThan(15).
					ROEGreaterThan(0.20).
					DividendYieldGreaterThan(0.02).
					MarginOfSafetyGreaterThan(0.15)
			},
			expected: `[["pe_ratio","<",15],["roe",">",0.2],["dividend_yield",">",0.02],["margin_of_safety",">",0.15]]`,
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
		{"ValueStocks", ExampleFilters.ValueStocks},
		{"DividendStocks", ExampleFilters.DividendStocks},
		{"UndervaluedStocks", ExampleFilters.UndervaluedStocks},
		{"GrowthStocks", ExampleFilters.GrowthStocks},
		{"BargainStocks", ExampleFilters.BargainStocks},
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
			filter, err := screener.ParseFilterFromJSON(tt.filterJSON)

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
