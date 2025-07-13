package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/finsights-ai/backend/packages/screener"
)

type ScreenerClient interface {
	ScreenStocks(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error)
}

// DatabaseScreenerClient implements ScreenerClient using the database
type DatabaseScreenerClient struct {
	db *sql.DB
}

func NewDatabaseScreenerClient(db *sql.DB) *DatabaseScreenerClient {
	return &DatabaseScreenerClient{db: db}
}

func (c *DatabaseScreenerClient) ScreenStocks(filter screener.ScreenerFilter) ([]screener.ScreenerResult, error) {
	return screener.ScreenStocks(c.db, filter)
}

type ScreenerHandler struct {
	client ScreenerClient
}

func NewScreenerHandler(client ScreenerClient) *ScreenerHandler {
	return &ScreenerHandler{
		client: client,
	}
}

type ScreenerResponse struct {
	Data       []screener.ScreenerResult `json:"data"`
	Page       int                       `json:"page"`
	Limit      int                       `json:"limit"`
	TotalCount int                       `json:"total_count"`
	HasMore    bool                      `json:"has_more"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (h *ScreenerHandler) GetScreenerData(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Parse pagination parameters
	page, err := parseIntParam(query.Get("page"), 1)
	if err != nil || page < 1 {
		h.sendError(w, http.StatusBadRequest, "INVALID_PAGE", "Page must be a positive integer")
		return
	}

	limit, err := parseIntParam(query.Get("limit"), 50)
	if err != nil || limit < 1 || limit > 1000 {
		h.sendError(w, http.StatusBadRequest, "INVALID_LIMIT", "Limit must be between 1 and 1000")
		return
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Parse filter parameters
	filters := query.Get("filters")

	// Parse the filter from JSON format
	baseFilter, err := screener.ParseFilterFromJSON(filters)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "INVALID_FILTER", "Invalid filter format: "+err.Error())
		return
	}

	// Parse sort parameter
	sort := query.Get("sort")
	if sort == "" {
		sort = "pe_ratio.asc" // Default sort
	}

	// Create final filter with pagination
	filter := screener.ScreenerFilter{
		Conditions: baseFilter.Conditions,
		Sort:       sort,
		Limit:      limit + 1, // Request one extra to check if there are more results
		Offset:     offset,
	}

	// Call custom screener
	results, err := h.client.ScreenStocks(filter)
	if err != nil {
		log.Printf("Error calling ScreenStocks: %v", err)
		h.sendError(w, http.StatusInternalServerError, "SCREENER_ERROR", "Failed to fetch screener data")
		return
	}

	// Determine if there are more results
	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit] // Remove the extra result
	}

	// Create response
	response := ScreenerResponse{
		Data:       results,
		Page:       page,
		Limit:      limit,
		TotalCount: len(results), // Current page size
		HasMore:    hasMore,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		h.sendError(w, http.StatusInternalServerError, "ENCODING_ERROR", "Failed to encode response")
		return
	}
}

func (h *ScreenerHandler) sendError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Error:   errorCode,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Error encoding error response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func parseIntParam(param string, defaultValue int) (int, error) {
	if param == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(param)
}

// FilterBuilder helps construct filter queries in the correct JSON array format
type FilterBuilder struct {
	filters [][]any
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		filters: make([][]any, 0),
	}
}

// AddFilter adds a filter condition [field, operator, value]
func (fb *FilterBuilder) AddFilter(field, operator string, value any) *FilterBuilder {
	fb.filters = append(fb.filters, []any{field, operator, value})
	return fb
}

// Build returns the JSON string representation of the filters
func (fb *FilterBuilder) Build() (string, error) {
	if len(fb.filters) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(fb.filters); err != nil {
		return "", err
	}

	// Remove the trailing newline that Encode adds
	result := buf.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// Common filter builder methods for convenience
func (fb *FilterBuilder) PELessThan(value float64) *FilterBuilder {
	return fb.AddFilter("pe_ratio", "<", value)
}

func (fb *FilterBuilder) PEGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("pe_ratio", ">", value)
}

func (fb *FilterBuilder) ROEGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("roe", ">", value)
}

func (fb *FilterBuilder) ROELessThan(value float64) *FilterBuilder {
	return fb.AddFilter("roe", "<", value)
}

func (fb *FilterBuilder) DividendYieldGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("dividend_yield", ">", value)
}

func (fb *FilterBuilder) MarginOfSafetyGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("margin_of_safety", ">", value)
}

func (fb *FilterBuilder) PriceBelowSMA50() *FilterBuilder {
	return fb.AddFilter("price_vs_sma50", "<", 1.0)
}

func (fb *FilterBuilder) PriceBelowSMA200() *FilterBuilder {
	return fb.AddFilter("price_vs_sma200", "<", 1.0)
}

func (fb *FilterBuilder) EarningsOutlook(outlook string) *FilterBuilder {
	return fb.AddFilter("earnings_outlook", "=", outlook)
}

func (fb *FilterBuilder) Ticker(ticker string) *FilterBuilder {
	return fb.AddFilter("ticker", "=", ticker)
}

var (
	// Common filter presets
	DefaultFilters = `[["pe_ratio","<",20]]`

	// Example filter builders
	ExampleFilters = struct {
		ValueStocks       string
		DividendStocks    string
		UndervaluedStocks string
		GrowthStocks      string
		BargainStocks     string
	}{
		ValueStocks:       `[["pe_ratio","<",15],["roe",">",0.15]]`,
		DividendStocks:    `[["dividend_yield",">",0.03],["dividend_growth_5y",">",0.05]]`,
		UndervaluedStocks: `[["margin_of_safety",">",0.20],["intrinsic_value",">",0]]`,
		GrowthStocks:      `[["roe",">",0.20],["earnings_outlook","=","positive"]]`,
		BargainStocks:     `[["pe_ratio","<",10],["price_vs_sma200","<",1.0]]`,
	}
)
