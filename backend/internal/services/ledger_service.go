package services

import (
	"context"
	"database/sql"
	"fmt"
	"psm-backend/internal/database"
	"psm-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LedgerService struct {
	db *database.DB
}

func NewLedgerService(db *database.DB) *LedgerService {
	return &LedgerService{db: db}
}

// CreateEvent creates a new ledger event (transaction)
func (s *LedgerService) CreateEvent(ctx context.Context, userID uuid.UUID, req models.CreateLedgerEventRequest) (*models.LedgerEvent, error) {
	// Calculate total amount
	totalAmount := req.Quantity.Mul(req.Price)
	
	// For buy transactions, add fees and taxes
	if req.EventType == models.EventTypeBuy {
		totalAmount = totalAmount.Add(req.Fee).Add(req.Tax)
	} else if req.EventType == models.EventTypeSell {
		// For sell transactions, subtract fees and taxes
		totalAmount = totalAmount.Sub(req.Fee).Sub(req.Tax)
	}

	event := &models.LedgerEvent{
		EventID:     uuid.New(),
		UserID:      userID,
		PortfolioID: req.PortfolioID,
		EventType:   req.EventType,
		Symbol:      req.Symbol,
		Quantity:    req.Quantity,
		Price:       req.Price,
		Fee:         req.Fee,
		Tax:         req.Tax,
		TotalAmount: totalAmount,
		OccurredAt:  req.OccurredAt,
		RecordedAt:  time.Now(),
		Source:      "manual",
		Notes:       req.Notes,
	}

	query := `
		INSERT INTO ledger_events (
			event_id, user_id, portfolio_id, event_type, symbol,
			quantity, price, fee, tax, total_amount,
			occurred_at, recorded_at, source, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		RETURNING event_id, recorded_at
	`

	err := s.db.QueryRowContext(ctx, query,
		event.EventID, event.UserID, event.PortfolioID, event.EventType, event.Symbol,
		event.Quantity, event.Price, event.Fee, event.Tax, event.TotalAmount,
		event.OccurredAt, event.RecordedAt, event.Source, event.Notes,
	).Scan(&event.EventID, &event.RecordedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create ledger event: %w", err)
	}

	// Refresh positions materialized view
	if err := s.RefreshPositions(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh positions: %w", err)
	}

	return event, nil
}

// GetEvents retrieves ledger events for a portfolio
func (s *LedgerService) GetEvents(ctx context.Context, portfolioID uuid.UUID, limit int) ([]models.LedgerEvent, error) {
	query := `
		SELECT 
			event_id, user_id, portfolio_id, event_type, symbol,
			quantity, price, fee, tax, total_amount,
			occurred_at, recorded_at, source, notes, payload
		FROM ledger_events
		WHERE portfolio_id = $1
		ORDER BY occurred_at DESC, recorded_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, portfolioID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	// 初始化為空數組而不是nil
	events := make([]models.LedgerEvent, 0)
	for rows.Next() {
		var event models.LedgerEvent
		err := rows.Scan(
			&event.EventID, &event.UserID, &event.PortfolioID, &event.EventType, &event.Symbol,
			&event.Quantity, &event.Price, &event.Fee, &event.Tax, &event.TotalAmount,
			&event.OccurredAt, &event.RecordedAt, &event.Source, &event.Notes, &event.Payload,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// GetEventsBySymbol retrieves ledger events for a specific symbol
func (s *LedgerService) GetEventsBySymbol(ctx context.Context, portfolioID uuid.UUID, symbol string) ([]models.LedgerEvent, error) {
	query := `
		SELECT 
			event_id, user_id, portfolio_id, event_type, symbol,
			quantity, price, fee, tax, total_amount,
			occurred_at, recorded_at, source, notes, payload
		FROM ledger_events
		WHERE portfolio_id = $1 AND symbol = $2
		ORDER BY occurred_at DESC, recorded_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, portfolioID, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by symbol: %w", err)
	}
	defer rows.Close()

	// 初始化為空數組而不是nil
	events := make([]models.LedgerEvent, 0)
	for rows.Next() {
		var event models.LedgerEvent
		err := rows.Scan(
			&event.EventID, &event.UserID, &event.PortfolioID, &event.EventType, &event.Symbol,
			&event.Quantity, &event.Price, &event.Fee, &event.Tax, &event.TotalAmount,
			&event.OccurredAt, &event.RecordedAt, &event.Source, &event.Notes, &event.Payload,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// RefreshPositions refreshes the materialized view for positions
func (s *LedgerService) RefreshPositions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "SELECT refresh_positions()")
	if err != nil {
		return fmt.Errorf("failed to refresh positions: %w", err)
	}
	return nil
}

