package services

import (
	"context"
	"encoding/json"
	"fmt"
	"psm-backend/internal/database"
	"time"

	"github.com/markcheno/go-talib"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type TechnicalAnalysisService struct {
	db          *database.DB
	redisClient *redis.Client
}

func NewTechnicalAnalysisService(db *database.DB, redisClient *redis.Client) *TechnicalAnalysisService {
	return &TechnicalAnalysisService{
		db:          db,
		redisClient: redisClient,
	}
}

// IndicatorData represents calculated indicator results
type IndicatorData struct {
	Symbol     string                   `json:"symbol"`
	Indicator  string                   `json:"indicator"`
	Timeframe  string                   `json:"timeframe"`
	Params     map[string]interface{}   `json:"params"`
	Values     []IndicatorValue         `json:"values"`
	Calculated time.Time                `json:"calculated_at"`
}

type IndicatorValue struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     interface{}            `json:"value"` // Can be single value or multiple (e.g., MACD has 3 lines)
}

// MAResult represents Moving Average result
type MAResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Value     decimal.Decimal `json:"value"`
}

// RSIResult represents RSI indicator result
type RSIResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Value     decimal.Decimal `json:"value"`
}

// MACDResult represents MACD indicator result (3 lines)
type MACDResult struct {
	Timestamp time.Time       `json:"timestamp"`
	MACD      decimal.Decimal `json:"macd"`
	Signal    decimal.Decimal `json:"signal"`
	Histogram decimal.Decimal `json:"histogram"`
}

// BBResult represents Bollinger Bands result (3 bands)
type BBResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Upper     decimal.Decimal `json:"upper"`
	Middle    decimal.Decimal `json:"middle"`
	Lower     decimal.Decimal `json:"lower"`
}

// KDJResult represents KDJ indicator result (3 lines)
type KDJResult struct {
	Timestamp time.Time       `json:"timestamp"`
	K         decimal.Decimal `json:"k"`
	D         decimal.Decimal `json:"d"`
	J         decimal.Decimal `json:"j"`
}

// CalculateMA calculates Moving Average (SMA or EMA)
func (s *TechnicalAnalysisService) CalculateMA(ctx context.Context, symbol string, period int, maType string, limit int) ([]MAResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("indicator:%s:MA:%s:%d", symbol, maType, period)
	if cached, err := s.getCache(ctx, cacheKey); err == nil && cached != nil {
		var results []MAResult
		if err := json.Unmarshal(cached, &results); err == nil {
			// Return last N results
			if len(results) > limit {
				return results[len(results)-limit:], nil
			}
			return results, nil
		}
	}

	// Fetch ALL OHLCV data (we need enough for calculation)
	ohlcv, err := s.getOHLCVData(ctx, symbol, 1000) // Get more data
	if err != nil {
		return nil, err
	}

	if len(ohlcv) < period {
		return nil, fmt.Errorf("insufficient data: need %d, got %d", period, len(ohlcv))
	}

	// Extract close prices
	closePrices := make([]float64, len(ohlcv))
	for i, candle := range ohlcv {
		closePrices[i], _ = candle.Close.Float64()
	}

	// Calculate MA
	var maValues []float64
	if maType == "EMA" {
		maValues = talib.Ema(closePrices, period)
	} else {
		// Default to SMA
		maValues = talib.Sma(closePrices, period)
	}

	// Build results
	results := make([]MAResult, 0)
	for i, val := range maValues {
		if val != 0 && !isNaN(val) { // Skip invalid values
			results = append(results, MAResult{
				Timestamp: ohlcv[i].Timestamp,
				Value:     decimal.NewFromFloat(val),
			})
		}
	}

	// Cache ALL results
	if data, err := json.Marshal(results); err == nil {
		s.setCache(ctx, cacheKey, data, 24*time.Hour)
	}

	// Return last N results
	if len(results) > limit {
		results = results[len(results)-limit:]
	}

	return results, nil
}

