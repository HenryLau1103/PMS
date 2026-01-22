package handlers

import (
	"psm-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type StockHandler struct {
	stockService *services.StockService
}

func NewStockHandler(stockService *services.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
	}
}

// SearchStocks handles GET /api/v1/stocks/search?q=2330
func (h *StockHandler) SearchStocks(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "query parameter 'q' is required",
		})
	}

	limit := c.QueryInt("limit", 20)

	stocks, err := h.stockService.SearchStocks(c.Context(), query, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stocks)
}

// GetStock handles GET /api/v1/stocks/:symbol
func (h *StockHandler) GetStock(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol parameter is required",
		})
	}

	stock, err := h.stockService.GetStockBySymbol(c.Context(), symbol)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stock)
}
