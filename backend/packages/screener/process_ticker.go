package screener

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

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

	period := fund.GetLatestPeriod("Financials::Balance_Sheet::yearly")
	if period == "" {
		log.Fatal("No financial data available")
	}

	equity := fund.GetFloat(fmt.Sprintf("Financials::Balance_Sheet::yearly::%s::totalStockholderEquity", period))
	netIncome := fund.GetFloat(fmt.Sprintf("Financials::Income_Statement::yearly::%s::netIncome", period))
	roe, _ := CalculateROE(netIncome, equity)

	// Calculate EPS growth rate (CAGR) from EPS 5 years ago to latest
	epsPast := fund.GetFloat("Earnings::History::2018-12-31::epsActual")
	growthRate := calculateCAGR(epsPast, eps, 5)
	if growthRate == 0 {
		growthRate = 0.05 // Fallback to 5% conservative estimate
	}

	bondYield := 4.4 // Conservative fixed value. Can be dynamic if needed

	today := time.Now().Format("2006-01-02")
	divs, err := client.GetDividends(ticker, "2014-01-01", today)
	if err != nil {
		return fmt.Errorf("error getting dividends: %v", err)
	}

	divPerShareLast := sumOfDividendsForYear(divs, 2023) // TODO: years need to be dynamic
	divPerSharePast := sumOfDividendsForYear(divs, 2018)

	divYield := CalculateDividendYield(divPerShareLast, price)
	divGrowth := CalculateDividendCAGR(divPerSharePast, divPerShareLast, 5)

	intrinsic, _ := CalculateIntrinsicValue(eps, growthRate, bondYield)
	safetyMargin := CalculateMarginOfSafety(intrinsic, price)

	SaveValuationMetrics(db, ticker, divYield, divGrowth, intrinsic, safetyMargin)

	// 5. Save ROE and PE
	// outlook := ExtractOutlookFromNews(ticker) // optionally
	return SaveROE(db, ticker, roe, pe, "")
}

func sumOfDividendsForYear(divs []eodhd.Dividend, year int) float64 {
	total := 0.0
	for _, d := range divs {
		if strings.HasPrefix(d.Date, fmt.Sprintf("%d", year)) {
			total += d.Value
		}
	}
	return total
}

func calculateCAGR(start, end float64, years int) float64 {
	if start <= 0 || end <= 0 || years <= 0 {
		return 0
	}
	return math.Pow(end/start, 1.0/float64(years)) - 1
}