// CalculateRSI calculates Relative Strength Index
func (s *TechnicalAnalysisService) CalculateRSI(ctx context.Context, symbol string, period int, limit int) ([]RSIResult, error) {
	cacheKey := fmt.Sprintf("indicator:%s:RSI:%d", symbol, period)
	if cached, err := s.getCache(ctx, cacheKey); err == nil && cached != nil {
		var results []RSIResult
		if err := json.Unmarshal(cached, &results); err == nil {
			if len(results) > limit {
				return results[len(results)-limit:], nil
			}
			return results, nil
		}
	}

	ohlcv, err := s.getOHLCVData(ctx, symbol, 1000)
	if err != nil {
		return nil, err
	}

	if len(ohlcv) < period {
		return nil, fmt.Errorf("insufficient data for RSI")
	}

	closePrices := make([]float64, len(ohlcv))
	for i, candle := range ohlcv {
		closePrices[i], _ = candle.Close.Float64()
	}

	rsiValues := talib.Rsi(closePrices, period)

	results := make([]RSIResult, 0)
	for i, val := range rsiValues {
		if val != 0 && !isNaN(val) {
			results = append(results, RSIResult{
				Timestamp: ohlcv[i].Timestamp,
				Value:     decimal.NewFromFloat(val),
			})
		}
	}

	if data, err := json.Marshal(results); err == nil {
		s.setCache(ctx, cacheKey, data, 24*time.Hour)
	}

	if len(results) > limit {
		results = results[len(results)-limit:]
	}

	return results, nil
}

// CalculateMACD calculates MACD indicator (12, 26, 9 default)
func (s *TechnicalAnalysisService) CalculateMACD(ctx context.Context, symbol string, fastPeriod, slowPeriod, signalPeriod int, limit int) ([]MACDResult, error) {
	cacheKey := fmt.Sprintf("indicator:%s:MACD:%d:%d:%d", symbol, fastPeriod, slowPeriod, signalPeriod)
	if cached, err := s.getCache(ctx, cacheKey); err == nil && cached != nil {
		var results []MACDResult
		if err := json.Unmarshal(cached, &results); err == nil {
			if len(results) > limit {
				return results[len(results)-limit:], nil
			}
			return results, nil
		}
	}

	ohlcv, err := s.getOHLCVData(ctx, symbol, 1000)
	if err != nil {
		return nil, err
	}

	closePrices := make([]float64, len(ohlcv))
	for i, candle := range ohlcv {
		closePrices[i], _ = candle.Close.Float64()
	}

	macd, signal, histogram := talib.Macd(closePrices, fastPeriod, slowPeriod, signalPeriod)

	results := make([]MACDResult, 0)
	for i := range macd {
		if !isNaN(macd[i]) && !isNaN(signal[i]) && !isNaN(histogram[i]) {
			results = append(results, MACDResult{
				Timestamp: ohlcv[i].Timestamp,
				MACD:      decimal.NewFromFloat(macd[i]),
				Signal:    decimal.NewFromFloat(signal[i]),
				Histogram: decimal.NewFromFloat(histogram[i]),
			})
		}
	}

	if data, err := json.Marshal(results); err == nil {
		s.setCache(ctx, cacheKey, data, 24*time.Hour)
	}

	if len(results) > limit {
		results = results[len(results)-limit:]
	}

	return results, nil
}

// CalculateBollingerBands calculates Bollinger Bands (20, 2 default)
func (s *TechnicalAnalysisService) CalculateBollingerBands(ctx context.Context, symbol string, period int, stdDev float64, limit int) ([]BBResult, error) {
	cacheKey := fmt.Sprintf("indicator:%s:BB:%d:%.1f", symbol, period, stdDev)
	if cached, err := s.getCache(ctx, cacheKey); err == nil && cached != nil {
		var results []BBResult
		if err := json.Unmarshal(cached, &results); err == nil {
			if len(results) > limit {
				return results[len(results)-limit:], nil
			}
			return results, nil
		}
	}

	ohlcv, err := s.getOHLCVData(ctx, symbol, 1000)
	if err != nil {
		return nil, err
	}

	closePrices := make([]float64, len(ohlcv))
	for i, candle := range ohlcv {
		closePrices[i], _ = candle.Close.Float64()
	}

	upper, middle, lower := talib.BBands(closePrices, period, stdDev, stdDev, talib.SMA)

	results := make([]BBResult, 0)
	for i := range upper {
		if !isNaN(upper[i]) && !isNaN(middle[i]) && !isNaN(lower[i]) {
			results = append(results, BBResult{
				Timestamp: ohlcv[i].Timestamp,
				Upper:     decimal.NewFromFloat(upper[i]),
				Middle:    decimal.NewFromFloat(middle[i]),
				Lower:     decimal.NewFromFloat(lower[i]),
			})
		}
	}

	if data, err := json.Marshal(results); err == nil {
		s.setCache(ctx, cacheKey, data, 24*time.Hour)
	}

	if len(results) > limit {
		results = results[len(results)-limit:]
	}

	return results, nil
}

