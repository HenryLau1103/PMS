package services

import (
	"context"
	"os"
	"psm-backend/internal/database"
	"regexp"
	"strings"
	"time"
)

// SentimentService handles sentiment analysis for news articles
type SentimentService struct {
	db         *database.DB
	openaiKey  string
}

func NewSentimentService(db *database.DB) *SentimentService {
	return &SentimentService{
		db:        db,
		openaiKey: os.Getenv("OPENAI_API_KEY"),
	}
}

// SentimentResult represents the result of sentiment analysis
type SentimentResult struct {
	Sentiment      string   `json:"sentiment"`       // positive, negative, neutral
	Score          float64  `json:"score"`           // -1.0 to 1.0
	Confidence     float64  `json:"confidence"`      // 0.0 to 1.0
	Keywords       []string `json:"keywords"`        // Key terms that influenced the decision
	Method         string   `json:"method"`          // keyword, openai
}

// Keyword-based sentiment analysis dictionaries (Traditional Chinese financial terms)
var (
	positiveKeywords = []string{
		// 上漲相關
		"漲", "漲停", "大漲", "飆漲", "狂漲", "急漲", "強漲", "噴出", "衝高", "創新高", "歷史新高",
		"突破", "站上", "攻上", "拉升", "走揚", "翻紅", "轉強", "反彈", "回升", "復甦",
		// 正面評價
		"利多", "看好", "看漲", "樂觀", "強勢", "亮眼", "優異", "成長", "獲利", "賺錢",
		"營收", "增加", "提升", "超標", "超預期", "優於預期", "創紀錄", "創高",
		"買進", "加碼", "建議買入", "目標價上調", "評等調升", "推薦",
		// 市場情緒
		"熱絡", "活絡", "爆量", "資金湧入", "外資買超", "投信買超", "法人買超",
		"信心", "期待", "潛力", "機會", "契機", "紅盤", "開紅", "收紅",
		// 技術面
		"黃金交叉", "站穩", "支撐", "底部", "築底", "打底", "翻多",
	}

	negativeKeywords = []string{
		// 下跌相關
		"跌", "跌停", "大跌", "暴跌", "狂跌", "急跌", "重挫", "崩盤", "崩跌", "殺低",
		"破底", "創新低", "跌破", "失守", "摜壓", "下殺", "走跌", "翻黑", "轉弱", "回檔",
		// 負面評價
		"利空", "看壞", "看跌", "悲觀", "弱勢", "衰退", "虧損", "虧錢", "下滑", "下降",
		"減少", "衰減", "不如預期", "低於預期", "遜於預期", "警訊", "警告",
		"賣出", "減碼", "建議賣出", "目標價下調", "評等調降", "降評",
		// 市場情緒
		"觀望", "冷清", "量縮", "資金撤出", "外資賣超", "投信賣超", "法人賣超",
		"擔憂", "恐慌", "風險", "危機", "不確定", "綠盤", "開低", "收黑",
		// 技術面
		"死亡交叉", "跌破", "壓力", "頭部", "做頭", "翻空", "套牢",
		// 負面事件
		"違約", "倒閉", "破產", "調查", "裁員", "停工", "停產", "召回",
	}

	// Intensity modifiers
	intensifiers = map[string]float64{
		"大":  1.5,
		"狂":  1.8,
		"暴":  2.0,
		"急":  1.5,
		"猛":  1.6,
		"強":  1.4,
		"重":  1.5,
		"超":  1.5,
		"極":  1.8,
		"非常": 1.6,
		"相當": 1.3,
		"略":  0.5,
		"微":  0.4,
		"小":  0.5,
		"稍":  0.6,
	}
)

// AnalyzeSentiment performs keyword-based sentiment analysis on text
func (s *SentimentService) AnalyzeSentiment(text string) SentimentResult {
	// Combine title and content for analysis
	text = strings.ToLower(text)
	
	positiveScore := 0.0
	negativeScore := 0.0
	positiveMatches := []string{}
	negativeMatches := []string{}

	// Check for positive keywords
	for _, keyword := range positiveKeywords {
		if strings.Contains(text, keyword) {
			score := 1.0
			// Check for intensifiers
			for intensifier, multiplier := range intensifiers {
				if strings.Contains(text, intensifier+keyword) {
					score *= multiplier
					break
				}
			}
			positiveScore += score
			positiveMatches = append(positiveMatches, keyword)
		}
	}

	// Check for negative keywords
	for _, keyword := range negativeKeywords {
		if strings.Contains(text, keyword) {
			score := 1.0
			// Check for intensifiers
			for intensifier, multiplier := range intensifiers {
				if strings.Contains(text, intensifier+keyword) {
					score *= multiplier
					break
				}
			}
			negativeScore += score
			negativeMatches = append(negativeMatches, keyword)
		}
	}

	// Calculate final sentiment
	totalScore := positiveScore + negativeScore
	var sentiment string
	var score float64
	var confidence float64

	if totalScore == 0 {
		sentiment = "neutral"
		score = 0
		confidence = 0.3 // Low confidence when no keywords found
	} else {
		// Normalize to -1 to 1 range
		score = (positiveScore - negativeScore) / totalScore
		
		// Determine sentiment category
		if score > 0.2 {
			sentiment = "positive"
		} else if score < -0.2 {
			sentiment = "negative"
		} else {
			sentiment = "neutral"
		}
		
		// Calculate confidence based on total matches and score magnitude
		matchCount := len(positiveMatches) + len(negativeMatches)
		confidence = min(0.95, 0.3 + float64(matchCount)*0.1 + abs(score)*0.3)
	}

	// Combine all matched keywords
	allKeywords := append(positiveMatches, negativeMatches...)
	if len(allKeywords) > 10 {
		allKeywords = allKeywords[:10] // Limit to top 10
	}

	return SentimentResult{
		Sentiment:  sentiment,
		Score:      score,
		Confidence: confidence,
		Keywords:   allKeywords,
		Method:     "keyword",
	}
}

