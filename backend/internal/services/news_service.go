package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"psm-backend/internal/database"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// NewsService handles news fetching and storage
type NewsService struct {
	db *database.DB
}

func NewNewsService(db *database.DB) *NewsService {
	return &NewsService{db: db}
}

// NewsArticle represents a news article
type NewsArticle struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Content      string    `json:"content,omitempty"`
	Source       string    `json:"source"`
	SourceURL    string    `json:"source_url"`
	PublishedAt  time.Time `json:"published_at"`
	FetchedAt    time.Time `json:"fetched_at"`
	Sentiment    string    `json:"sentiment,omitempty"`
	SentimentScore *float64 `json:"sentiment_score,omitempty"`
	Category     string    `json:"category,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
}

// CnyesNewsResponse represents the Cnyes API response (search endpoint)
type CnyesNewsResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Items      struct {
		Total   int             `json:"total"`
		PerPage int             `json:"per_page"`
		Data    []CnyesNewsItem `json:"data"`
	} `json:"items"`
}

// CnyesNewsItem represents a single news item from Cnyes
type CnyesNewsItem struct {
	NewsID    int64    `json:"newsId"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Keyword   []string `json:"keyword"`
	PublishAt int64    `json:"publishAt"` // Unix timestamp
	Payment   int      `json:"payment"`   // 0 = free, 1 = paid
}

