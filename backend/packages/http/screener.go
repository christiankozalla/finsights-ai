package http

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/finsights-ai/backend/packages/eodhd"
)

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
func (fb *FilterBuilder) Exchange(exchange string) *FilterBuilder {
	return fb.AddFilter("exchange", "=", exchange)
}

func (fb *FilterBuilder) MarketCapGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("market_capitalization", ">", value)
}

func (fb *FilterBuilder) MarketCapLessThan(value float64) *FilterBuilder {
	return fb.AddFilter("market_capitalization", "<", value)
}

func (fb *FilterBuilder) Sector(sector string) *FilterBuilder {
	return fb.AddFilter("sector", "=", sector)
}

func (fb *FilterBuilder) DividendYieldGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("dividend_yield", ">", value)
}

func (fb *FilterBuilder) EarningsShareGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("earnings_share", ">", value)
}

func (fb *FilterBuilder) Type(assetType string) *FilterBuilder {
	return fb.AddFilter("type", "=", assetType)
}

func (fb *FilterBuilder) AvgVolume200DGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("avgvol_200d", ">", value)
}

func (fb *FilterBuilder) Return5DGreaterThan(value float64) *FilterBuilder {
	return fb.AddFilter("refund_5d_p", ">", value)
}

var (
	// Common filter presets
	DefaultFilters = `[["market_capitalization",">",1000000]]`

	// Example filter builders
	ExampleFilters = struct {
		XETRAHighCap    string
		USTechDividend  string
		EnergyEarnings  string
		SmallCapReturns string
		ETFHighVolume   string
	}{
		XETRAHighCap:    `[["exchange","=","XETRA"],["market_capitalization",">",10000000000]]`,
		USTechDividend:  `[["exchange","=","US"],["sector","=","Technology"],["dividend_yield",">",0.01]]`,
		EnergyEarnings:  `[["sector","=","Energy"],["earnings_share",">",0],["exchange","=","F"]]`,
		SmallCapReturns: `[["market_capitalization","<",100000000],["refund_5d_p",">",10]]`,
		ETFHighVolume:   `[["exchange","=","F"],["type","=","etf"],["avgvol_200d",">",50000]]`,
	}
)

type ScreenerClient interface {
	ScreenStocks(filter eodhd.ScreenerFilter) ([]eodhd.ScreenerResult, error)
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
	Data       []eodhd.ScreenerResult `json:"data"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalCount int                    `json:"total_count"`
	HasMore    bool                   `json:"has_more"`
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
	if filters == "" {
		filters = DefaultFilters // Default filter: market cap > 1M
	}

	sort := query.Get("sort")
	if sort == "" {
		sort = "market_capitalization.desc" // Default sort by market cap descending
	}

	// Create screener filter
	filter := eodhd.ScreenerFilter{
		Filters: filters,
		Sort:    sort,
		Limit:   limit + 1, // Request one extra to check if there are more results
		Offset:  offset,
	}

	// Call EODHD API
	results, err := h.client.ScreenStocks(filter)
	if err != nil {
		log.Printf("Error calling ScreenStocks: %v", err)
		h.sendError(w, http.StatusInternalServerError, "API_ERROR", "Failed to fetch screener data")
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
		TotalCount: len(results), // Note: EODHD doesn't provide total count, so we use current page size
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
