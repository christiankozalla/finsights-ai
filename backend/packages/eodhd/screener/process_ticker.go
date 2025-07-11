package screener

import (
	"database/sql"
	"fmt"

	"github.com/finsights-ai/backend/packages/eodhd"
)

func ProcessTicker(db *sql.DB, client *eodhd.Client, ticker string) error {
	// 1. Get historical prices
	prices, err := client.GetEODData(ticker, "", "")
	if err != nil || len(prices) < 200 {
		return fmt.Errorf("not enough EOD data: %v", err)
	}

	// 2. Prepare for SMA calculation
	eod := []EOD{}
	for _, p := range prices {
		eod = append(eod, EOD{Date: p.Date, Close: p.AdjustedClose})
	}

	sma50, _ := CalculateSMA(eod, 50)
	sma200, _ := CalculateSMA(eod, 200)

	latest := eod[0]
	_ = SaveSMA(db, ticker, latest.Date, latest.Close, sma50, sma200)

	// 3. Get fundamentals
	fund, err := client.GetFundamentalsRaw(ticker)
	if err != nil {
		return fmt.Errorf("error getting fundamentals: %v", err)
	}

	// 4. Calculate PE and ROE
	eps := fund.GetFloat("Earnings::History::2023-12-31::epsActual")
	price := latest.Close
	pe := price / eps

	netIncome := fund.GetFloat("Financials::Income_Statement::yearly::2023-12-31::netIncome")
	equity := fund.GetFloat("Financials::Balance_Sheet::yearly::2023-12-31::totalStockholderEquity")
	roe, _ := CalculateROE(netIncome, equity)

	// 5. Save ROE and PE
	// outlook := ExtractOutlookFromNews(ticker) // optionally
	return SaveROE(db, ticker, roe, pe, "")
}
