package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"psm-backend/internal/database"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// BulkSyncService handles efficient bulk sync using MI_INDEX API
// This API returns ALL stocks data for a single date in one request
// Much more efficient than fetching each stock individually
type BulkSyncService struct {
	db *database.DB
}

func NewBulkSyncService(db *database.DB) *BulkSyncService {
	return &BulkSyncService{db: db}
}

// MIIndexResponse represents the TWSE MI_INDEX API response
type MIIndexResponse struct {
	Stat   string         `json:"stat"`
	Date   string         `json:"date"`
	Tables []MIIndexTable `json:"tables"`
}

// MIIndexTable represents a table in MI_INDEX response
type MIIndexTable struct {
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

// DailyStockData represents parsed daily stock data
type DailyStockData struct {
	Symbol    string
	Date      time.Time
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    int64
	Turnover  decimal.Decimal
}

// FetchAllStocksForDate fetches all stock data for a single date using MI_INDEX API
// This is much more efficient than fetching each stock individually
// Rate limit: Use at least 5 seconds between requests to avoid being banned
func (s *BulkSyncService) FetchAllStocksForDate(ctx context.Context, date time.Time) ([]DailyStockData, error) {
	dateStr := date.Format("20060102")
	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/afterTrading/MI_INDEX?response=json&date=%s&type=ALL", dateStr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-TW,zh;q=0.9,en;q=0.8")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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

	var result MIIndexResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if result.Stat != "OK" {
		// This typically means a non-trading day (holiday, weekend that got through)
		// Return empty data instead of error - this is expected behavior
		return []DailyStockData{}, nil
	}

	// Find the stock price table (Table 8 - 每日收盤行情)
	// Title contains "每日收盤行情" or is the table with stock data
	var stockTable *MIIndexTable
	for i := range result.Tables {
		// Table 8 typically has the most rows and contains stock data
		if len(result.Tables[i].Data) > 1000 {
			stockTable = &result.Tables[i]
			break
		}
	}

	if stockTable == nil {
		return nil, fmt.Errorf("stock data table not found in response")
	}

	// Parse date from response (format: 20241220)
	parsedDate, err := time.Parse("20060102", result.Date)
	if err != nil {
		parsedDate = date // fallback to requested date
	}

	// Parse stock data
	// Fields: 證券代號, 證券名稱, 成交股數, 成交筆數, 成交金額, 開盤價, 最高價, 最低價, 收盤價, ...
	var stocks []DailyStockData
	for _, row := range stockTable.Data {
		if len(row) < 9 {
			continue
		}

		symbol := strings.TrimSpace(row[0])
		if symbol == "" {
			continue
		}

		// Skip ETFs and special securities for now (only regular stocks with 4-digit codes)
		// Actually, include all - they're all valid securities
		
		// Parse prices (columns 5-8: 開盤價, 最高價, 最低價, 收盤價)
		openPrice, err := parseDecimalFromTWSE(row[5])
		if err != nil {
			continue // Skip if no valid price
		}
		highPrice, err := parseDecimalFromTWSE(row[6])
		if err != nil {
			continue
		}
		lowPrice, err := parseDecimalFromTWSE(row[7])
		if err != nil {
			continue
		}
		closePrice, err := parseDecimalFromTWSE(row[8])
		if err != nil {
			continue
		}

		// Skip if no trading (all prices are 0)
		if openPrice.IsZero() && closePrice.IsZero() {
			continue
		}

		// Parse volume (column 2: 成交股數)
		volume := parseVolumeFromTWSE(row[2])

		// Parse turnover (column 4: 成交金額)
		turnover, _ := parseDecimalFromTWSE(row[4])

		stocks = append(stocks, DailyStockData{
			Symbol:   symbol,
			Date:     parsedDate,
			Open:     openPrice,
			High:     highPrice,
			Low:      lowPrice,
			Close:    closePrice,
			Volume:   volume,
			Turnover: turnover,
		})
	}

	return stocks, nil
}

// SaveBulkOHLCV saves multiple stock data in a single transaction
func (s *BulkSyncService) SaveBulkOHLCV(ctx context.Context, data []DailyStockData) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
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
		return 0, err
	}
	defer stmt.Close()

	savedCount := 0
	for _, d := range data {
		_, err := stmt.ExecContext(ctx,
			d.Symbol,
			d.Date,
			d.Open,
			d.High,
			d.Low,
			d.Close,
			d.Volume,
			d.Turnover,
		)
		if err != nil {
			// Log error but continue with other records
			continue
		}
		savedCount++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return savedCount, nil
}

