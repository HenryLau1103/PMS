package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"psm-backend/internal/database"
	"psm-backend/internal/services"
)

type BulkSyncHandler struct {
	service         *services.MarketDataService
	bulkSyncService *services.BulkSyncService
	db              *database.DB
	syncStatus      *SyncStatus
	mu              sync.RWMutex
	stopChan        chan struct{}
}

type SyncStatus struct {
	IsRunning       bool      `json:"is_running"`
	Mode            string    `json:"mode"` // "date" (new) or "symbol" (legacy)
	TotalDays       int       `json:"total_days"`
	ProcessedDays   int       `json:"processed_days"`
	TotalSymbols    int       `json:"total_symbols"`
	ProcessedCount  int       `json:"processed_count"`
	SuccessCount    int       `json:"success_count"`
	FailedCount     int       `json:"failed_count"`
	SkippedCount    int       `json:"skipped_count"`
	CurrentDate     string    `json:"current_date"`
	CurrentSymbol   string    `json:"current_symbol"`
	StartedAt       time.Time `json:"started_at,omitempty"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	FailedDates     []string  `json:"failed_dates,omitempty"`
	FailedSymbols   []string  `json:"failed_symbols,omitempty"`
	EstimatedTime   string    `json:"estimated_time,omitempty"`
}

func NewBulkSyncHandler(service *services.MarketDataService, db *database.DB) *BulkSyncHandler {
	return &BulkSyncHandler{
		service:         service,
		bulkSyncService: services.NewBulkSyncService(db),
		db:              db,
		syncStatus: &SyncStatus{
			IsRunning: false,
		},
		stopChan: make(chan struct{}),
	}
}

// GetSyncStatus returns current sync status
// GET /api/v1/market/bulk-sync/status
func (h *BulkSyncHandler) GetSyncStatus(c *fiber.Ctx) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return c.JSON(fiber.Map{
		"success": true,
		"status":  h.syncStatus,
	})
}

// GetSyncInfo returns information about existing synced data
// GET /api/v1/market/bulk-sync/info
func (h *BulkSyncHandler) GetSyncInfo(c *fiber.Ctx) error {
	ctx := c.Context()

	firstDate, _ := h.bulkSyncService.GetFirstSyncedDate(ctx)
	lastDate, _ := h.bulkSyncService.GetLastSyncedDate(ctx)

	var firstDateStr, lastDateStr string
	if firstDate != nil {
		firstDateStr = firstDate.Format("2006-01-02")
	}
	if lastDate != nil {
		lastDateStr = lastDate.Format("2006-01-02")
	}

	// Count total synced days
	syncedCount := 0
	if firstDate != nil && lastDate != nil {
		syncedDates, _ := h.bulkSyncService.GetSyncedDates(ctx, *firstDate, *lastDate)
		syncedCount = len(syncedDates)
	}

	// Get gaps count
	gaps, _ := h.bulkSyncService.GetSyncGaps(ctx)
	gapsCount := len(gaps)

	return c.JSON(fiber.Map{
		"success": true,
		"info": fiber.Map{
			"first_synced_date": firstDateStr,
			"last_synced_date":  lastDateStr,
			"synced_days_count": syncedCount,
			"gaps_count":        gapsCount,
		},
	})
}

// StartBulkSync starts syncing all stocks using the new date-based approach
// POST /api/v1/market/bulk-sync/start
// Body: {"start_date": "2024-01-01", "end_date": "2026-01-22", "skip_synced": true}
func (h *BulkSyncHandler) StartBulkSync(c *fiber.Ctx) error {
	h.mu.Lock()
	if h.syncStatus.IsRunning {
		h.mu.Unlock()
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "sync is already running",
		})
	}

	// Reset stop channel
	h.stopChan = make(chan struct{})

	// Reset status
	h.syncStatus = &SyncStatus{
		IsRunning:   true,
		Mode:        "date",
		StartedAt:   time.Now(),
		FailedDates: []string{},
	}
	h.mu.Unlock()

	var req struct {
		PortfolioID      string `json:"portfolio_id"`
		StartDate        string `json:"start_date"`
		EndDate          string `json:"end_date"`
		PriorityHoldings bool   `json:"priority_holdings"`
		SkipSynced       bool   `json:"skip_synced"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.mu.Lock()
		h.syncStatus.IsRunning = false
		h.syncStatus.ErrorMessage = "invalid request body"
		h.mu.Unlock()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.mu.Lock()
		h.syncStatus.IsRunning = false
		h.syncStatus.ErrorMessage = "invalid start_date format"
		h.mu.Unlock()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid start_date format, use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.mu.Lock()
		h.syncStatus.IsRunning = false
		h.syncStatus.ErrorMessage = "invalid end_date format"
		h.mu.Unlock()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid end_date format, use YYYY-MM-DD",
		})
	}

	// Use the skip_synced value from request (defaults to true if not specified)
	skipSynced := req.SkipSynced

	// Start sync in background goroutine
	go h.runDateBasedBulkSync(startDate, endDate, skipSynced)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "bulk sync started (date-based mode)",
		"mode":    "date",
	})
}

