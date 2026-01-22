package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"psm-backend/internal/services"
)

type NewsHandler struct {
	newsService *services.NewsService
}

func NewNewsHandler(newsService *services.NewsService) *NewsHandler {
	return &NewsHandler{newsService: newsService}
}

// GetNews retrieves news for a specific symbol
// GET /api/v1/news/:symbol
func (h *NewsHandler) GetNews(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	articles, err := h.newsService.GetNewsForSymbol(c.Context(), symbol, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"symbol":  symbol,
		"count":   len(articles),
		"news":    articles,
	})
}

// GetRecentNews retrieves recent news across all symbols
// GET /api/v1/news
func (h *NewsHandler) GetRecentNews(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "30"))

	articles, err := h.newsService.GetRecentNews(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(articles),
		"news":    articles,
	})
}

// FetchNews fetches new articles from Cnyes for a specific symbol
// POST /api/v1/news/:symbol/fetch
func (h *NewsHandler) FetchNews(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	articles, err := h.newsService.FetchNewsForSymbol(c.Context(), symbol, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"fetched": 0,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"symbol":  symbol,
		"fetched": len(articles),
		"news":    articles,
	})
}

// FetchGeneralNews fetches general Taiwan stock market news
// POST /api/v1/news/fetch
func (h *NewsHandler) FetchGeneralNews(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "30"))

	articles, err := h.newsService.FetchGeneralNews(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"fetched": 0,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"fetched": len(articles),
		"news":    articles,
	})
}
