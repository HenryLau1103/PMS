package handlers

import (
	"psm-backend/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// AIHandler handles AI analysis endpoints
type AIHandler struct {
	aiService *services.AIService
}

func NewAIHandler(aiService *services.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// GetAnalysis returns AI analysis for a symbol
// GET /api/v1/ai/:symbol/analysis
func (h *AIHandler) GetAnalysis(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	// Check if API key is configured
	if !h.aiService.HasAPIKey() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "AI service not configured",
			"message": "請設定 GEMINI_API_KEY 環境變數以啟用 AI 分析功能",
		})
	}

	// Get analysis type (default: daily_summary)
	analysisTypeStr := c.Query("type", "daily_summary")
	var analysisType services.AnalysisType
	switch analysisTypeStr {
	case "daily_summary":
		analysisType = services.AnalysisTypeDailySummary
	case "investment_advice":
		analysisType = services.AnalysisTypeInvestmentAdvice
	case "risk_assessment":
		analysisType = services.AnalysisTypeRiskAssessment
	case "news_digest":
		analysisType = services.AnalysisTypeNewsDigest
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid analysis type",
			"valid":   []string{"daily_summary", "investment_advice", "risk_assessment", "news_digest"},
		})
	}

	result, err := h.aiService.GetAnalysis(c.Context(), symbol, analysisType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "分析失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetDailySummary returns daily summary for a symbol
// GET /api/v1/ai/:symbol/daily
func (h *AIHandler) GetDailySummary(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	if !h.aiService.HasAPIKey() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "AI service not configured",
			"message": "請設定 GEMINI_API_KEY 環境變數以啟用 AI 分析功能",
		})
	}

	result, err := h.aiService.GetAnalysis(c.Context(), symbol, services.AnalysisTypeDailySummary)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "分析失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetInvestmentAdvice returns investment advice for a symbol
// GET /api/v1/ai/:symbol/advice
func (h *AIHandler) GetInvestmentAdvice(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	if !h.aiService.HasAPIKey() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "AI service not configured",
			"message": "請設定 GEMINI_API_KEY 環境變數以啟用 AI 分析功能",
		})
	}

	result, err := h.aiService.GetAnalysis(c.Context(), symbol, services.AnalysisTypeInvestmentAdvice)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "分析失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetCachedAnalyses returns all cached analyses for a symbol
// GET /api/v1/ai/:symbol/history
func (h *AIHandler) GetCachedAnalyses(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.aiService.GetCachedAnalyses(c.Context(), symbol, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "查詢失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(results),
		"data":    results,
	})
}

// ClearCache clears cached analyses for a symbol
// DELETE /api/v1/ai/:symbol/cache
func (h *AIHandler) ClearCache(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol is required",
		})
	}

	if err := h.aiService.ClearCache(c.Context(), symbol); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "清除快取失敗: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "快取已清除",
	})
}

// GetStatus returns AI service status
// GET /api/v1/ai/status
func (h *AIHandler) GetStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success":    true,
		"configured": h.aiService.HasAPIKey(),
		"message": func() string {
			if h.aiService.HasAPIKey() {
				return "AI 服務已啟用"
			}
			return "AI 服務未設定。請設定 GEMINI_API_KEY 環境變數。"
		}(),
	})
}
