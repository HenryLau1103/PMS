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
	service     *services.MarketDataService
	db          *database.DB
	syncStatus  *SyncStatus
	mu          sync.RWMutex
}

type SyncStatus struct {
	IsRunning       bool      `json:"is_running"`
	TotalSymbols    int       `json:"total_symbols"`
	ProcessedCount  int       `json:"processed_count"`
	SuccessCount    int       `json:"success_count"`
	FailedCount     int       `json:"failed_count"`
	CurrentSymbol   string    `json:"current_symbol"`
	StartedAt       time.Time `json:"started_at,omitempty"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	FailedSymbols   []string  `json:"failed_symbols,omitempty"`
}

func NewBulkSyncHandler(service *services.MarketDataService, db *database.DB) *BulkSyncHandler {
	return &BulkSyncHandler{
		service: service,
		db:      db,
		syncStatus: &SyncStatus{
			IsRunning: false,
		},
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

// StartBulkSync starts syncing all stocks or portfolio stocks
// POST /api/v1/market/bulk-sync/start
// Body: {"portfolio_id": "uuid", "start_date": "2024-01-01", "end_date": "2024-12-31", "priority_holdings": true}
func (h *BulkSyncHandler) StartBulkSync(c *fiber.Ctx) error {
	h.mu.Lock()
	if h.syncStatus.IsRunning {
		h.mu.Unlock()
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "sync is already running",
		})
	}

	// Reset status
	h.syncStatus = &SyncStatus{
		IsRunning:     true,
		StartedAt:     time.Now(),
		FailedSymbols: []string{},
	}
	h.mu.Unlock()

	var req struct {
		PortfolioID      string `json:"portfolio_id"`
		StartDate        string `json:"start_date"`
		EndDate          string `json:"end_date"`
		PriorityHoldings bool   `json:"priority_holdings"`
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

	// Start sync in background goroutine
	go h.runBulkSync(req.PortfolioID, startDate, endDate, req.PriorityHoldings)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "bulk sync started",
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

	h.syncStatus.IsRunning = false
	h.syncStatus.ErrorMessage = "stopped by user"
	h.syncStatus.CompletedAt = time.Now()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "sync stopped",
	})
}

func (h *BulkSyncHandler) runBulkSync(portfolioID string, startDate, endDate time.Time, priorityHoldings bool) {
	ctx := context.Background()

	// Get list of symbols to sync
	symbols, err := h.getSymbolsToSync(ctx, portfolioID, priorityHoldings)
	if err != nil {
		h.mu.Lock()
		h.syncStatus.IsRunning = false
		h.syncStatus.ErrorMessage = err.Error()
		h.syncStatus.CompletedAt = time.Now()
		h.mu.Unlock()
		return
	}

	h.mu.Lock()
	h.syncStatus.TotalSymbols = len(symbols)
	h.mu.Unlock()

	// Sync each symbol with rate limiting (avoid TWSE API rate limits)
	for _, symbol := range symbols {
		h.mu.RLock()
		if !h.syncStatus.IsRunning {
			h.mu.RUnlock()
			break
		}
		h.mu.RUnlock()

		h.mu.Lock()
		h.syncStatus.CurrentSymbol = symbol
		h.mu.Unlock()

		// Fetch and save data
		data, err := h.service.FetchDailyData(ctx, symbol, startDate, endDate)
		if err != nil {
			h.mu.Lock()
			h.syncStatus.FailedCount++
			h.syncStatus.FailedSymbols = append(h.syncStatus.FailedSymbols, symbol)
			h.mu.Unlock()
		} else if len(data) > 0 {
			if err := h.service.SaveOHLCV(ctx, data); err != nil {
				h.mu.Lock()
				h.syncStatus.FailedCount++
				h.syncStatus.FailedSymbols = append(h.syncStatus.FailedSymbols, symbol)
				h.mu.Unlock()
			} else {
				h.mu.Lock()
				h.syncStatus.SuccessCount++
				h.mu.Unlock()
			}
		}

		h.mu.Lock()
		h.syncStatus.ProcessedCount++
		h.mu.Unlock()

		// Rate limiting: 3 seconds between requests (TWSE allows ~20 req/min)
		time.Sleep(3 * time.Second)
	}

	// Refresh aggregates at the end
	h.service.RefreshContinuousAggregates(ctx)

	h.mu.Lock()
	h.syncStatus.IsRunning = false
	h.syncStatus.CompletedAt = time.Now()
	h.mu.Unlock()
}

func (h *BulkSyncHandler) getSymbolsToSync(ctx context.Context, portfolioID string, priorityHoldings bool) ([]string, error) {
	var symbols []string

	if priorityHoldings && portfolioID != "" {
		// Get holdings first
		query := `
			SELECT DISTINCT symbol 
			FROM positions_current 
			WHERE portfolio_id = $1 AND quantity > 0
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
			symbols = append(symbols, symbol)
		}
	}

	// Get all stocks
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
		// Skip if already in holdings list
		if !contains(symbols, symbol) {
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
