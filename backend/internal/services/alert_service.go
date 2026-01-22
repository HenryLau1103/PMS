package services

import (
	"context"
	"encoding/json"
	"fmt"
	"psm-backend/internal/database"
	"time"
)

// AlertService handles anomaly detection and alerts
type AlertService struct {
	db *database.DB
}

func NewAlertService(db *database.DB) *AlertService {
	return &AlertService{db: db}
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeVolumeSpike    AlertType = "volume_spike"
	AlertTypePriceBreakout  AlertType = "price_breakout"
	AlertTypeSentimentShift AlertType = "sentiment_shift"
	AlertTypeLimitHit       AlertType = "limit_hit"
	AlertTypeMABreakout     AlertType = "ma_breakout"
	AlertTypeRSIExtreme     AlertType = "rsi_extreme"
)

// AlertSeverity defines the severity level
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// StockAlert represents an alert
type StockAlert struct {
	ID             string          `json:"id"`
	Symbol         string          `json:"symbol"`
	AlertType      AlertType       `json:"alert_type"`
	Severity       AlertSeverity   `json:"severity"`
	Title          string          `json:"title"`
	Message        string          `json:"message"`
	Data           json.RawMessage `json:"data,omitempty"`
	TriggeredAt    time.Time       `json:"triggered_at"`
	AcknowledgedAt *time.Time      `json:"acknowledged_at,omitempty"`
	ReferencePrice float64         `json:"reference_price,omitempty"`
	ReferenceVolume int64          `json:"reference_volume,omitempty"`
	ThresholdValue float64         `json:"threshold_value,omitempty"`
}

// VolumeAnalysis represents volume analysis result
type VolumeAnalysis struct {
	Symbol           string  `json:"symbol"`
	CurrentVolume    int64   `json:"current_volume"`
	AvgVolume20Days  int64   `json:"avg_volume_20_days"`
	VolumeRatio      float64 `json:"volume_ratio"`
	IsSpike          bool    `json:"is_spike"`
	SpikeThreshold   float64 `json:"spike_threshold"`
}

// PriceAnalysis represents price analysis result
type PriceAnalysis struct {
	Symbol          string  `json:"symbol"`
	CurrentPrice    float64 `json:"current_price"`
	PreviousClose   float64 `json:"previous_close"`
	Change          float64 `json:"change"`
	ChangePercent   float64 `json:"change_percent"`
	High52Week      float64 `json:"high_52_week"`
	Low52Week       float64 `json:"low_52_week"`
	IsNear52WeekHigh bool   `json:"is_near_52_week_high"`
	IsNear52WeekLow  bool   `json:"is_near_52_week_low"`
}

// DetectVolumeSpike detects abnormal volume for a symbol
func (s *AlertService) DetectVolumeSpike(ctx context.Context, symbol string, threshold float64) (*VolumeAnalysis, error) {
	if threshold <= 0 {
		threshold = 2.0 // Default: 2x average volume
	}

	query := `
		WITH recent_data AS (
			SELECT 
				volume,
				timestamp,
				ROW_NUMBER() OVER (ORDER BY timestamp DESC) as rn
			FROM stock_ohlcv
			WHERE symbol = $1
			ORDER BY timestamp DESC
			LIMIT 21
		),
		latest AS (
			SELECT volume FROM recent_data WHERE rn = 1
		),
		avg_volume AS (
			SELECT COALESCE(AVG(volume)::bigint, 0) as avg_vol FROM recent_data WHERE rn > 1
		)
		SELECT 
			COALESCE(l.volume, 0),
			COALESCE(a.avg_vol, 0)
		FROM latest l, avg_volume a
	`

	var currentVolume, avgVolume int64
	err := s.db.QueryRowContext(ctx, query, symbol).Scan(&currentVolume, &avgVolume)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze volume: %w", err)
	}

	var ratio float64
	if avgVolume > 0 {
		ratio = float64(currentVolume) / float64(avgVolume)
	}

	analysis := &VolumeAnalysis{
		Symbol:          symbol,
		CurrentVolume:   currentVolume,
		AvgVolume20Days: avgVolume,
		VolumeRatio:     ratio,
		IsSpike:         ratio >= threshold,
		SpikeThreshold:  threshold,
	}

	// Create alert if spike detected
	if analysis.IsSpike {
		alertData, _ := json.Marshal(analysis)
		severity := AlertSeverityInfo
		if ratio >= 3.0 {
			severity = AlertSeverityWarning
		}
		if ratio >= 5.0 {
			severity = AlertSeverityCritical
		}

		alert := &StockAlert{
			Symbol:          symbol,
			AlertType:       AlertTypeVolumeSpike,
			Severity:        severity,
			Title:           fmt.Sprintf("%s 成交量異常", symbol),
			Message:         fmt.Sprintf("成交量達到20日均量的 %.1f 倍", ratio),
			Data:            alertData,
			ReferenceVolume: currentVolume,
			ThresholdValue:  threshold,
		}
		s.CreateAlert(ctx, alert)
	}

	return analysis, nil
}

