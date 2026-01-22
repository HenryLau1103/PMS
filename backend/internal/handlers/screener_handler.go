package handlers

import (
	"psm-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ScreenerHandler handles stock screening endpoints
type ScreenerHandler struct {
	screenerService *services.ScreenerService
}

func NewScreenerHandler(screenerService *services.ScreenerService) *ScreenerHandler {
	return &ScreenerHandler{
		screenerService: screenerService,
	}
}

// GetPresets returns available screening presets
// GET /api/v1/screener/presets
func (h *ScreenerHandler) GetPresets(c *fiber.Ctx) error {
	presets := h.screenerService.GetPresets()
	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(presets),
		"data":    presets,
	})
}

// RunPreset runs a preset screening
// GET /api/v1/screener/preset/:name
func (h *ScreenerHandler) RunPreset(c *fiber.Ctx) error {
	presetName := c.Params("name")
	if presetName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "preset name is required",
		})
	}

	results, err := h.screenerService.RunPreset(c.Context(), presetName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"preset":  presetName,
		"count":   len(results),
		"data":    results,
	})
}

// ScreenStocks screens stocks with custom criteria
// POST /api/v1/screener/screen
func (h *ScreenerHandler) ScreenStocks(c *fiber.Ctx) error {
	var criteria services.ScreenerCriteria
	if err := c.BodyParser(&criteria); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Set defaults
	if criteria.Limit <= 0 {
		criteria.Limit = 50
	}
	if criteria.SortBy == "" {
		criteria.SortBy = "score"
	}

	results, err := h.screenerService.ScreenStocks(c.Context(), &criteria)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "篩選失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"criteria": criteria,
		"count":    len(results),
		"data":     results,
	})
}

// QuickScreen provides quick screening shortcuts
// GET /api/v1/screener/quick/:type
func (h *ScreenerHandler) QuickScreen(c *fiber.Ctx) error {
	screenType := c.Params("type")
	
	var criteria services.ScreenerCriteria
	criteria.Limit = 20
	criteria.SortDesc = true
	criteria.MinPrice = 10

	switch screenType {
	case "gainers":
		// 今日漲幅前20
		criteria.MinChangePercent = 1
		criteria.SortBy = "change_percent"
	case "losers":
		// 今日跌幅前20
		criteria.MaxChangePercent = -1
		criteria.SortBy = "change_percent"
		criteria.SortDesc = false
	case "volume":
		// 成交量異常
		criteria.MinVolumeRatio = 2.0
		criteria.SortBy = "volume_ratio"
	case "momentum":
		// 動能股
		criteria.AboveMA20 = true
		criteria.MinChangePercent = 0
		criteria.SortBy = "score"
	case "breakout":
		// 突破股
		criteria.Near52WeekHigh = true
		criteria.SortBy = "change_percent"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid screen type",
			"valid":   []string{"gainers", "losers", "volume", "momentum", "breakout"},
		})
	}

	results, err := h.screenerService.ScreenStocks(c.Context(), &criteria)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "篩選失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"type":    screenType,
		"count":   len(results),
		"data":    results,
	})
}
