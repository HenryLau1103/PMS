package services

import (
	"context"
	"fmt"
	"psm-backend/internal/database"
	"sort"
)

// ScreenerService handles stock screening and recommendations
type ScreenerService struct {
	db *database.DB
}

func NewScreenerService(db *database.DB) *ScreenerService {
	return &ScreenerService{db: db}
}

// ScreenerCriteria defines screening criteria
type ScreenerCriteria struct {
	// Price criteria
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	
	// Volume criteria
	MinVolume         int64   `json:"min_volume"`
	MinVolumeRatio    float64 `json:"min_volume_ratio"`    // vs 20-day avg
	
	// Technical criteria
	AboveMA20         bool    `json:"above_ma20"`
	AboveMA60         bool    `json:"above_ma60"`
	RSIMin            float64 `json:"rsi_min"`
	RSIMax            float64 `json:"rsi_max"`
	GoldenCross       bool    `json:"golden_cross"`        // MA5 > MA20 recently
	
	// Performance criteria
	MinChangePercent  float64 `json:"min_change_percent"`
	MaxChangePercent  float64 `json:"max_change_percent"`
	Near52WeekHigh    bool    `json:"near_52_week_high"`
	Near52WeekLow     bool    `json:"near_52_week_low"`
	
	// Sentiment criteria
	PositiveSentiment bool    `json:"positive_sentiment"`
	
	// Sorting and limits
	SortBy            string  `json:"sort_by"` // volume_ratio, change_percent, rsi
	SortDesc          bool    `json:"sort_desc"`
	Limit             int     `json:"limit"`
}

// ScreenerResult represents a single screening result
type ScreenerResult struct {
	Symbol           string   `json:"symbol"`
	Name             string   `json:"name"`
	CurrentPrice     float64  `json:"current_price"`
	PreviousClose    float64  `json:"previous_close"`
	Change           float64  `json:"change"`
	ChangePercent    float64  `json:"change_percent"`
	Volume           int64    `json:"volume"`
	AvgVolume        int64    `json:"avg_volume"`
	VolumeRatio      float64  `json:"volume_ratio"`
	High52Week       float64  `json:"high_52_week"`
	Low52Week        float64  `json:"low_52_week"`
	MA5              float64  `json:"ma5"`
	MA20             float64  `json:"ma20"`
	MA60             float64  `json:"ma60"`
	RSI              float64  `json:"rsi"`
	Sentiment        string   `json:"sentiment"`
	SentimentScore   float64  `json:"sentiment_score"`
	Score            float64  `json:"score"`  // Composite score
	MatchedCriteria  []string `json:"matched_criteria"`
}