// DetectPriceBreakout detects significant price movements
func (s *AlertService) DetectPriceBreakout(ctx context.Context, symbol string) (*PriceAnalysis, error) {
	query := `
		WITH price_data AS (
			SELECT 
				close,
				timestamp,
				ROW_NUMBER() OVER (ORDER BY timestamp DESC) as rn
			FROM stock_ohlcv
			WHERE symbol = $1
			ORDER BY timestamp DESC
			LIMIT 2
		),
		yearly_range AS (
			SELECT 
				MAX(high) as high_52,
				MIN(low) as low_52
			FROM stock_ohlcv
			WHERE symbol = $1 AND timestamp >= NOW() - INTERVAL '365 days'
		)
		SELECT 
			COALESCE((SELECT close FROM price_data WHERE rn = 1), 0),
			COALESCE((SELECT close FROM price_data WHERE rn = 2), 0),
			COALESCE(y.high_52, 0),
			COALESCE(y.low_52, 0)
		FROM yearly_range y
	`

	var currentPrice, prevClose, high52, low52 float64
	err := s.db.QueryRowContext(ctx, query, symbol).Scan(&currentPrice, &prevClose, &high52, &low52)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze price: %w", err)
	}

	var change, changePct float64
	if prevClose > 0 {
		change = currentPrice - prevClose
		changePct = change / prevClose * 100
	}

	// Near 52-week high/low threshold: within 3%
	nearThreshold := 0.03
	isNearHigh := high52 > 0 && (high52-currentPrice)/high52 <= nearThreshold
	isNearLow := low52 > 0 && (currentPrice-low52)/low52 <= nearThreshold

	analysis := &PriceAnalysis{
		Symbol:           symbol,
		CurrentPrice:     currentPrice,
		PreviousClose:    prevClose,
		Change:           change,
		ChangePercent:    changePct,
		High52Week:       high52,
		Low52Week:        low52,
		IsNear52WeekHigh: isNearHigh,
		IsNear52WeekLow:  isNearLow,
	}

	// Create alerts if near 52-week extremes
	if isNearHigh {
		alertData, _ := json.Marshal(analysis)
		s.CreateAlert(ctx, &StockAlert{
			Symbol:         symbol,
			AlertType:      AlertTypePriceBreakout,
			Severity:       AlertSeverityInfo,
			Title:          fmt.Sprintf("%s 接近52週新高", symbol),
			Message:        fmt.Sprintf("目前價格 %.2f 接近52週高點 %.2f", currentPrice, high52),
			Data:           alertData,
			ReferencePrice: currentPrice,
		})
	}

	if isNearLow {
		alertData, _ := json.Marshal(analysis)
		s.CreateAlert(ctx, &StockAlert{
			Symbol:         symbol,
			AlertType:      AlertTypePriceBreakout,
			Severity:       AlertSeverityWarning,
			Title:          fmt.Sprintf("%s 接近52週新低", symbol),
			Message:        fmt.Sprintf("目前價格 %.2f 接近52週低點 %.2f", currentPrice, low52),
			Data:           alertData,
			ReferencePrice: currentPrice,
		})
	}

	return analysis, nil
}

