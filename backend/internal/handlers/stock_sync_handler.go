package handlers

import (
	"psm-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type StockSyncHandler struct {
	syncService *services.StockSyncService
}

func NewStockSyncHandler(syncService *services.StockSyncService) *StockSyncHandler {
	return &StockSyncHandler{
		syncService: syncService,
	}
}

// SyncStocks handles POST /api/v1/stocks/sync
func (h *StockSyncHandler) SyncStocks(c *fiber.Ctx) error {
	result, err := h.syncService.SyncAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Stock synchronization completed",
		"result":  result,
	})
}
