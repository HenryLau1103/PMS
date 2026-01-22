package models

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// EventType represents the type of ledger event
type EventType string

const (
	EventTypeBuy        EventType = "BUY"
	EventTypeSell       EventType = "SELL"
	EventTypeDividend   EventType = "DIVIDEND"
	EventTypeSplit      EventType = "SPLIT"
	EventTypeRights     EventType = "RIGHTS"
	EventTypeCorrection EventType = "CORRECTION"
)

// LedgerEvent represents a transaction in the immutable ledger
type LedgerEvent struct {
	EventID     uuid.UUID       `json:"event_id" db:"event_id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	PortfolioID uuid.UUID       `json:"portfolio_id" db:"portfolio_id"`
	EventType   EventType       `json:"event_type" db:"event_type"`
	Symbol      string          `json:"symbol" db:"symbol"`
	Quantity    decimal.Decimal `json:"quantity" db:"quantity"`
	Price       decimal.Decimal `json:"price" db:"price"`
	Fee         decimal.Decimal `json:"fee" db:"fee"`
	Tax         decimal.Decimal `json:"tax" db:"tax"`
	TotalAmount decimal.Decimal `json:"total_amount" db:"total_amount"`
	OccurredAt  time.Time       `json:"occurred_at" db:"occurred_at"`
	RecordedAt  time.Time       `json:"recorded_at" db:"recorded_at"`
	Source      string          `json:"source" db:"source"`
	Notes       *string         `json:"notes,omitempty" db:"notes"`
	Payload     *string         `json:"payload,omitempty" db:"payload"`
}

// CreateLedgerEventRequest is the payload for creating a new transaction
type CreateLedgerEventRequest struct {
	PortfolioID uuid.UUID       `json:"portfolio_id" validate:"required"`
	EventType   EventType       `json:"event_type" validate:"required"`
	Symbol      string          `json:"symbol" validate:"required,taiwan_symbol"`
	Quantity    decimal.Decimal `json:"quantity" validate:"required,gt=0"`
	Price       decimal.Decimal `json:"price" validate:"required,gte=0"`
	Fee         decimal.Decimal `json:"fee" validate:"gte=0"`
	Tax         decimal.Decimal `json:"tax" validate:"gte=0"`
	OccurredAt  time.Time       `json:"occurred_at" validate:"required"`
	Notes       *string         `json:"notes,omitempty"`
}

// Position represents current holdings for a symbol
type Position struct {
	PortfolioID      uuid.UUID       `json:"portfolio_id" db:"portfolio_id"`
	Symbol           string          `json:"symbol" db:"symbol"`
	TotalQuantity    decimal.Decimal `json:"total_quantity" db:"total_quantity"`
	TotalCost        decimal.Decimal `json:"total_cost" db:"total_cost"`
	AvgCostPerShare  decimal.Decimal `json:"avg_cost_per_share" db:"avg_cost_per_share"`
	LastUpdated      time.Time       `json:"last_updated" db:"last_updated"`
}

// UnrealizedPnL represents unrealized profit/loss for a position
type UnrealizedPnL struct {
	Symbol            string          `json:"symbol"`
	Quantity          decimal.Decimal `json:"quantity"`
	AvgCost           decimal.Decimal `json:"avg_cost"`
	CurrentPrice      decimal.Decimal `json:"current_price"`
	MarketValue       decimal.Decimal `json:"market_value"`
	CostBasis         decimal.Decimal `json:"cost_basis"`
	UnrealizedPnL     decimal.Decimal `json:"unrealized_pnl"`
	UnrealizedPnLPct  decimal.Decimal `json:"unrealized_pnl_pct"`
}

// Portfolio represents a user's portfolio
type Portfolio struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	Currency    string     `json:"currency" db:"currency"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// User represents a system user
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// RealizedPnL represents closed position profit/loss
type RealizedPnL struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	PortfolioID  uuid.UUID       `json:"portfolio_id" db:"portfolio_id"`
	Symbol       string          `json:"symbol" db:"symbol"`
	BuyEventID   uuid.UUID       `json:"buy_event_id" db:"buy_event_id"`
	SellEventID  uuid.UUID       `json:"sell_event_id" db:"sell_event_id"`
	Quantity     decimal.Decimal `json:"quantity" db:"quantity"`
	BuyPrice     decimal.Decimal `json:"buy_price" db:"buy_price"`
	SellPrice    decimal.Decimal `json:"sell_price" db:"sell_price"`
	RealizedPnL  decimal.Decimal `json:"realized_pnl" db:"realized_pnl"`
	TotalFees    decimal.Decimal `json:"total_fees" db:"total_fees"`
	TotalTaxes   decimal.Decimal `json:"total_taxes" db:"total_taxes"`
	BuyDate      time.Time       `json:"buy_date" db:"buy_date"`
	SellDate     time.Time       `json:"sell_date" db:"sell_date"`
	HoldingDays  int             `json:"holding_days" db:"holding_days"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// TaiwanStock represents a Taiwan stock symbol
type TaiwanStock struct {
	Symbol    string     `json:"symbol" db:"symbol"`
	Name      string     `json:"name" db:"name"`
	NameEn    *string    `json:"name_en,omitempty" db:"name_en"`
	Market    string     `json:"market" db:"market"`
	Industry  *string    `json:"industry,omitempty" db:"industry"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}
