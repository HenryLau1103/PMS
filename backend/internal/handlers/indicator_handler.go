package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"psm-backend/internal/services"
)

type IndicatorHandler struct {
	service *services.TechnicalAnalysisService
}

func NewIndicatorHandler(service *services.TechnicalAnalysisService) *IndicatorHandler {
	return &IndicatorHandler{service: service}
}

// GetMA calculates Moving Average
// GET /api/v1/indicators/:symbol/ma?period=20&type=SMA&limit=100
func (h *IndicatorHandler) GetMA(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	period := c.QueryInt("period", 20)
	maType := c.Query("type", "SMA") // SMA or EMA
	limit := c.QueryInt("limit", 100)

	if period < 2 || period > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period must be between 2 and 200",
		})
	}

	if maType != "SMA" && maType != "EMA" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "type must be SMA or EMA",
		})
	}

	ctx := context.Background()
	results, err := h.service.CalculateMA(ctx, symbol, period, maType, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to calculate MA",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"symbol":    symbol,
		"indicator": "MA",
		"params": fiber.Map{
			"period": period,
			"type":   maType,
		},
		"count": len(results),
		"data":  results,
	})
}

// GetRSI calculates RSI
// GET /api/v1/indicators/:symbol/rsi?period=14&limit=100
func (h *IndicatorHandler) GetRSI(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	period := c.QueryInt("period", 14)
	limit := c.QueryInt("limit", 100)

	if period < 2 || period > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period must be between 2 and 100",
		})
	}

	ctx := context.Background()
	results, err := h.service.CalculateRSI(ctx, symbol, period, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to calculate RSI",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"symbol":    symbol,
		"indicator": "RSI",
		"params": fiber.Map{
			"period": period,
		},
		"count": len(results),
		"data":  results,
	})
}

// GetMACD calculates MACD
// GET /api/v1/indicators/:symbol/macd?fast=12&slow=26&signal=9&limit=100
func (h *IndicatorHandler) GetMACD(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	fast := c.QueryInt("fast", 12)
	slow := c.QueryInt("slow", 26)
	signal := c.QueryInt("signal", 9)
	limit := c.QueryInt("limit", 100)

	ctx := context.Background()
	results, err := h.service.CalculateMACD(ctx, symbol, fast, slow, signal, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to calculate MACD",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"symbol":    symbol,
		"indicator": "MACD",
		"params": fiber.Map{
			"fast":   fast,
			"slow":   slow,
			"signal": signal,
		},
		"count": len(results),
		"data":  results,
	})
}

// GetBollingerBands calculates Bollinger Bands
// GET /api/v1/indicators/:symbol/bb?period=20&stddev=2&limit=100
func (h *IndicatorHandler) GetBollingerBands(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	period := c.QueryInt("period", 20)
	limit := c.QueryInt("limit", 100)
	
	stdDevStr := c.Query("stddev", "2.0")
	stdDev, err := strconv.ParseFloat(stdDevStr, 64)
	if err != nil {
		stdDev = 2.0
	}

	ctx := context.Background()
	results, err := h.service.CalculateBollingerBands(ctx, symbol, period, stdDev, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to calculate Bollinger Bands",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"symbol":    symbol,
		"indicator": "BB",
		"params": fiber.Map{
			"period": period,
			"stddev": stdDev,
		},
		"count": len(results),
		"data":  results,
	})
}

// GetKDJ calculates KDJ indicator
// GET /api/v1/indicators/:symbol/kdj?period=9&limit=100
func (h *IndicatorHandler) GetKDJ(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	period := c.QueryInt("period", 9)
	limit := c.QueryInt("limit", 100)

	if period < 2 || period > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "period must be between 2 and 100",
		})
	}

	ctx := context.Background()
	results, err := h.service.CalculateKDJ(ctx, symbol, period, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to calculate KDJ",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"symbol":    symbol,
		"indicator": "KDJ",
		"params": fiber.Map{
			"period": period,
		},
		"count": len(results),
		"data":  results,
	})
}

// GetBatchIndicators calculates multiple indicators at once
// POST /api/v1/indicators/:symbol/batch
// Body: {"indicators": ["MA", "RSI", "MACD"], "params": {...}}
func (h *IndicatorHandler) GetBatchIndicators(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	var req struct {
		Indicators []string               `json:"indicators"`
		Params     map[string]interface{} `json:"params"`
		Limit      int                    `json:"limit"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	ctx := context.Background()
	response := fiber.Map{
		"success": true,
		"symbol":  symbol,
		"data":    fiber.Map{},
	}

	// Calculate each requested indicator
	for _, indicator := range req.Indicators {
		switch indicator {
		case "MA":
			period := 20
			maType := "SMA"
			if p, ok := req.Params["ma_period"].(float64); ok {
				period = int(p)
			}
			if t, ok := req.Params["ma_type"].(string); ok {
				maType = t
			}
			if results, err := h.service.CalculateMA(ctx, symbol, period, maType, req.Limit); err == nil {
				response["data"].(fiber.Map)["MA"] = results
			}

		case "RSI":
			period := 14
			if p, ok := req.Params["rsi_period"].(float64); ok {
				period = int(p)
			}
			if results, err := h.service.CalculateRSI(ctx, symbol, period, req.Limit); err == nil {
				response["data"].(fiber.Map)["RSI"] = results
			}

		case "MACD":
			if results, err := h.service.CalculateMACD(ctx, symbol, 12, 26, 9, req.Limit); err == nil {
				response["data"].(fiber.Map)["MACD"] = results
			}

		case "BB":
			period := 20
			if p, ok := req.Params["bb_period"].(float64); ok {
				period = int(p)
			}
			if results, err := h.service.CalculateBollingerBands(ctx, symbol, period, 2.0, req.Limit); err == nil {
				response["data"].(fiber.Map)["BB"] = results
			}

		case "KDJ":
			period := 9
			if p, ok := req.Params["kdj_period"].(float64); ok {
				period = int(p)
			}
			if results, err := h.service.CalculateKDJ(ctx, symbol, period, req.Limit); err == nil {
				response["data"].(fiber.Map)["KDJ"] = results
			}
		}
	}

	return c.JSON(response)
}
