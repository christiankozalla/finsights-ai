package eodhd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	apiToken string
	baseURL  string
	client   *http.Client
	cache    *Cache
}

func NewClient(apiToken string, cachePath string) (*Client, error) {
	cache, err := NewCache(cachePath)
	if err != nil {
		return nil, err
	}
	return &Client{
		apiToken: apiToken,
		baseURL:  "https://eodhd.com/api",
		client:   &http.Client{Timeout: 10 * time.Second},
		cache:    cache,
	}, nil
}

func (c *Client) get(endpoint string, params url.Values, v any) error {
	params.Set("api_token", c.apiToken)
	params.Set("fmt", "json")
	fullURL := fmt.Sprintf("%s/%s?%s", c.baseURL, endpoint, params.Encode())

	// Try cache first
	if c.cache != nil {
		found, err := c.cache.Get(fullURL, v)
		if err != nil {
			return fmt.Errorf("cache get error: %w", err)
		}
		if found {
			return nil
		}
	}

	// Fetch from API
	resp, err := c.client.Get(fullURL)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s", body)
	}

	// Decode and cache
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	if c.cache != nil {
		_ = c.cache.Set(fullURL, v, 24*time.Hour) // 1d TTL
	}

	return nil
}

type EODData struct {
	Date          string  `json:"date"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	AdjustedClose float64 `json:"adjusted_close"`
	Volume        int64   `json:"volume"`
}

// GetEODData retrieves historical End-of-Day data for a given ticker and optional date range.
// from and to in YYYY-MM-DD format.
func (c *Client) GetEODData(ticker string, from, to string) ([]EODData, error) {
	endpoint := fmt.Sprintf("eod/%s", ticker)
	params := url.Values{}
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}

	var result []EODData
	err := c.get(endpoint, params, &result)
	return result, err
}

// Fundamentals is a wrapper over raw fundamentals data
type Fundamentals struct {
	raw map[string]any
}

// GetFloat returns a float64 from a "::" path like "Earnings::History::2023-12-31::epsActual"
func (f *Fundamentals) GetFloat(path string) float64 {
	keys := strings.Split(path, "::")
	current := f.raw

	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].(float64); ok {
				return val
			}
			return 0
		}

		next, ok := current[key].(map[string]any)
		if !ok {
			return 0
		}
		current = next
	}
	return 0
}

func (c *Client) GetFundamentalsRaw(ticker string) (*Fundamentals, error) {
	endpoint := fmt.Sprintf("fundamentals/%s", ticker)
	params := url.Values{}

	var raw map[string]interface{}
	if err := c.get(endpoint, params, &raw); err != nil {
		return nil, err
	}

	return &Fundamentals{raw: raw}, nil
}

// SearchResult represents a single search result
type SearchResult struct {
	Code     string `json:"Code"`
	Name     string `json:"Name"`
	Exchange string `json:"Exchange"`
	Country  string `json:"Country"`
	Type     string `json:"Type"`
}

// SearchStocks looks up tickers by query
func (c *Client) SearchStocks(query string, limit int) ([]SearchResult, error) {
	endpoint := "search"
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))

	var results []SearchResult
	err := c.get(endpoint, params, &results)
	return results, err
}

// Screener
type ScreenerResult struct {
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	Exchange           string  `json:"exchange"`
	MarketCap          float64 `json:"market_capitalization"`
	DividendYield      float64 `json:"dividend_yield"`
	EarningsShare      float64 `json:"earnings_share"`
	Sector             string  `json:"sector"`
	Industry           string  `json:"industry"`
	AdjustedClosePrice float64 `json:"adjusted_close"`
}

// ScreenerFilter allows for filtering
type ScreenerFilter struct {
	Filters string
	Sort    string
	Limit   int
	Offset  int
}

// ScreenStocks executes a market screener query
func (c *Client) ScreenStocks(filter ScreenerFilter) ([]ScreenerResult, error) {
	endpoint := "screener"
	params := url.Values{}
	params.Set("filters", filter.Filters)
	params.Set("sort", filter.Sort)
	params.Set("limit", fmt.Sprintf("%d", filter.Limit))
	params.Set("offset", fmt.Sprintf("%d", filter.Offset))

	var response struct {
		Data []ScreenerResult `json:"data"`
	}
	err := c.get(endpoint, params, &response)
	return response.Data, err
}

// Stock Details
type FundamentalsGeneral struct {
	Code        string `json:"Code"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Industry    string `json:"Industry"`
	Sector      string `json:"Sector"`
	IPODate     string `json:"IPODate"`
	ISIN        string `json:"ISIN"`
	Country     string `json:"CountryName"`
}

func (c *Client) GetFundamentalsGeneral(ticker string) (FundamentalsGeneral, error) {
	endpoint := fmt.Sprintf("fundamentals/%s", ticker)
	params := url.Values{}
	params.Set("filter", "General")

	var result FundamentalsGeneral
	err := c.get(endpoint, params, &result)
	return result, err
}
