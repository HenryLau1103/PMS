package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"psm-backend/internal/database"
	"time"

	"github.com/shopspring/decimal"
)

type MarketDataService struct {
	db *database.DB
}

func NewMarketDataService(db *database.DB) *MarketDataService {
	return &MarketDataService{db: db}
}

// OHLCV represents candlestick data
type OHLCV struct {
	Symbol    string          `json:"symbol"`
	Timestamp time.Time       `json:"timestamp"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	Volume    int64           `json:"volume"`
	Turnover  decimal.Decimal `json:"turnover"`
}

// TWS E API response structure for daily data
type TWSEDailyData struct {
	Date        string   `json:"date"`        // "20240122"
	TradeVolume string   `json:"TradeVolume"` // Volume
	TradeValue  string   `json:"TradeValue"`  // Turnover
	Open        string   `json:"OpeningPrice"`
	High        string   `json:"HighestPrice"`
	Low         string   `json:"LowestPrice"`
	Close       string   `json:"ClosingPrice"`
	Change      string   `json:"Change"`
	Transaction string   `json:"Transaction"`
}

// FetchDailyData fetches daily OHLCV data from TWSE API
func (s *MarketDataService) FetchDailyData(ctx context.Context, symbol string, startDate, endDate time.Time) ([]OHLCV, error) {
	// TWSE API endpoint for historical data
	// Format: https://www.twse.com.tw/exchangeReport/STOCK_DAY?response=json&date=20240101&stockNo=2330
	
	var allData []OHLCV
	current := startDate

	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("20060102")
		url := fmt.Sprintf("https://www.twse.com.tw/exchangeReport/STOCK_DAY?response=json&date=%s&stockNo=%s", dateStr, symbol)

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch data: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var result struct {
			Stat   string          `json:"stat"`
			Title  string          `json:"title"`
			Fields []string        `json:"fields"`
			Data   [][]string      `json:"data"`
			Notes  []string        `json:"notes"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Parse data rows
		for _, row := range result.Data {
			if len(row) < 9 {
				continue
			}

			// Parse date (format: "113/01/22" -> ROC year/month/day)
			dateStr := row[0]
			timestamp, err := parseROCDate(dateStr)
			if err != nil {
				continue
			}

			ohlcv := OHLCV{
				Symbol:    symbol,
				Timestamp: timestamp,
			}

			// Parse OHLC prices (remove commas)
			if open, err := parseDecimal(row[3]); err == nil {
				ohlcv.Open = open
			}
			if high, err := parseDecimal(row[4]); err == nil {
				ohlcv.High = high
			}
			if low, err := parseDecimal(row[5]); err == nil {
				ohlcv.Low = low
			}
			if close, err := parseDecimal(row[6]); err == nil {
				ohlcv.Close = close
			}

			// Parse volume and turnover
			if volume, err := parseInt64(row[1]); err == nil {
				ohlcv.Volume = volume
			}
			if turnover, err := parseDecimal(row[2]); err == nil {
				ohlcv.Turnover = turnover
			}

			allData = append(allData, ohlcv)
		}

		// Move to next month
		current = current.AddDate(0, 1, 0)

		// Rate limiting - TWSE has request limits
		time.Sleep(3 * time.Second)
	}

	return allData, nil
}

// SaveOHLCV inserts OHLCV data into database
func (s *MarketDataService) SaveOHLCV(ctx context.Context, data []OHLCV) error {
	if len(data) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO stock_ohlcv (symbol, timestamp, open, high, low, close, volume, turnover)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (symbol, timestamp) 
		DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume,
			turnover = EXCLUDED.turnover
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ohlcv := range data {
		_, err := stmt.ExecContext(ctx, 
			ohlcv.Symbol,
			ohlcv.Timestamp,
			ohlcv.Open,
			ohlcv.High,
			ohlcv.Low,
			ohlcv.Close,
			ohlcv.Volume,
			ohlcv.Turnover,
		)
		if err != nil {
			return fmt.Errorf("failed to insert OHLCV for %s at %s: %w", ohlcv.Symbol, ohlcv.Timestamp, err)
		}
	}

	return tx.Commit()
}

// GetOHLCV retrieves OHLCV data from database
func (s *MarketDataService) GetOHLCV(ctx context.Context, symbol string, startDate, endDate time.Time, limit int) ([]OHLCV, error) {
	query := `
		SELECT symbol, timestamp, open, high, low, close, volume, turnover
		FROM stock_ohlcv
		WHERE symbol = $1 
			AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
		LIMIT $4
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []OHLCV
	for rows.Next() {
		var ohlcv OHLCV
		err := rows.Scan(
			&ohlcv.Symbol,
			&ohlcv.Timestamp,
			&ohlcv.Open,
			&ohlcv.High,
			&ohlcv.Low,
			&ohlcv.Close,
			&ohlcv.Volume,
			&ohlcv.Turnover,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, ohlcv)
	}

	return results, nil
}

// RefreshContinuousAggregates manually refreshes the materialized views
func (s *MarketDataService) RefreshContinuousAggregates(ctx context.Context) error {
	queries := []string{
		"CALL refresh_continuous_aggregate('ohlcv_daily', NULL, NULL);",
		"CALL refresh_continuous_aggregate('ohlcv_weekly', NULL, NULL);",
		"CALL refresh_continuous_aggregate('ohlcv_monthly', NULL, NULL);",
	}

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to refresh aggregate: %w", err)
		}
	}

	return nil
}

// Helper functions

func parseROCDate(rocDate string) (time.Time, error) {
	// ROC date format: "113/01/22" (ROC year 113 = AD 2024)
	// Convert to AD year and parse
	var year, month, day int
	_, err := fmt.Sscanf(rocDate, "%d/%d/%d", &year, &month, &day)
	if err != nil {
		return time.Time{}, err
	}

	adYear := year + 1911 // ROC year 0 = AD 1911
	return time.Date(adYear, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func parseDecimal(s string) (decimal.Decimal, error) {
	// Remove commas and parse
	cleaned := ""
	for _, r := range s {
		if r != ',' {
			cleaned += string(r)
		}
	}
	return decimal.NewFromString(cleaned)
}

func parseInt64(s string) (int64, error) {
	// Remove commas
	cleaned := ""
	for _, r := range s {
		if r != ',' {
			cleaned += string(r)
		}
	}
	
	var result int64
	_, err := fmt.Sscanf(cleaned, "%d", &result)
	return result, err
}
