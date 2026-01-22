package handlers

import (
	"psm-backend/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// SentimentHandler handles sentiment analysis endpoints
type SentimentHandler struct {
	sentimentService *services.SentimentService
}

func NewSentimentHandler(sentimentService *services.SentimentService) *SentimentHandler {
	return &SentimentHandler{
		sentimentService: sentimentService,
	}
}

// GetSentimentSummary returns sentiment summary for a symbol
// GET /api/v1/sentiment/:symbol
func (h *SentimentHandler) GetSentimentSummary(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	// Get days parameter (default 7)
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	summary, err := h.sentimentService.GetSentimentSummary(c.Context(), symbol, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get sentiment summary: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

// AnalyzeUnanalyzedNews triggers batch analysis of unanalyzed news
// POST /api/v1/sentiment/analyze
func (h *SentimentHandler) AnalyzeUnanalyzedNews(c *fiber.Ctx) error {
	// Get limit parameter (default 100)
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	count, err := h.sentimentService.AnalyzeUnanalyzedNews(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to analyze news: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":        true,
		"analyzed_count": count,
		"message":        "情感分析完成",
	})
}

// AnalyzeSingleArticle analyzes a single article by ID
// POST /api/v1/sentiment/article/:id
func (h *SentimentHandler) AnalyzeSingleArticle(c *fiber.Ctx) error {
	articleID := c.Params("id")
	if articleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article ID is required",
		})
	}

	result, err := h.sentimentService.AnalyzeNewsArticle(c.Context(), articleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to analyze article: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// AnalyzeText performs sentiment analysis on provided text
// POST /api/v1/sentiment/text
func (h *SentimentHandler) AnalyzeText(c *fiber.Ctx) error {
	var req struct {
		Text string `json:"text"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "text is required",
		})
	}

	result := h.sentimentService.AnalyzeSentiment(req.Text)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