// GetTradingDays returns a list of potential trading days between start and end date
// Excludes weekends, but holidays need to be handled by API response
func (s *BulkSyncService) GetTradingDays(startDate, endDate time.Time) []time.Time {
	var days []time.Time
	current := startDate

	for !current.After(endDate) {
		// Skip weekends
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			days = append(days, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return days
}

// GetSyncedDates returns dates that already have COMPLETE data for the given date range
// A date is considered "synced" if it has more than 1000 records (full day sync)
// This prevents skipping dates that only have partial data from old sync method
func (s *BulkSyncService) GetSyncedDates(ctx context.Context, startDate, endDate time.Time) (map[string]bool, error) {
	syncedDates := make(map[string]bool)

	// Only consider a date as "synced" if it has more than 1000 records
	// (Full day sync typically has 14,000+ records)
	query := `
		SELECT DATE(timestamp) as sync_date, COUNT(*) as cnt
		FROM stock_ohlcv
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY DATE(timestamp)
		HAVING COUNT(*) > 1000
	`

	rows, err := s.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return syncedDates, err
	}
	defer rows.Close()

	for rows.Next() {
		var syncDate time.Time
		var cnt int
		if err := rows.Scan(&syncDate, &cnt); err == nil {
			syncedDates[syncDate.Format("2006-01-02")] = true
		}
	}

	return syncedDates, nil
}

// GetLastSyncedDate returns the most recent date that has complete data (>1000 records)
// Returns nil if no complete sync exists
func (s *BulkSyncService) GetLastSyncedDate(ctx context.Context) (*time.Time, error) {
	query := `
		SELECT DATE(timestamp) as sync_date
		FROM stock_ohlcv
		GROUP BY DATE(timestamp)
		HAVING COUNT(*) > 1000
		ORDER BY DATE(timestamp) DESC
		LIMIT 1
	`

	var lastDate time.Time
	err := s.db.QueryRowContext(ctx, query).Scan(&lastDate)
	if err != nil {
		return nil, nil // No complete sync exists
	}

	return &lastDate, nil
}

// GetFirstSyncedDate returns the earliest date that has complete data (>1000 records)
func (s *BulkSyncService) GetFirstSyncedDate(ctx context.Context) (*time.Time, error) {
	query := `
		SELECT DATE(timestamp) as sync_date
		FROM stock_ohlcv
		GROUP BY DATE(timestamp)
		HAVING COUNT(*) > 1000
		ORDER BY DATE(timestamp) ASC
		LIMIT 1
	`

	var firstDate time.Time
	err := s.db.QueryRowContext(ctx, query).Scan(&firstDate)
	if err != nil {
		return nil, nil // No complete sync exists
	}

	return &firstDate, nil
}

// GetSyncGaps returns date ranges that are missing complete data between first and last synced dates
func (s *BulkSyncService) GetSyncGaps(ctx context.Context) ([]time.Time, error) {
	firstDate, err := s.GetFirstSyncedDate(ctx)
	if err != nil || firstDate == nil {
		return nil, nil
	}

	lastDate, err := s.GetLastSyncedDate(ctx)
	if err != nil || lastDate == nil {
		return nil, nil
	}

	syncedDates, err := s.GetSyncedDates(ctx, *firstDate, *lastDate)
	if err != nil {
		return nil, err
	}

	// Find gaps (trading days without complete data)
	var gaps []time.Time
	allDays := s.GetTradingDays(*firstDate, *lastDate)
	for _, day := range allDays {
		if !syncedDates[day.Format("2006-01-02")] {
			gaps = append(gaps, day)
		}
	}

	return gaps, nil
}

// Helper functions

func parseDecimalFromTWSE(s string) (decimal.Decimal, error) {
	// Remove commas and trim spaces
	cleaned := strings.TrimSpace(s)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	
	if cleaned == "" || cleaned == "--" || cleaned == "---" {
		return decimal.Zero, fmt.Errorf("empty value")
	}

	return decimal.NewFromString(cleaned)
}

func parseVolumeFromTWSE(s string) int64 {
	cleaned := strings.TrimSpace(s)
	cleaned = strings.ReplaceAll(cleaned, ",", "")

	if cleaned == "" {
		return 0
	}

	var volume int64
	fmt.Sscanf(cleaned, "%d", &volume)
	return volume
}
