package handlers

import (
	"psm-backend/internal/models"
	"psm-backend/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LedgerHandler struct {
	ledgerService *services.LedgerService
}

func NewLedgerHandler(ledgerService *services.LedgerService) *LedgerHandler {
	return &LedgerHandler{
		ledgerService: ledgerService,
	}
}

// CreateEvent handles POST /api/v1/events
func (h *LedgerHandler) CreateEvent(c *fiber.Ctx) error {
	var req models.CreateLedgerEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// For demo, use hardcoded user ID
	// In production, extract from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	event, err := h.ledgerService.CreateEvent(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(event)
}

// GetEvents handles GET /api/v1/portfolios/:portfolio_id/events
func (h *LedgerHandler) GetEvents(c *fiber.Ctx) error {
	portfolioID, err := uuid.Parse(c.Params("portfolio_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	events, err := h.ledgerService.GetEvents(c.Context(), portfolioID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(events)
}

// GetEventsBySymbol handles GET /api/v1/portfolios/:portfolio_id/events/:symbol
func (h *LedgerHandler) GetEventsBySymbol(c *fiber.Ctx) error {
	portfolioID, err := uuid.Parse(c.Params("portfolio_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	symbol := c.Params("symbol")

	events, err := h.ledgerService.GetEventsBySymbol(c.Context(), portfolioID, symbol)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(events)
}

// GetPositions handles GET /api/v1/portfolios/:portfolio_id/positions
func (h *LedgerHandler) GetPositions(c *fiber.Ctx) error {
	portfolioID, err := uuid.Parse(c.Params("portfolio_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	positions, err := h.ledgerService.GetPositions(c.Context(), portfolioID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(positions)
}

// GetPosition handles GET /api/v1/portfolios/:portfolio_id/positions/:symbol
func (h *LedgerHandler) GetPosition(c *fiber.Ctx) error {
	portfolioID, err := uuid.Parse(c.Params("portfolio_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	symbol := c.Params("symbol")

	position, err := h.ledgerService.GetPosition(c.Context(), portfolioID, symbol)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(position)
}

// CalculateUnrealizedPnL handles GET /api/v1/portfolios/:portfolio_id/positions/:symbol/pnl
func (h *LedgerHandler) CalculateUnrealizedPnL(c *fiber.Ctx) error {
	portfolioID, err := uuid.Parse(c.Params("portfolio_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	symbol := c.Params("symbol")

	currentPriceStr := c.Query("current_price")
	if currentPriceStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "current_price query parameter is required",
		})
	}

	currentPrice, err := decimal.NewFromString(currentPriceStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid current_price format",
		})
	}

	pnl, err := h.ledgerService.CalculateUnrealizedPnL(c.Context(), portfolioID, symbol, currentPrice)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(pnl)
}

// GetPortfolio handles GET /api/v1/portfolios/:portfolio_id
func (h *LedgerHandler) GetPortfolio(c *fiber.Ctx) error {
	portfolioID, err := uuid.Parse(c.Params("portfolio_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	portfolio, err := h.ledgerService.GetPortfolio(c.Context(), portfolioID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(portfolio)
}

// GetUserPortfolios handles GET /api/v1/users/:user_id/portfolios
func (h *LedgerHandler) GetUserPortfolios(c *fiber.Ctx) error {
	// For demo, use hardcoded user ID
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	portfolios, err := h.ledgerService.GetUserPortfolios(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(portfolios)
}