// StopBulkSync stops the running sync
// POST /api/v1/market/bulk-sync/stop
func (h *BulkSyncHandler) StopBulkSync(c *fiber.Ctx) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.syncStatus.IsRunning {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no sync is running",
		})
	}

	// Signal stop
	close(h.stopChan)

	h.syncStatus.IsRunning = false
	h.syncStatus.ErrorMessage = "stopped by user"
	h.syncStatus.CompletedAt = time.Now()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "sync stopped",
	})
}

// runDateBasedBulkSync uses the new MI_INDEX API to fetch all stocks per date
// Much more efficient: ~500 API calls for 2 years vs ~45,000 calls
func (h *BulkSyncHandler) runDateBasedBulkSync(startDate, endDate time.Time, skipSynced bool) {
	ctx := context.Background()

	// Get all potential trading days
	allDays := h.bulkSyncService.GetTradingDays(startDate, endDate)

	// Get already synced dates if skipping
	syncedDates := make(map[string]bool)
	if skipSynced {
		var err error
		syncedDates, err = h.bulkSyncService.GetSyncedDates(ctx, startDate, endDate)
		if err != nil {
			h.mu.Lock()
			h.syncStatus.ErrorMessage = "failed to get synced dates: " + err.Error()
			h.mu.Unlock()
		}
	}

	// Filter out already synced dates
	var daysToSync []time.Time
	for _, day := range allDays {
		if !syncedDates[day.Format("2006-01-02")] {
			daysToSync = append(daysToSync, day)
		}
	}

	skippedCount := len(allDays) - len(daysToSync)

	h.mu.Lock()
	h.syncStatus.TotalDays = len(daysToSync)
	h.syncStatus.SkippedCount = skippedCount
	// Estimate time: 5 seconds per day
	estimatedMinutes := (len(daysToSync) * 5) / 60
	h.syncStatus.EstimatedTime = formatDuration(time.Duration(estimatedMinutes) * time.Minute)
	h.mu.Unlock()

	// Process each date with rate limiting
	// IMPORTANT: 5 seconds between requests to avoid being banned by TWSE
	rateLimitDelay := 5 * time.Second

	for i, day := range daysToSync {
		// Check for stop signal
		select {
		case <-h.stopChan:
			h.mu.Lock()
			h.syncStatus.IsRunning = false
			h.syncStatus.ErrorMessage = "stopped by user"
			h.syncStatus.CompletedAt = time.Now()
			h.mu.Unlock()
			return
		default:
		}

		h.mu.RLock()
		if !h.syncStatus.IsRunning {
			h.mu.RUnlock()
			break
		}
		h.mu.RUnlock()

		dateStr := day.Format("2006-01-02")

		h.mu.Lock()
		h.syncStatus.CurrentDate = dateStr
		h.syncStatus.ProcessedDays = i
		// Update estimated time remaining
		remainingDays := len(daysToSync) - i
		remainingSeconds := remainingDays * 5
		h.syncStatus.EstimatedTime = formatDuration(time.Duration(remainingSeconds) * time.Second)
		h.mu.Unlock()

		// Fetch all stocks for this date
		data, err := h.bulkSyncService.FetchAllStocksForDate(ctx, day)
		if err != nil {
			// Real error (network, parsing, etc.)
			h.mu.Lock()
			h.syncStatus.FailedCount++
			h.syncStatus.FailedDates = append(h.syncStatus.FailedDates, dateStr)
			h.mu.Unlock()
		} else if len(data) == 0 {
			// Non-trading day (holiday) - count as skipped, not failed
			h.mu.Lock()
			h.syncStatus.SkippedCount++
			h.mu.Unlock()
		} else {
			// Save all stock data for this date
			savedCount, err := h.bulkSyncService.SaveBulkOHLCV(ctx, data)
			if err != nil {
				h.mu.Lock()
				h.syncStatus.FailedCount++
				h.syncStatus.FailedDates = append(h.syncStatus.FailedDates, dateStr)
				h.mu.Unlock()
			} else {
				h.mu.Lock()
				h.syncStatus.SuccessCount++
				h.syncStatus.TotalSymbols += len(data)
				h.syncStatus.ProcessedCount += savedCount
				h.mu.Unlock()
			}
		}

		h.mu.Lock()
		h.syncStatus.ProcessedDays = i + 1
		h.mu.Unlock()

		// Rate limiting: 5 seconds between requests to avoid being banned
		// This is conservative - TWSE may allow faster but being safe is better
		time.Sleep(rateLimitDelay)
	}

	// Refresh aggregates at the end
	h.service.RefreshContinuousAggregates(ctx)

	h.mu.Lock()
	h.syncStatus.IsRunning = false
	h.syncStatus.CompletedAt = time.Now()
	h.syncStatus.EstimatedTime = "completed"
	h.mu.Unlock()
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1 min"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return formatPlural(hours, "hour") + " " + formatPlural(minutes, "min")
	}
	return formatPlural(minutes, "min")
}