// FetchNewsForSymbol fetches news from Cnyes for a specific stock symbol
func (s *NewsService) FetchNewsForSymbol(ctx context.Context, symbol string, limit int) ([]NewsArticle, error) {
	startTime := time.Now()
	
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Cnyes search API - search by stock code
	url := fmt.Sprintf("https://api.cnyes.com/media/api/v1/search?q=%s&limit=%d", symbol, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logFetch("cnyes", &symbol, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.cnyes.com")
	req.Header.Set("Referer", "https://www.cnyes.com/")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logFetch("cnyes", &symbol, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logFetch("cnyes", &symbol, 0, 0, "failed", fmt.Sprintf("HTTP %d", resp.StatusCode), time.Since(startTime))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logFetch("cnyes", &symbol, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cnyesResp CnyesNewsResponse
	if err := json.Unmarshal(body, &cnyesResp); err != nil {
		s.logFetch("cnyes", &symbol, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to NewsArticle
	articles := make([]NewsArticle, 0, len(cnyesResp.Items.Data))
	for _, item := range cnyesResp.Items.Data {
		// Skip paid articles
		if item.Payment == 1 {
			continue
		}

		// Build source URL
		sourceURL := fmt.Sprintf("https://news.cnyes.com/news/id/%d", item.NewsID)

		// Extract summary from content (first 200 chars)
		summary := cleanHTML(item.Content)
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}

		article := NewsArticle{
			Symbol:      symbol,
			Title:       item.Title,
			Summary:     summary,
			Content:     cleanHTML(item.Content),
			Source:      "cnyes",
			SourceURL:   sourceURL,
			PublishedAt: time.Unix(item.PublishAt, 0),
			FetchedAt:   time.Now(),
			Category:    "股票",
			Tags:        item.Keyword,
		}
		articles = append(articles, article)
	}

	// Save to database
	newCount, err := s.SaveArticles(ctx, articles)
	if err != nil {
		s.logFetch("cnyes", &symbol, len(articles), 0, "partial", err.Error(), time.Since(startTime))
		return articles, err
	}

	s.logFetch("cnyes", &symbol, len(articles), newCount, "success", "", time.Since(startTime))
	return articles, nil
}

// FetchGeneralNews fetches general Taiwan stock market news
func (s *NewsService) FetchGeneralNews(ctx context.Context, limit int) ([]NewsArticle, error) {
	startTime := time.Now()
	
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	// Cnyes general Taiwan stock news
	url := fmt.Sprintf("https://api.cnyes.com/media/api/v1/newslist/category/tw_stock?limit=%d", limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logFetch("cnyes", nil, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.cnyes.com")
	req.Header.Set("Referer", "https://www.cnyes.com/")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logFetch("cnyes", nil, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logFetch("cnyes", nil, 0, 0, "failed", fmt.Sprintf("HTTP %d", resp.StatusCode), time.Since(startTime))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logFetch("cnyes", nil, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cnyesResp CnyesNewsResponse
	if err := json.Unmarshal(body, &cnyesResp); err != nil {
		s.logFetch("cnyes", nil, 0, 0, "failed", err.Error(), time.Since(startTime))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to NewsArticle
	articles := make([]NewsArticle, 0, len(cnyesResp.Items.Data))
	for _, item := range cnyesResp.Items.Data {
		// Skip paid articles
		if item.Payment == 1 {
			continue
		}

		// Try to extract stock symbol from title/content
		symbol := extractStockSymbol(item.Title + " " + item.Content)

		sourceURL := fmt.Sprintf("https://news.cnyes.com/news/id/%d", item.NewsID)

		// Extract summary from content (first 200 chars)
		summary := cleanHTML(item.Content)
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}

		article := NewsArticle{
			Symbol:      symbol,
			Title:       item.Title,
			Summary:     summary,
			Content:     cleanHTML(item.Content),
			Source:      "cnyes",
			SourceURL:   sourceURL,
			PublishedAt: time.Unix(item.PublishAt, 0),
			FetchedAt:   time.Now(),
			Category:    "台股",
			Tags:        item.Keyword,
		}
		articles = append(articles, article)
	}

	newCount, err := s.SaveArticles(ctx, articles)
	if err != nil {
		s.logFetch("cnyes", nil, len(articles), 0, "partial", err.Error(), time.Since(startTime))
		return articles, err
	}

	s.logFetch("cnyes", nil, len(articles), newCount, "success", "", time.Since(startTime))
	return articles, nil
}

// SaveArticles saves articles to database, returning count of new articles
func (s *NewsService) SaveArticles(ctx context.Context, articles []NewsArticle) (int, error) {
	if len(articles) == 0 {
		return 0, nil
	}

	newCount := 0
	for _, article := range articles {
		id := uuid.New().String()
		
		// Use INSERT ... ON CONFLICT
		query := `
			INSERT INTO stock_news (id, symbol, title, summary, content, source, source_url, published_at, fetched_at, category, tags)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (source, source_url) DO NOTHING
		`
		
		result, err := s.db.ExecContext(ctx,
			query,
			id,
			article.Symbol,
			article.Title,
			article.Summary,
			article.Content,
			article.Source,
			article.SourceURL,
			article.PublishedAt,
			article.FetchedAt,
			article.Category,
			pq.Array(article.Tags),
		)
		if err != nil {
			// Log error but continue with other records
			continue
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			newCount++
		}
	}

	return newCount, nil
}

// GetNewsForSymbol retrieves news for a symbol from database
func (s *NewsService) GetNewsForSymbol(ctx context.Context, symbol string, limit int, offset int) ([]NewsArticle, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, symbol, title, summary, source, source_url, published_at, fetched_at, 
		       sentiment, sentiment_score, category, tags
		FROM stock_news
		WHERE symbol = $1
		ORDER BY published_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []NewsArticle
	for rows.Next() {
		var a NewsArticle
		var sentimentScore *float64

		err := rows.Scan(
			&a.ID, &a.Symbol, &a.Title, &a.Summary, &a.Source, &a.SourceURL,
			&a.PublishedAt, &a.FetchedAt, &a.Sentiment, &sentimentScore, &a.Category, pq.Array(&a.Tags),
		)
		if err != nil {
			continue
		}
		a.SentimentScore = sentimentScore
		articles = append(articles, a)
	}

	return articles, nil
}

// GetRecentNews retrieves recent news across all symbols
func (s *NewsService) GetRecentNews(ctx context.Context, limit int) ([]NewsArticle, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, symbol, title, summary, source, source_url, published_at, fetched_at,
		       sentiment, sentiment_score, category, tags
		FROM stock_news
		ORDER BY published_at DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []NewsArticle
	for rows.Next() {
		var a NewsArticle
		var sentimentScore *float64

		err := rows.Scan(
			&a.ID, &a.Symbol, &a.Title, &a.Summary, &a.Source, &a.SourceURL,
			&a.PublishedAt, &a.FetchedAt, &a.Sentiment, &sentimentScore, &a.Category, pq.Array(&a.Tags),
		)
		if err != nil {
			continue
		}
		a.SentimentScore = sentimentScore
		articles = append(articles, a)
	}

	return articles, nil
}

// logFetch logs a fetch operation
func (s *NewsService) logFetch(source string, symbol *string, found, newCount int, status, errMsg string, duration time.Duration) {
	query := `
		INSERT INTO news_fetch_log (source, symbol, articles_found, articles_new, status, error_message, duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	var sym *string
	if symbol != nil && *symbol != "" {
		sym = symbol
	}
	
	s.db.Exec(query, source, sym, found, newCount, status, errMsg, duration.Milliseconds())
}

// Helper functions

// cleanHTML removes HTML tags from text
func cleanHTML(s string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	
	// Decode common HTML entities
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	
	// Trim whitespace
	s = strings.TrimSpace(s)
	
	return s
}

// extractStockSymbol tries to extract a Taiwan stock symbol from text
func extractStockSymbol(text string) string {
	// Pattern: 4-digit number followed by optional company name in parentheses
	// Examples: "2330", "台積電(2330)", "2330台積電"
	re := regexp.MustCompile(`\b(\d{4})\b`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
