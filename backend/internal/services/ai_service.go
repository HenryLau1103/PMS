package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"psm-backend/internal/database"
	"strings"
	"time"
)

// AIService handles Google Gemini integration for stock analysis
type AIService struct {
	db         *database.DB
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewAIService(db *database.DB) *AIService {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}

	return &AIService{
		db:     db,
		apiKey: os.Getenv("GEMINI_API_KEY"),
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// HasAPIKey returns whether Gemini API key is configured
func (s *AIService) HasAPIKey() bool {
	return s.apiKey != ""
}

// AnalysisType defines the type of AI analysis
type AnalysisType string

const (
	AnalysisTypeDailySummary     AnalysisType = "daily_summary"
	AnalysisTypeInvestmentAdvice AnalysisType = "investment_advice"
	AnalysisTypeRiskAssessment   AnalysisType = "risk_assessment"
	AnalysisTypeNewsDigest       AnalysisType = "news_digest"
)

// AIAnalysisResult represents the result of AI analysis
type AIAnalysisResult struct {
	Symbol       string       `json:"symbol"`
	AnalysisType AnalysisType `json:"analysis_type"`
	Content      string       `json:"content"`
	Model        string       `json:"model"`
	InputTokens  int          `json:"input_tokens"`
	OutputTokens int          `json:"output_tokens"`
	CreatedAt    time.Time    `json:"created_at"`
	Cached       bool         `json:"cached"`
}

// StockContext holds all the context data for AI analysis
type StockContext struct {
	Symbol             string
	Name               string
	CurrentPrice       float64
	PriceChange        float64
	PriceChangePercent float64
	Volume             int64
	AvgVolume          int64
	MA5                float64
	MA20               float64
	MA60               float64
	RSI                float64
	MACD               float64
	MACDSignal         float64
	BB_Upper           float64
	BB_Middle          float64
	BB_Lower           float64
	RecentNews         []NewsItem
	SentimentSummary   *SentimentSummary
}

type NewsItem struct {
	Title     string
	Summary   string
	Sentiment string
	Score     float64
	Date      time.Time
}

// GetAnalysis retrieves or generates AI analysis for a symbol
func (s *AIService) GetAnalysis(ctx context.Context, symbol string, analysisType AnalysisType) (*AIAnalysisResult, error) {
	if !s.HasAPIKey() {
		return nil, fmt.Errorf("Gemini API key not configured. Set GEMINI_API_KEY environment variable.")
	}

	today := time.Now().Format("2006-01-02")

	// Check cache first
	cached, err := s.getCachedAnalysis(ctx, symbol, analysisType, today)
	if err == nil && cached != nil {
		cached.Cached = true
		return cached, nil
	}

	// Build context for analysis
	stockContext, err := s.buildStockContext(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to build stock context: %w", err)
	}

	// Generate analysis
	result, err := s.generateAnalysis(ctx, stockContext, analysisType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}

	// Cache the result
	if err := s.cacheAnalysis(ctx, result, today); err != nil {
		// Log but don't fail - caching is not critical
		fmt.Printf("Warning: failed to cache analysis: %v\n", err)
	}

	return result, nil
}

// getCachedAnalysis retrieves cached analysis from database
func (s *AIService) getCachedAnalysis(ctx context.Context, symbol string, analysisType AnalysisType, date string) (*AIAnalysisResult, error) {
	query := `
		SELECT symbol, analysis_type, content, model, COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), created_at
		FROM ai_analysis_cache
		WHERE symbol = $1 AND analysis_type = $2 AND analysis_date = $3
		AND (expires_at IS NULL OR expires_at > NOW())
	`

	var result AIAnalysisResult
	err := s.db.QueryRowContext(ctx, query, symbol, string(analysisType), date).Scan(
		&result.Symbol,
		&result.AnalysisType,
		&result.Content,
		&result.Model,
		&result.InputTokens,
		&result.OutputTokens,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// cacheAnalysis saves analysis to database
func (s *AIService) cacheAnalysis(ctx context.Context, result *AIAnalysisResult, date string) error {
	query := `
		INSERT INTO ai_analysis_cache (symbol, analysis_type, analysis_date, content, model, input_tokens, output_tokens, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW() + INTERVAL '24 hours')
		ON CONFLICT (symbol, analysis_type, analysis_date) 
		DO UPDATE SET content = $4, model = $5, input_tokens = $6, output_tokens = $7, created_at = NOW(), expires_at = NOW() + INTERVAL '24 hours'
	`

	_, err := s.db.ExecContext(ctx, query,
		result.Symbol,
		string(result.AnalysisType),
		date,
		result.Content,
		result.Model,
		result.InputTokens,
		result.OutputTokens,
	)
	return err
}

// buildStockContext gathers all relevant data for AI analysis
func (s *AIService) buildStockContext(ctx context.Context, symbol string) (*StockContext, error) {
	stockContext := &StockContext{
		Symbol: symbol,
	}

	// Get stock info
	stockQuery := `SELECT COALESCE(name, name_en, symbol) FROM taiwan_stocks WHERE symbol = $1`
	s.db.QueryRowContext(ctx, stockQuery, symbol).Scan(&stockContext.Name)

	// Get latest OHLCV data
	ohlcvQuery := `
		SELECT close, volume, 
		       close - LAG(close) OVER (ORDER BY timestamp) as change,
		       CASE WHEN LAG(close) OVER (ORDER BY timestamp) > 0 
		            THEN (close - LAG(close) OVER (ORDER BY timestamp)) / LAG(close) OVER (ORDER BY timestamp) * 100 
		            ELSE 0 END as change_pct
		FROM stock_ohlcv 
		WHERE symbol = $1 
		ORDER BY timestamp DESC 
		LIMIT 2
	`
	rows, err := s.db.QueryContext(ctx, ohlcvQuery, symbol)
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			var change, changePct sql.NullFloat64
			rows.Scan(&stockContext.CurrentPrice, &stockContext.Volume, &change, &changePct)
			if change.Valid {
				stockContext.PriceChange = change.Float64
			}
			if changePct.Valid {
				stockContext.PriceChangePercent = changePct.Float64
			}
		}
	}

	// Get average volume (20 days)
	avgVolQuery := `SELECT COALESCE(AVG(volume), 0) FROM stock_ohlcv WHERE symbol = $1 AND timestamp >= NOW() - INTERVAL '30 days'`
	s.db.QueryRowContext(ctx, avgVolQuery, symbol).Scan(&stockContext.AvgVolume)

	// Get recent news with sentiment
	newsQuery := `
		SELECT title, COALESCE(summary, ''), COALESCE(sentiment, 'neutral'), COALESCE(sentiment_score, 0), published_at
		FROM stock_news
		WHERE symbol = $1
		ORDER BY published_at DESC
		LIMIT 10
	`
	newsRows, err := s.db.QueryContext(ctx, newsQuery, symbol)
	if err == nil {
		defer newsRows.Close()
		for newsRows.Next() {
			var item NewsItem
			newsRows.Scan(&item.Title, &item.Summary, &item.Sentiment, &item.Score, &item.Date)
			stockContext.RecentNews = append(stockContext.RecentNews, item)
		}
	}

	// Get sentiment summary
	sentimentQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN sentiment = 'positive' THEN 1 END) as positive,
			COUNT(CASE WHEN sentiment = 'negative' THEN 1 END) as negative,
			COALESCE(AVG(sentiment_score), 0) as avg_score
		FROM stock_news
		WHERE symbol = $1 AND published_at >= NOW() - INTERVAL '7 days' AND sentiment IS NOT NULL
	`
	var summary SentimentSummary
	summary.Symbol = symbol
	summary.Days = 7
	err = s.db.QueryRowContext(ctx, sentimentQuery, symbol).Scan(
		&summary.TotalArticles,
		&summary.PositiveCount,
		&summary.NegativeCount,
		&summary.AverageScore,
	)
	if err == nil {
		summary.NeutralCount = summary.TotalArticles - summary.PositiveCount - summary.NegativeCount
		if summary.AverageScore > 0.15 {
			summary.OverallSentiment = "positive"
		} else if summary.AverageScore < -0.15 {
			summary.OverallSentiment = "negative"
		} else {
			summary.OverallSentiment = "neutral"
		}
		stockContext.SentimentSummary = &summary
	}

	return stockContext, nil
}

// generateAnalysis calls Gemini API to generate analysis
func (s *AIService) generateAnalysis(ctx context.Context, stockContext *StockContext, analysisType AnalysisType) (*AIAnalysisResult, error) {
	prompt := s.buildPrompt(stockContext, analysisType)
	systemPrompt := s.getSystemPrompt(analysisType)

	// Call Gemini API
	response, err := s.callGemini(ctx, systemPrompt, prompt)
	if err != nil {
		return nil, err
	}

	return &AIAnalysisResult{
		Symbol:       stockContext.Symbol,
		AnalysisType: analysisType,
		Content:      response.Content,
		Model:        s.model,
		InputTokens:  response.InputTokens,
		OutputTokens: response.OutputTokens,
		CreatedAt:    time.Now(),
		Cached:       false,
	}, nil
}

func (s *AIService) getSystemPrompt(analysisType AnalysisType) string {
	base := `你是一位專業的台股分析師，專精於技術分析和基本面分析。你的分析應該：
1. 使用繁體中文回答
2. 客觀專業，避免過度樂觀或悲觀
3. 提供具體的數據支持
4. 考慮台股市場的特性（如漲跌停、交易時間等）
5. 在適當時候提醒投資風險

`
	switch analysisType {
	case AnalysisTypeDailySummary:
		return base + "請提供今日行情摘要，包括價格走勢、成交量變化、和新聞影響。"
	case AnalysisTypeInvestmentAdvice:
		return base + "請提供投資建議，包括短期和中長期觀點，以及建議的操作策略。記得提醒投資風險。"
	case AnalysisTypeRiskAssessment:
		return base + "請進行風險評估，識別潛在的風險因素和需要關注的警訊。"
	case AnalysisTypeNewsDigest:
		return base + "請總結近期新聞對股價的可能影響，分析市場情緒。"
	default:
		return base
	}
}

func (s *AIService) buildPrompt(ctx *StockContext, analysisType AnalysisType) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## 股票資訊\n"))
	sb.WriteString(fmt.Sprintf("- 代碼: %s\n", ctx.Symbol))
	sb.WriteString(fmt.Sprintf("- 名稱: %s\n", ctx.Name))

	if ctx.CurrentPrice > 0 {
		sb.WriteString(fmt.Sprintf("- 當前價格: %.2f\n", ctx.CurrentPrice))
		sb.WriteString(fmt.Sprintf("- 漲跌: %.2f (%.2f%%)\n", ctx.PriceChange, ctx.PriceChangePercent))
	}

	if ctx.Volume > 0 {
		sb.WriteString(fmt.Sprintf("- 成交量: %d\n", ctx.Volume))
		if ctx.AvgVolume > 0 {
			ratio := float64(ctx.Volume) / float64(ctx.AvgVolume) * 100
			sb.WriteString(fmt.Sprintf("- 成交量比 (相對20日均量): %.1f%%\n", ratio))
		}
	}

	// Sentiment summary
	if ctx.SentimentSummary != nil && ctx.SentimentSummary.TotalArticles > 0 {
		sb.WriteString(fmt.Sprintf("\n## 近7日新聞情緒\n"))
		sb.WriteString(fmt.Sprintf("- 總篇數: %d\n", ctx.SentimentSummary.TotalArticles))
		sb.WriteString(fmt.Sprintf("- 正面: %d, 負面: %d, 中性: %d\n",
			ctx.SentimentSummary.PositiveCount,
			ctx.SentimentSummary.NegativeCount,
			ctx.SentimentSummary.NeutralCount))
		sb.WriteString(fmt.Sprintf("- 平均情緒分數: %.2f (範圍 -1 到 1)\n", ctx.SentimentSummary.AverageScore))
		sb.WriteString(fmt.Sprintf("- 整體情緒: %s\n", ctx.SentimentSummary.OverallSentiment))
	}

	// Recent news
	if len(ctx.RecentNews) > 0 {
		sb.WriteString(fmt.Sprintf("\n## 近期新聞\n"))
		for i, news := range ctx.RecentNews {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("%d. [%s] %s (%s)\n", i+1, news.Sentiment, news.Title, news.Date.Format("01/02")))
		}
	}

	// Analysis request
	sb.WriteString(fmt.Sprintf("\n---\n"))
	switch analysisType {
	case AnalysisTypeDailySummary:
		sb.WriteString("請根據以上資訊，提供今日行情摘要分析。")
	case AnalysisTypeInvestmentAdvice:
		sb.WriteString("請根據以上資訊，提供投資建議和操作策略。")
	case AnalysisTypeRiskAssessment:
		sb.WriteString("請根據以上資訊，進行風險評估。")
	case AnalysisTypeNewsDigest:
		sb.WriteString("請根據以上新聞，分析對股價的可能影響。")
	}

	return sb.String()
}

// Gemini API types
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

type geminiResult struct {
	Content      string
	InputTokens  int
	OutputTokens int
}

func (s *AIService) callGemini(ctx context.Context, systemPrompt, userPrompt string) (*geminiResult, error) {
	// Build Gemini API URL
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.model, s.apiKey)

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userPrompt}},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 2048,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s (code: %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	return &geminiResult{
		Content:      geminiResp.Candidates[0].Content.Parts[0].Text,
		InputTokens:  geminiResp.UsageMetadata.PromptTokenCount,
		OutputTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
	}, nil
}

// GetCachedAnalyses returns all cached analyses for a symbol
func (s *AIService) GetCachedAnalyses(ctx context.Context, symbol string, limit int) ([]AIAnalysisResult, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT symbol, analysis_type, content, model, COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), created_at
		FROM ai_analysis_cache
		WHERE symbol = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AIAnalysisResult
	for rows.Next() {
		var r AIAnalysisResult
		if err := rows.Scan(&r.Symbol, &r.AnalysisType, &r.Content, &r.Model, &r.InputTokens, &r.OutputTokens, &r.CreatedAt); err != nil {
			continue
		}
		r.Cached = true
		results = append(results, r)
	}

	return results, nil
}

// ClearCache clears cached analysis for a symbol
func (s *AIService) ClearCache(ctx context.Context, symbol string) error {
	query := `DELETE FROM ai_analysis_cache WHERE symbol = $1`
	_, err := s.db.ExecContext(ctx, query, symbol)
	return err
}