// AnalyzeNewsArticle analyzes a single news article and updates the database
func (s *SentimentService) AnalyzeNewsArticle(ctx context.Context, articleID string) (*SentimentResult, error) {
	// Get article from database
	var title, summary, content string
	query := `SELECT title, COALESCE(summary, ''), COALESCE(content, '') FROM stock_news WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, articleID).Scan(&title, &summary, &content)
	if err != nil {
		return nil, err
	}

	// Combine text for analysis (title is weighted more)
	text := title + " " + title + " " + summary + " " + content

	// Perform analysis
	result := s.AnalyzeSentiment(text)

	// Update database
	updateQuery := `
		UPDATE stock_news 
		SET sentiment = $1, sentiment_score = $2, sentiment_analyzed_at = $3
		WHERE id = $4
	`
	_, err = s.db.ExecContext(ctx, updateQuery, result.Sentiment, result.Score, time.Now(), articleID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AnalyzeUnanalyzedNews analyzes all news articles that haven't been analyzed yet
func (s *SentimentService) AnalyzeUnanalyzedNews(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}

	// Get unanalyzed articles
	query := `
		SELECT id, title, COALESCE(summary, ''), COALESCE(content, '')
		FROM stock_news
		WHERE sentiment IS NULL
		ORDER BY published_at DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	analyzedCount := 0
	for rows.Next() {
		var id, title, summary, content string
		if err := rows.Scan(&id, &title, &summary, &content); err != nil {
			continue
		}

		// Combine text for analysis
		text := title + " " + title + " " + summary + " " + content
		result := s.AnalyzeSentiment(text)

		// Update database
		updateQuery := `
			UPDATE stock_news 
			SET sentiment = $1, sentiment_score = $2, sentiment_analyzed_at = $3
			WHERE id = $4
		`
		_, err = s.db.ExecContext(ctx, updateQuery, result.Sentiment, result.Score, time.Now(), id)
		if err != nil {
			continue
		}
		analyzedCount++
	}

	return analyzedCount, nil
}

// GetSentimentSummary returns sentiment summary for a symbol
func (s *SentimentService) GetSentimentSummary(ctx context.Context, symbol string, days int) (*SentimentSummary, error) {
	if days <= 0 {
		days = 7
	}

	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN sentiment = 'positive' THEN 1 END) as positive_count,
			COUNT(CASE WHEN sentiment = 'negative' THEN 1 END) as negative_count,
			COUNT(CASE WHEN sentiment = 'neutral' THEN 1 END) as neutral_count,
			COALESCE(AVG(sentiment_score), 0) as avg_score
		FROM stock_news
		WHERE symbol = $1 
		  AND published_at >= NOW() - INTERVAL '%d days'
		  AND sentiment IS NOT NULL
	`
	query = strings.Replace(query, "%d", string(rune('0'+days)), 1)
	// Fix: use proper formatting
	query = `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN sentiment = 'positive' THEN 1 END) as positive_count,
			COUNT(CASE WHEN sentiment = 'negative' THEN 1 END) as negative_count,
			COUNT(CASE WHEN sentiment = 'neutral' THEN 1 END) as neutral_count,
			COALESCE(AVG(sentiment_score), 0) as avg_score
		FROM stock_news
		WHERE symbol = $1 
		  AND published_at >= NOW() - $2::interval
		  AND sentiment IS NOT NULL
	`

	var summary SentimentSummary
	interval := time.Duration(days) * 24 * time.Hour
	intervalStr := interval.String()
	
	err := s.db.QueryRowContext(ctx, query, symbol, intervalStr).Scan(
		&summary.TotalArticles,
		&summary.PositiveCount,
		&summary.NegativeCount,
		&summary.NeutralCount,
		&summary.AverageScore,
	)
	if err != nil {
		return nil, err
	}

	summary.Symbol = symbol
	summary.Days = days

	// Calculate overall sentiment
	if summary.TotalArticles == 0 {
		summary.OverallSentiment = "unknown"
	} else if summary.AverageScore > 0.15 {
		summary.OverallSentiment = "positive"
	} else if summary.AverageScore < -0.15 {
		summary.OverallSentiment = "negative"
	} else {
		summary.OverallSentiment = "neutral"
	}

	return &summary, nil
}

// SentimentSummary represents aggregated sentiment for a symbol
type SentimentSummary struct {
	Symbol           string  `json:"symbol"`
	Days             int     `json:"days"`
	TotalArticles    int     `json:"total_articles"`
	PositiveCount    int     `json:"positive_count"`
	NegativeCount    int     `json:"negative_count"`
	NeutralCount     int     `json:"neutral_count"`
	AverageScore     float64 `json:"average_score"`
	OverallSentiment string  `json:"overall_sentiment"`
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ExtractStockMentions extracts stock symbols mentioned in text
func ExtractStockMentions(text string) []string {
	// Pattern: 4-digit numbers that could be stock codes
	re := regexp.MustCompile(`\b(\d{4})\b`)
	matches := re.FindAllString(text, -1)
	
	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}
	
	return unique
}