// ScreenStocks screens stocks based on criteria
func (s *ScreenerService) ScreenStocks(ctx context.Context, criteria *ScreenerCriteria) ([]ScreenerResult, error) {
	if criteria.Limit <= 0 {
		criteria.Limit = 50
	}

	// Get all stocks with recent data and calculate metrics
	query := `
		WITH recent_prices AS (
			SELECT 
				symbol,
				close as current_price,
				volume,
				timestamp,
				LAG(close) OVER (PARTITION BY symbol ORDER BY timestamp) as prev_close,
				ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) as rn
			FROM stock_ohlcv
			WHERE timestamp >= NOW() - INTERVAL '2 days'
		),
		latest_prices AS (
			SELECT symbol, current_price, volume, prev_close
			FROM recent_prices
			WHERE rn = 1
		),
		moving_averages AS (
			SELECT 
				symbol,
				AVG(CASE WHEN rn <= 5 THEN close END) as ma5,
				AVG(CASE WHEN rn <= 20 THEN close END) as ma20,
				AVG(CASE WHEN rn <= 60 THEN close END) as ma60,
				AVG(CASE WHEN rn <= 20 THEN volume END)::bigint as avg_volume
			FROM (
				SELECT symbol, close, volume,
					   ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) as rn
				FROM stock_ohlcv
				WHERE timestamp >= NOW() - INTERVAL '90 days'
			) sub
			GROUP BY symbol
		),
		yearly_range AS (
			SELECT 
				symbol,
				MAX(high) as high_52,
				MIN(low) as low_52
			FROM stock_ohlcv
			WHERE timestamp >= NOW() - INTERVAL '365 days'
			GROUP BY symbol
		),
		sentiment_data AS (
			SELECT 
				symbol,
				CASE 
					WHEN AVG(sentiment_score) > 0.15 THEN 'positive'
					WHEN AVG(sentiment_score) < -0.15 THEN 'negative'
					ELSE 'neutral'
				END as sentiment,
				COALESCE(AVG(sentiment_score), 0) as sentiment_score
			FROM stock_news
			WHERE published_at >= NOW() - INTERVAL '7 days' AND sentiment_score IS NOT NULL
			GROUP BY symbol
		)
		SELECT 
			lp.symbol,
			COALESCE(st.name, st.name_en, lp.symbol) as name,
			lp.current_price,
			COALESCE(lp.prev_close, lp.current_price) as prev_close,
			lp.volume,
			COALESCE(ma.avg_volume, 0) as avg_volume,
			COALESCE(ma.ma5, 0) as ma5,
			COALESCE(ma.ma20, 0) as ma20,
			COALESCE(ma.ma60, 0) as ma60,
			COALESCE(yr.high_52, 0) as high_52,
			COALESCE(yr.low_52, 0) as low_52,
			COALESCE(sd.sentiment, 'unknown') as sentiment,
			COALESCE(sd.sentiment_score, 0) as sentiment_score
		FROM latest_prices lp
		LEFT JOIN moving_averages ma ON lp.symbol = ma.symbol
		LEFT JOIN yearly_range yr ON lp.symbol = yr.symbol
		LEFT JOIN sentiment_data sd ON lp.symbol = sd.symbol
		LEFT JOIN taiwan_stocks st ON lp.symbol = st.symbol
		WHERE lp.current_price > 0
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to screen stocks: %w", err)
	}
	defer rows.Close()

	var results []ScreenerResult
	for rows.Next() {
		var r ScreenerResult
		if err := rows.Scan(
			&r.Symbol, &r.Name, &r.CurrentPrice, &r.PreviousClose, &r.Volume,
			&r.AvgVolume, &r.MA5, &r.MA20, &r.MA60, &r.High52Week, &r.Low52Week,
			&r.Sentiment, &r.SentimentScore,
		); err != nil {
			continue
		}

		// Calculate derived metrics
		if r.PreviousClose > 0 {
			r.Change = r.CurrentPrice - r.PreviousClose
			r.ChangePercent = r.Change / r.PreviousClose * 100
		}
		if r.AvgVolume > 0 {
			r.VolumeRatio = float64(r.Volume) / float64(r.AvgVolume)
		}

		// Apply filters
		if !s.matchesCriteria(&r, criteria) {
			continue
		}

		// Calculate composite score
		r.Score = s.calculateScore(&r, criteria)

		results = append(results, r)
	}

	// Sort results
	s.sortResults(results, criteria)

	// Apply limit
	if len(results) > criteria.Limit {
		results = results[:criteria.Limit]
	}

	return results, nil
}

func (s *ScreenerService) matchesCriteria(r *ScreenerResult, c *ScreenerCriteria) bool {
	r.MatchedCriteria = []string{}

	// Price filters
	if c.MinPrice > 0 && r.CurrentPrice < c.MinPrice {
		return false
	}
	if c.MaxPrice > 0 && r.CurrentPrice > c.MaxPrice {
		return false
	}

	// Volume filters
	if c.MinVolume > 0 && r.Volume < c.MinVolume {
		return false
	}
	if c.MinVolumeRatio > 0 && r.VolumeRatio < c.MinVolumeRatio {
		return false
	}
	if c.MinVolumeRatio > 0 && r.VolumeRatio >= c.MinVolumeRatio {
		r.MatchedCriteria = append(r.MatchedCriteria, "成交量放大")
	}

	// MA filters
	if c.AboveMA20 && r.MA20 > 0 && r.CurrentPrice <= r.MA20 {
		return false
	}
	if c.AboveMA20 && r.CurrentPrice > r.MA20 {
		r.MatchedCriteria = append(r.MatchedCriteria, "站上20日均線")
	}

	if c.AboveMA60 && r.MA60 > 0 && r.CurrentPrice <= r.MA60 {
		return false
	}
	if c.AboveMA60 && r.CurrentPrice > r.MA60 {
		r.MatchedCriteria = append(r.MatchedCriteria, "站上60日均線")
	}

	// Golden cross check (MA5 > MA20)
	if c.GoldenCross && r.MA5 > 0 && r.MA20 > 0 && r.MA5 <= r.MA20 {
		return false
	}
	if c.GoldenCross && r.MA5 > r.MA20 {
		r.MatchedCriteria = append(r.MatchedCriteria, "黃金交叉")
	}

	// Change percent filters
	if c.MinChangePercent != 0 && r.ChangePercent < c.MinChangePercent {
		return false
	}
	if c.MaxChangePercent != 0 && r.ChangePercent > c.MaxChangePercent {
		return false
	}

	// 52-week filters
	if c.Near52WeekHigh && r.High52Week > 0 {
		threshold := (r.High52Week - r.CurrentPrice) / r.High52Week
		if threshold > 0.03 {
			return false
		}
		r.MatchedCriteria = append(r.MatchedCriteria, "接近52週新高")
	}
	if c.Near52WeekLow && r.Low52Week > 0 {
		threshold := (r.CurrentPrice - r.Low52Week) / r.Low52Week
		if threshold > 0.03 {
			return false
		}
		r.MatchedCriteria = append(r.MatchedCriteria, "接近52週新低")
	}

	// Sentiment filter
	if c.PositiveSentiment && r.Sentiment != "positive" {
		return false
	}
	if c.PositiveSentiment && r.Sentiment == "positive" {
		r.MatchedCriteria = append(r.MatchedCriteria, "正面情緒")
	}

	return true
}

func (s *ScreenerService) calculateScore(r *ScreenerResult, c *ScreenerCriteria) float64 {
	score := 0.0

	// Volume factor (0-20 points)
	if r.VolumeRatio > 1 {
		score += min(r.VolumeRatio*5, 20)
	}

	// Trend factor (0-30 points)
	if r.CurrentPrice > r.MA20 && r.MA20 > 0 {
		score += 10
	}
	if r.CurrentPrice > r.MA60 && r.MA60 > 0 {
		score += 10
	}
	if r.MA5 > r.MA20 && r.MA5 > 0 && r.MA20 > 0 {
		score += 10
	}

	// Momentum factor (0-20 points)
	if r.ChangePercent > 0 {
		score += min(r.ChangePercent*2, 20)
	}

	// Sentiment factor (0-20 points)
	if r.Sentiment == "positive" {
		score += 10 + r.SentimentScore*10
	} else if r.Sentiment == "neutral" {
		score += 5
	}

	// 52-week position factor (0-10 points)
	if r.High52Week > 0 && r.Low52Week > 0 {
		position := (r.CurrentPrice - r.Low52Week) / (r.High52Week - r.Low52Week)
		if position > 0.8 {
			score += 10 // Near high is bullish
		}
	}

	return score
}

func (s *ScreenerService) sortResults(results []ScreenerResult, c *ScreenerCriteria) {
	sort.Slice(results, func(i, j int) bool {
		var vi, vj float64
		switch c.SortBy {
		case "volume_ratio":
			vi, vj = results[i].VolumeRatio, results[j].VolumeRatio
		case "change_percent":
			vi, vj = results[i].ChangePercent, results[j].ChangePercent
		case "sentiment_score":
			vi, vj = results[i].SentimentScore, results[j].SentimentScore
		default: // score
			vi, vj = results[i].Score, results[j].Score
		}
		if c.SortDesc {
			return vi > vj
		}
		return vi < vj
	})
}

// PresetScreen represents a preset screening configuration
type PresetScreen struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Criteria    ScreenerCriteria `json:"criteria"`
}

// GetPresets returns available screening presets
func (s *ScreenerService) GetPresets() []PresetScreen {
	return []PresetScreen{
		{
			Name:        "volume_breakout",
			Description: "成交量突破 - 成交量達2倍以上均量的股票",
			Criteria: ScreenerCriteria{
				MinVolumeRatio: 2.0,
				MinPrice:       10,
				SortBy:         "volume_ratio",
				SortDesc:       true,
				Limit:          20,
			},
		},
		{
			Name:        "golden_cross",
			Description: "黃金交叉 - MA5突破MA20的股票",
			Criteria: ScreenerCriteria{
				GoldenCross: true,
				AboveMA20:   true,
				MinPrice:    10,
				SortBy:      "score",
				SortDesc:    true,
				Limit:       20,
			},
		},
		{
			Name:        "trend_following",
			Description: "趨勢跟蹤 - 站上60日均線且正向動能",
			Criteria: ScreenerCriteria{
				AboveMA60:        true,
				MinChangePercent: 0,
				MinPrice:         10,
				SortBy:           "change_percent",
				SortDesc:         true,
				Limit:            20,
			},
		},
		{
			Name:        "positive_sentiment",
			Description: "正面情緒 - 近期新聞情緒正面的股票",
			Criteria: ScreenerCriteria{
				PositiveSentiment: true,
				MinPrice:          10,
				SortBy:            "sentiment_score",
				SortDesc:          true,
				Limit:             20,
			},
		},
		{
			Name:        "52_week_high",
			Description: "52週新高 - 接近52週最高價的股票",
			Criteria: ScreenerCriteria{
				Near52WeekHigh: true,
				MinPrice:       10,
				SortBy:         "change_percent",
				SortDesc:       true,
				Limit:          20,
			},
		},
		{
			Name:        "value_hunting",
			Description: "價值獵手 - 接近52週低點的潛在反彈股",
			Criteria: ScreenerCriteria{
				Near52WeekLow:    true,
				MinVolumeRatio:   1.2,
				MinChangePercent: 0,
				MinPrice:         10,
				SortBy:           "volume_ratio",
				SortDesc:         true,
				Limit:            20,
			},
		},
	}
}

// RunPreset runs a preset screening
func (s *ScreenerService) RunPreset(ctx context.Context, presetName string) ([]ScreenerResult, error) {
	presets := s.GetPresets()
	for _, p := range presets {
		if p.Name == presetName {
			return s.ScreenStocks(ctx, &p.Criteria)
		}
	}
	return nil, fmt.Errorf("preset not found: %s", presetName)
}
