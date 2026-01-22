package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"psm-backend/internal/services"
)

type MarketDataHandler struct {
	service *services.MarketDataService
}

func NewMarketDataHandler(service *services.MarketDataService) *MarketDataHandler {
	return &MarketDataHandler{service: service}
}

// GetOHLCVRequest represents query parameters for OHLCV data
type GetOHLCVRequest struct {
	Symbol    string `query:"symbol"`
	StartDate string `query:"from"`
	EndDate   string `query:"to"`
	Limit     int    `query:"limit"`
}

// SyncMarketDataRequest represents request body for syncing market data
type SyncMarketDataRequest struct {
	Symbol    string `json:"symbol"`
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`   // YYYY-MM-DD
}

// GetOHLCV returns OHLCV data for a symbol
// GET /api/v1/stocks/:symbol/ohlcv?from=2024-01-01&to=2024-12-31&limit=100
func (h *MarketDataHandler) GetOHLCV(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	// Parse query parameters
	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")
	limit := c.QueryInt("limit", 100)

	var startDate, endDate time.Time
	var err error

	if fromStr != "" {
		startDate, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid from date format, use YYYY-MM-DD",
			})
		}
	} else {
		// Default to 1 year ago
		startDate = time.Now().AddDate(-1, 0, 0)
	}

	if toStr != "" {
		endDate, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid to date format, use YYYY-MM-DD",
			})
		}
	} else {
		// Default to today
		endDate = time.Now()
	}

	// Validate limit
	if limit <= 0 || limit > 10000 {
		limit = 100
	}

	ctx := context.Background()
	data, err := h.service.GetOHLCV(ctx, symbol, startDate, endDate, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch OHLCV data",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"symbol":  symbol,
		"from":    startDate.Format("2006-01-02"),
		"to":      endDate.Format("2006-01-02"),
		"count":   len(data),
		"data":    data,
	})
}

// SyncMarketData fetches historical data from TWSE and saves to database
// POST /api/v1/market/sync
// Body: {"symbol": "2330", "start_date": "2023-01-01", "end_date": "2024-12-31"}
func (h *MarketDataHandler) SyncMarketData(c *fiber.Ctx) error {
	var req SyncMarketDataRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate symbol
	if req.Symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid start_date format, use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid end_date format, use YYYY-MM-DD",
		})
	}

	// Validate date range
	if startDate.After(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "start_date must be before end_date",
		})
	}

	ctx := context.Background()

	// Fetch data from TWSE API
	data, err := h.service.FetchDailyData(ctx, req.Symbol, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to fetch data from TWSE",
			"details": err.Error(),
		})
	}

	if len(data) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no data found for the specified date range",
		})
	}

	// Save to database
	if err := h.service.SaveOHLCV(ctx, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to save data to database",
			"details": err.Error(),
		})
	}

	// Refresh continuous aggregates
	if err := h.service.RefreshContinuousAggregates(ctx); err != nil {
		// Log error but don't fail the request
		// Aggregates will be refreshed automatically by policy
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "market data synced successfully",
		"result": fiber.Map{
			"symbol":     req.Symbol,
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
			"records":    len(data),
		},
	})
}

// RefreshAggregates manually triggers continuous aggregate refresh
// POST /api/v1/market/refresh-aggregates
func (h *MarketDataHandler) RefreshAggregates(c *fiber.Ctx) error {
	ctx := context.Background()

	if err := h.service.RefreshContinuousAggregates(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to refresh aggregates",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "continuous aggregates refreshed successfully",
	})
}