// GetPositions retrieves current positions for a portfolio
func (s *LedgerService) GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]models.Position, error) {
	query := `
		SELECT 
			portfolio_id, symbol, total_quantity, total_cost,
			avg_cost_per_share, last_updated
		FROM positions_current
		WHERE portfolio_id = $1
		ORDER BY symbol ASC
	`

	rows, err := s.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	// 初始化為空數組而不是nil
	positions := make([]models.Position, 0)
	for rows.Next() {
		var pos models.Position
		err := rows.Scan(
			&pos.PortfolioID, &pos.Symbol, &pos.TotalQuantity, &pos.TotalCost,
			&pos.AvgCostPerShare, &pos.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

// GetPosition retrieves a specific position
func (s *LedgerService) GetPosition(ctx context.Context, portfolioID uuid.UUID, symbol string) (*models.Position, error) {
	query := `
		SELECT 
			portfolio_id, symbol, total_quantity, total_cost,
			avg_cost_per_share, last_updated
		FROM positions_current
		WHERE portfolio_id = $1 AND symbol = $2
	`

	var pos models.Position
	err := s.db.QueryRowContext(ctx, query, portfolioID, symbol).Scan(
		&pos.PortfolioID, &pos.Symbol, &pos.TotalQuantity, &pos.TotalCost,
		&pos.AvgCostPerShare, &pos.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("position not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query position: %w", err)
	}

	return &pos, nil
}

// CalculateUnrealizedPnL calculates unrealized P&L for a position
func (s *LedgerService) CalculateUnrealizedPnL(ctx context.Context, portfolioID uuid.UUID, symbol string, currentPrice decimal.Decimal) (*models.UnrealizedPnL, error) {
	query := `
		SELECT * FROM calculate_unrealized_pnl($1, $2, $3)
	`

	var pnl models.UnrealizedPnL
	err := s.db.QueryRowContext(ctx, query, portfolioID, symbol, currentPrice).Scan(
		&pnl.Symbol, &pnl.Quantity, &pnl.AvgCost, &pnl.CurrentPrice,
		&pnl.MarketValue, &pnl.CostBasis, &pnl.UnrealizedPnL, &pnl.UnrealizedPnLPct,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("position not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to calculate unrealized P&L: %w", err)
	}

	return &pnl, nil
}

// GetPortfolio retrieves portfolio details
func (s *LedgerService) GetPortfolio(ctx context.Context, portfolioID uuid.UUID) (*models.Portfolio, error) {
	query := `
		SELECT id, user_id, name, description, currency, created_at, updated_at
		FROM portfolios
		WHERE id = $1
	`

	var portfolio models.Portfolio
	err := s.db.QueryRowContext(ctx, query, portfolioID).Scan(
		&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description,
		&portfolio.Currency, &portfolio.CreatedAt, &portfolio.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("portfolio not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query portfolio: %w", err)
	}

	return &portfolio, nil
}

// GetUserPortfolios retrieves all portfolios for a user
func (s *LedgerService) GetUserPortfolios(ctx context.Context, userID uuid.UUID) ([]models.Portfolio, error) {
	query := `
		SELECT id, user_id, name, description, currency, created_at, updated_at
		FROM portfolios
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query portfolios: %w", err)
	}
	defer rows.Close()

	// 初始化為空數組而不是nil
	portfolios := make([]models.Portfolio, 0)
	for rows.Next() {
		var portfolio models.Portfolio
		err := rows.Scan(
			&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description,
			&portfolio.Currency, &portfolio.CreatedAt, &portfolio.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan portfolio: %w", err)
		}
		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}