// CalculateKDJ calculates KDJ indicator (Stochastic with J line)
func (s *TechnicalAnalysisService) CalculateKDJ(ctx context.Context, symbol string, period int, limit int) ([]KDJResult, error) {
	cacheKey := fmt.Sprintf("indicator:%s:KDJ:%d", symbol, period)
	if cached, err := s.getCache(ctx, cacheKey); err == nil && cached != nil {
		var results []KDJResult
		if err := json.Unmarshal(cached, &results); err == nil {
			if len(results) > limit {
				return results[len(results)-limit:], nil
			}
			return results, nil
		}
	}

	ohlcv, err := s.getOHLCVData(ctx, symbol, 1000)
	if err != nil {
		return nil, err
	}

	highs := make([]float64, len(ohlcv))
	lows := make([]float64, len(ohlcv))
	closes := make([]float64, len(ohlcv))

	for i, candle := range ohlcv {
		highs[i], _ = candle.High.Float64()
		lows[i], _ = candle.Low.Float64()
		closes[i], _ = candle.Close.Float64()
	}

	// Calculate K and D using Stochastic
	k, d := talib.Stoch(highs, lows, closes, period, 3, talib.SMA, 3, talib.SMA)

	results := make([]KDJResult, 0)
	for i := range k {
		if !isNaN(k[i]) && !isNaN(d[i]) {
			// J = 3*K - 2*D
			j := 3*k[i] - 2*d[i]
			results = append(results, KDJResult{
				Timestamp: ohlcv[i].Timestamp,
				K:         decimal.NewFromFloat(k[i]),
				D:         decimal.NewFromFloat(d[i]),
				J:         decimal.NewFromFloat(j),
			})
		}
	}

	if data, err := json.Marshal(results); err == nil {
		s.setCache(ctx, cacheKey, data, 24*time.Hour)
	}

	if len(results) > limit {
		results = results[len(results)-limit:]
	}

	return results, nil
}

// Helper: Get OHLCV data from database
func (s *TechnicalAnalysisService) getOHLCVData(ctx context.Context, symbol string, limit int) ([]OHLCV, error) {
	query := `
		SELECT symbol, timestamp, open, high, low, close, volume, turnover
		FROM stock_ohlcv
		WHERE symbol = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []OHLCV
	for rows.Next() {
		var ohlcv OHLCV
		err := rows.Scan(
			&ohlcv.Symbol,
			&ohlcv.Timestamp,
			&ohlcv.Open,
			&ohlcv.High,
			&ohlcv.Low,
			&ohlcv.Close,
			&ohlcv.Volume,
			&ohlcv.Turnover,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, ohlcv)
	}

	// Reverse to chronological order (oldest first) for TA-Lib
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results, nil
}

// Helper: Get from Redis cache
func (s *TechnicalAnalysisService) getCache(ctx context.Context, key string) ([]byte, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}
	return s.redisClient.Get(ctx, key).Bytes()
}

// Helper: Set Redis cache
func (s *TechnicalAnalysisService) setCache(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if s.redisClient == nil {
		return nil // Silently skip if Redis not available
	}
	return s.redisClient.Set(ctx, key, data, ttl).Err()
}

// Helper: Check if float is NaN
func isNaN(f float64) bool {
	return f != f
}
