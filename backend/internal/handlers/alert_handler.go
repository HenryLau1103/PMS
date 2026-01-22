package handlers

import (
	"psm-backend/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// AlertHandler handles alert endpoints
type AlertHandler struct {
	alertService *services.AlertService
}

func NewAlertHandler(alertService *services.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// GetAlerts retrieves alerts
// GET /api/v1/alerts
func (h *AlertHandler) GetAlerts(c *fiber.Ctx) error {
	symbol := c.Query("symbol", "")
	unacknowledgedOnly := c.Query("unacknowledged", "false") == "true"
	
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	alerts, err := h.alertService.GetAlerts(c.Context(), symbol, unacknowledgedOnly, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "查詢警報失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(alerts),
		"data":    alerts,
	})
}

// GetAlertsBySymbol retrieves alerts for a specific symbol
// GET /api/v1/alerts/:symbol
func (h *AlertHandler) GetAlertsBySymbol(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	unacknowledgedOnly := c.Query("unacknowledged", "false") == "true"
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	alerts, err := h.alertService.GetAlerts(c.Context(), symbol, unacknowledgedOnly, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "查詢警報失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(alerts),
		"data":    alerts,
	})
}

// AcknowledgeAlert marks an alert as acknowledged
// POST /api/v1/alerts/:id/ack
func (h *AlertHandler) AcknowledgeAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	if alertID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "alert ID is required",
		})
	}

	if err := h.alertService.AcknowledgeAlert(c.Context(), alertID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "確認警報失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "警報已確認",
	})
}

// GetAlertStats returns alert statistics
// GET /api/v1/alerts/stats
func (h *AlertHandler) GetAlertStats(c *fiber.Ctx) error {
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	stats, err := h.alertService.GetAlertStats(c.Context(), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "查詢統計失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// DetectVolumeSpike detects volume spike for a symbol
// GET /api/v1/alerts/:symbol/volume
func (h *AlertHandler) DetectVolumeSpike(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	threshold := 2.0
	if threshStr := c.Query("threshold"); threshStr != "" {
		if t, err := strconv.ParseFloat(threshStr, 64); err == nil && t > 0 {
			threshold = t
		}
	}

	analysis, err := h.alertService.DetectVolumeSpike(c.Context(), symbol, threshold)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "分析失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    analysis,
	})
}

// DetectPriceBreakout detects price breakout for a symbol
// GET /api/v1/alerts/:symbol/price
func (h *AlertHandler) DetectPriceBreakout(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	analysis, err := h.alertService.DetectPriceBreakout(c.Context(), symbol)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "分析失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    analysis,
	})
}

// ScanAll scans all symbols for anomalies
// POST /api/v1/alerts/scan
func (h *AlertHandler) ScanAll(c *fiber.Ctx) error {
	threshold := 2.0
	if threshStr := c.Query("threshold"); threshStr != "" {
		if t, err := strconv.ParseFloat(threshStr, 64); err == nil && t > 0 {
			threshold = t
		}
	}

	result, err := h.alertService.ScanAllSymbols(c.Context(), threshold)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "掃描失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