func formatPlural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return formatInt(n) + " " + unit + "s"
}

func formatInt(n int) string {
	return stringFromInt(n)
}

func stringFromInt(n int) string {
	if n < 0 {
		return "-" + stringFromInt(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return stringFromInt(n/10) + string(rune('0'+n%10))
}

// Legacy functions kept for backward compatibility

func (h *BulkSyncHandler) getSymbolsToSync(ctx context.Context, portfolioID string, priorityHoldings bool, startDate, endDate time.Time) ([]string, error) {
	var symbols []string

	// Get symbols that already have data in the date range (to skip them)
	syncedSymbols := make(map[string]bool)
	skipQuery := `
		SELECT DISTINCT symbol 
		FROM stock_ohlcv 
		WHERE timestamp >= $1 AND timestamp <= $2
	`
	skipRows, err := h.db.Query(skipQuery, startDate, endDate)
	if err == nil {
		defer skipRows.Close()
		for skipRows.Next() {
			var symbol string
			if err := skipRows.Scan(&symbol); err == nil {
				syncedSymbols[symbol] = true
			}
		}
	}

	if priorityHoldings && portfolioID != "" {
		// Get holdings first
		query := `
			SELECT DISTINCT symbol 
			FROM positions_current 
			WHERE portfolio_id = $1 AND total_quantity > 0
			ORDER BY symbol
		`
		rows, err := h.db.Query(query, portfolioID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var symbol string
			if err := rows.Scan(&symbol); err != nil {
				continue
			}
			// Only add if not already synced
			if !syncedSymbols[symbol] {
				symbols = append(symbols, symbol)
			}
		}
	}

	// Get all stocks that haven't been synced yet
	query := `SELECT symbol FROM taiwan_stocks ORDER BY symbol`
	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allSymbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			continue
		}
		// Skip if already synced or in holdings list
		if !syncedSymbols[symbol] && !contains(symbols, symbol) {
			allSymbols = append(allSymbols, symbol)
		}
	}

	// Holdings first, then all other stocks
	return append(symbols, allSymbols...), nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
