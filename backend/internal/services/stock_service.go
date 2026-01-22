package services

import (
	"context"
	"database/sql"
	"fmt"
	"psm-backend/internal/database"
	"psm-backend/internal/models"
)

type StockService struct {
	db *database.DB
}

func NewStockService(db *database.DB) *StockService {
	return &StockService{db: db}
}

// SearchStocks searches for Taiwan stocks by symbol or name
// Supports partial matching for autocomplete functionality
// Prioritizes symbol search for better reliability
func (s *StockService) SearchStocks(ctx context.Context, query string, limit int) ([]models.TaiwanStock, error) {
	if limit <= 0 || limit > 50 {
		limit = 20 // Default limit
	}

	// Primary search: symbol prefix match (most common use case)
	sqlQuery := `
		SELECT symbol, name, name_en, market, industry, is_active, created_at, updated_at
		FROM taiwan_stocks
		WHERE is_active = true 
		  AND symbol LIKE $1 || '%'
		ORDER BY symbol
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search stocks: %w", err)
	}
	defer rows.Close()

	stocks := make([]models.TaiwanStock, 0)
	for rows.Next() {
		var stock models.TaiwanStock
		err := rows.Scan(
			&stock.Symbol,
			&stock.Name,
			&stock.NameEn,
			&stock.Market,
			&stock.Industry,
			&stock.IsActive,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}
		stocks = append(stocks, stock)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	return stocks, nil
}

// GetStockBySymbol retrieves a single stock by exact symbol match
func (s *StockService) GetStockBySymbol(ctx context.Context, symbol string) (*models.TaiwanStock, error) {
	sqlQuery := `
		SELECT symbol, name, name_en, market, industry, is_active, created_at, updated_at
		FROM taiwan_stocks
		WHERE symbol = $1
	`

	var stock models.TaiwanStock
	err := s.db.QueryRowContext(ctx, sqlQuery, symbol).Scan(
		&stock.Symbol,
		&stock.Name,
		&stock.NameEn,
		&stock.Market,
		&stock.Industry,
		&stock.IsActive,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stock not found: %s", symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return &stock, nil
}