// ScanAllSymbols scans all symbols for anomalies
func (s *AlertService) ScanAllSymbols(ctx context.Context, volumeThreshold float64) (*ScanResult, error) {
	// Get all symbols with recent data
	query := `
		SELECT DISTINCT symbol 
		FROM stock_ohlcv 
		WHERE timestamp >= NOW() - INTERVAL '5 days'
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err == nil {
			symbols = append(symbols, symbol)
		}
	}

	result := &ScanResult{
		ScannedAt:     time.Now(),
		TotalSymbols:  len(symbols),
		VolumeSpikes:  []VolumeAnalysis{},
		PriceBreakouts: []PriceAnalysis{},
	}

	for _, symbol := range symbols {
		// Check volume
		volAnalysis, err := s.DetectVolumeSpike(ctx, symbol, volumeThreshold)
		if err == nil && volAnalysis.IsSpike {
			result.VolumeSpikes = append(result.VolumeSpikes, *volAnalysis)
		}

		// Check price
		priceAnalysis, err := s.DetectPriceBreakout(ctx, symbol)
		if err == nil && (priceAnalysis.IsNear52WeekHigh || priceAnalysis.IsNear52WeekLow) {
			result.PriceBreakouts = append(result.PriceBreakouts, *priceAnalysis)
		}
	}

	result.AlertsGenerated = len(result.VolumeSpikes) + len(result.PriceBreakouts)
	return result, nil
}

// ScanResult represents the result of a full scan
type ScanResult struct {
	ScannedAt       time.Time       `json:"scanned_at"`
	TotalSymbols    int             `json:"total_symbols"`
	AlertsGenerated int             `json:"alerts_generated"`
	VolumeSpikes    []VolumeAnalysis `json:"volume_spikes"`
	PriceBreakouts  []PriceAnalysis  `json:"price_breakouts"`
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(ctx context.Context, alert *StockAlert) error {
	query := `
		INSERT INTO stock_alerts (symbol, alert_type, severity, title, message, data, reference_price, reference_volume, threshold_value)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, triggered_at
	`

	return s.db.QueryRowContext(ctx, query,
		alert.Symbol,
		string(alert.AlertType),
		string(alert.Severity),
		alert.Title,
		alert.Message,
		alert.Data,
		alert.ReferencePrice,
		alert.ReferenceVolume,
		alert.ThresholdValue,
	).Scan(&alert.ID, &alert.TriggeredAt)
}

// GetAlerts retrieves alerts with optional filters
func (s *AlertService) GetAlerts(ctx context.Context, symbol string, unacknowledgedOnly bool, limit int) ([]StockAlert, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, symbol, alert_type, severity, title, message, COALESCE(data, '{}'), triggered_at, acknowledged_at,
		       COALESCE(reference_price, 0), COALESCE(reference_volume, 0), COALESCE(threshold_value, 0)
		FROM stock_alerts
		WHERE ($1 = '' OR symbol = $1)
		  AND ($2 = FALSE OR acknowledged_at IS NULL)
		ORDER BY triggered_at DESC
		LIMIT $3
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, unacknowledgedOnly, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []StockAlert
	for rows.Next() {
		var a StockAlert
		if err := rows.Scan(
			&a.ID, &a.Symbol, &a.AlertType, &a.Severity, &a.Title, &a.Message, &a.Data,
			&a.TriggeredAt, &a.AcknowledgedAt, &a.ReferencePrice, &a.ReferenceVolume, &a.ThresholdValue,
		); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *AlertService) AcknowledgeAlert(ctx context.Context, alertID string) error {
	query := `UPDATE stock_alerts SET acknowledged_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, alertID)
	return err
}

// GetAlertStats returns alert statistics
func (s *AlertService) GetAlertStats(ctx context.Context, days int) (*AlertStats, error) {
	if days <= 0 {
		days = 7
	}

	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical,
			COUNT(CASE WHEN severity = 'warning' THEN 1 END) as warning,
			COUNT(CASE WHEN severity = 'info' THEN 1 END) as info,
			COUNT(CASE WHEN acknowledged_at IS NULL THEN 1 END) as unacknowledged
		FROM stock_alerts
		WHERE triggered_at >= NOW() - $1::interval
	`

	var stats AlertStats
	stats.Days = days
	interval := fmt.Sprintf("%d days", days)
	
	err := s.db.QueryRowContext(ctx, query, interval).Scan(
		&stats.Total,
		&stats.Critical,
		&stats.Warning,
		&stats.Info,
		&stats.Unacknowledged,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// AlertStats represents alert statistics
type AlertStats struct {
	Days           int `json:"days"`
	Total          int `json:"total"`
	Critical       int `json:"critical"`
	Warning        int `json:"warning"`
	Info           int `json:"info"`
	Unacknowledged int `json:"unacknowledged"`
}
