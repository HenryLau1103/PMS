package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"psm-backend/internal/database"

	"github.com/shopspring/decimal"
)

// RealtimeService handles real-time stock data fetching
type RealtimeService struct {
	db          *database.DB
	subscribers map[string]map[chan *RealtimeQuote]bool // symbol -> channels
	mu          sync.RWMutex
	stopChan    chan struct{}
	isRunning   bool
}

// RealtimeQuote represents a real-time stock quote
type RealtimeQuote struct {
	Symbol        string          `json:"symbol"`
	Name          string          `json:"name"`
	Price         decimal.Decimal `json:"price"`
	Change        decimal.Decimal `json:"change"`
	ChangePercent decimal.Decimal `json:"change_percent"`
	Open          decimal.Decimal `json:"open"`
	High          decimal.Decimal `json:"high"`
	Low           decimal.Decimal `json:"low"`
	PrevClose     decimal.Decimal `json:"prev_close"`
	Volume        int64           `json:"volume"`
	Turnover      decimal.Decimal `json:"turnover"`
	BidPrice      decimal.Decimal `json:"bid_price"`
	AskPrice      decimal.Decimal `json:"ask_price"`
	BidVolume     int64           `json:"bid_volume"`
	AskVolume     int64           `json:"ask_volume"`
	TradeTime     time.Time       `json:"trade_time"`
	IsOpen        bool            `json:"is_open"`
	LimitUp       decimal.Decimal `json:"limit_up"`
	LimitDown     decimal.Decimal `json:"limit_down"`
	UpdatedAt     time.Time       `json:"updated_at"`
	// 5-level order book
	OrderBook     *OrderBook      `json:"order_book,omitempty"`
}

// OrderBookLevel represents a single price level in the order book
type OrderBookLevel struct {
	Price  decimal.Decimal `json:"price"`
	Volume int64           `json:"volume"`
}

// OrderBook represents the 5-level bid/ask order book
type OrderBook struct {
	Bids []OrderBookLevel `json:"bids"` // Best bids (highest price first)
	Asks []OrderBookLevel `json:"asks"` // Best asks (lowest price first)
}

// MarketStatus represents the current market status
type MarketStatus struct {
	IsOpen       bool      `json:"is_open"`
	Status       string    `json:"status"` // "pre_market", "open", "lunch_break", "closed", "holiday"
	Message      string    `json:"message"`
	NextOpenTime time.Time `json:"next_open_time,omitempty"`
	ServerTime   time.Time `json:"server_time"`
}

func NewRealtimeService(db *database.DB) *RealtimeService {
	return &RealtimeService{
		db:          db,
		subscribers: make(map[string]map[chan *RealtimeQuote]bool),
		stopChan:    make(chan struct{}),
	}
}

// GetMarketStatus returns current Taiwan stock market status
func (s *RealtimeService) GetMarketStatus() *MarketStatus {
	now := time.Now()
	// Convert to Taiwan timezone (UTC+8)
	loc, _ := time.LoadLocation("Asia/Taipei")
	twTime := now.In(loc)
	
	weekday := twTime.Weekday()
	hour := twTime.Hour()
	minute := twTime.Minute()
	timeOfDay := hour*100 + minute

	status := &MarketStatus{
		ServerTime: twTime,
	}

	// Weekend check
	if weekday == time.Saturday || weekday == time.Sunday {
		status.IsOpen = false
		status.Status = "closed"
		status.Message = "休市 - 週末"
		// Calculate next Monday 9:00 AM
		daysUntilMonday := (8 - int(weekday)) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		nextMonday := twTime.AddDate(0, 0, daysUntilMonday)
		status.NextOpenTime = time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 9, 0, 0, 0, loc)
		return status
	}

	// Market hours: 9:00-13:30
	// Pre-market: 8:30-9:00
	// After hours: 13:30-14:30 (block trading)
	switch {
	case timeOfDay < 830:
		status.IsOpen = false
		status.Status = "closed"
		status.Message = "休市 - 尚未開盤"
		status.NextOpenTime = time.Date(twTime.Year(), twTime.Month(), twTime.Day(), 9, 0, 0, 0, loc)
	case timeOfDay >= 830 && timeOfDay < 900:
		status.IsOpen = false
		status.Status = "pre_market"
		status.Message = "盤前試撮 (08:30-09:00)"
		status.NextOpenTime = time.Date(twTime.Year(), twTime.Month(), twTime.Day(), 9, 0, 0, 0, loc)
	case timeOfDay >= 900 && timeOfDay < 1330:
		status.IsOpen = true
		status.Status = "open"
		status.Message = "交易時段 (09:00-13:30)"
	case timeOfDay >= 1330 && timeOfDay < 1430:
		status.IsOpen = false
		status.Status = "after_hours"
		status.Message = "盤後定價交易 (13:30-14:30)"
	default:
		status.IsOpen = false
		status.Status = "closed"
		status.Message = "休市 - 今日交易結束"
		// Next trading day
		nextDay := twTime.AddDate(0, 0, 1)
		if nextDay.Weekday() == time.Saturday {
			nextDay = nextDay.AddDate(0, 0, 2)
		} else if nextDay.Weekday() == time.Sunday {
			nextDay = nextDay.AddDate(0, 0, 1)
		}
		status.NextOpenTime = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 9, 0, 0, 0, loc)
	}

	return status
}

// FetchRealtimeQuote fetches real-time quote from TWSE API
func (s *RealtimeService) FetchRealtimeQuote(ctx context.Context, symbol string) (*RealtimeQuote, error) {
	// TWSE real-time API endpoint
	// Format: https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_2330.tw
	
	// Determine exchange (TSE or OTC)
	exchange := "tse"
	
	// Check if it's OTC stock from database
	var market string
	err := s.db.QueryRow("SELECT market FROM taiwan_stocks WHERE symbol = $1", symbol).Scan(&market)
	if err == nil && market == "OTC" {
		exchange = "otc"
	}

	url := fmt.Sprintf("https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=%s_%s.tw", exchange, symbol)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://mis.twse.com.tw/stock/fibest.jsp")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		MsgArray []struct {
			Symbol    string `json:"c"`     // Stock symbol
			Name      string `json:"n"`     // Stock name
			Price     string `json:"z"`     // Current price
			Open      string `json:"o"`     // Open price
			High      string `json:"h"`     // High price
			Low       string `json:"l"`     // Low price
			PrevClose string `json:"y"`     // Previous close
			Volume    string `json:"v"`     // Volume (張)
			Turnover  string `json:"tv"`    // Trade value
			BidPrice  string `json:"b"`     // Best bid price (comma separated)
			AskPrice  string `json:"a"`     // Best ask price (comma separated)
			BidVolume string `json:"g"`     // Best bid volume (comma separated)
			AskVolume string `json:"f"`     // Best ask volume (comma separated)
			TradeTime string `json:"t"`     // Trade time (HH:MM:SS)
			LimitUp   string `json:"u"`     // Limit up price
			LimitDown string `json:"w"`     // Limit down price
		} `json:"msgArray"`
		QueryTime struct {
			SysTime string `json:"sysTime"`
		} `json:"queryTime"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(result.MsgArray) == 0 {
		return nil, fmt.Errorf("no data found for symbol %s", symbol)
	}

	data := result.MsgArray[0]
	
	quote := &RealtimeQuote{
		Symbol:    data.Symbol,
		Name:      data.Name,
		UpdatedAt: time.Now(),
		IsOpen:    s.GetMarketStatus().IsOpen,
	}

	// Parse prices (handle "-" for no trade)
	if data.Price != "" && data.Price != "-" {
		quote.Price, _ = decimal.NewFromString(data.Price)
	}
	if data.Open != "" && data.Open != "-" {
		quote.Open, _ = decimal.NewFromString(data.Open)
	}
	if data.High != "" && data.High != "-" {
		quote.High, _ = decimal.NewFromString(data.High)
	}
	if data.Low != "" && data.Low != "-" {
		quote.Low, _ = decimal.NewFromString(data.Low)
	}
	if data.PrevClose != "" && data.PrevClose != "-" {
		quote.PrevClose, _ = decimal.NewFromString(data.PrevClose)
	}
	if data.LimitUp != "" && data.LimitUp != "-" {
		quote.LimitUp, _ = decimal.NewFromString(data.LimitUp)
	}
	if data.LimitDown != "" && data.LimitDown != "-" {
		quote.LimitDown, _ = decimal.NewFromString(data.LimitDown)
	}

	// Parse volume (in 張 = 1000 shares)
	if data.Volume != "" && data.Volume != "-" {
		var vol int64
		fmt.Sscanf(data.Volume, "%d", &vol)
		quote.Volume = vol * 1000
	}

	// Parse best bid/ask (first in comma-separated list)
	if data.BidPrice != "" && data.BidPrice != "-" {
		bids := strings.Split(data.BidPrice, "_")
		if len(bids) > 0 && bids[0] != "" {
			quote.BidPrice, _ = decimal.NewFromString(bids[0])
		}
	}
	if data.AskPrice != "" && data.AskPrice != "-" {
		asks := strings.Split(data.AskPrice, "_")
		if len(asks) > 0 && asks[0] != "" {
			quote.AskPrice, _ = decimal.NewFromString(asks[0])
		}
	}

	// Parse 5-level order book
	quote.OrderBook = parseOrderBook(data.BidPrice, data.AskPrice, data.BidVolume, data.AskVolume)

	// Calculate change
	if !quote.Price.IsZero() && !quote.PrevClose.IsZero() {
		quote.Change = quote.Price.Sub(quote.PrevClose)
		quote.ChangePercent = quote.Change.Div(quote.PrevClose).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return quote, nil
}

// FetchMultipleQuotes fetches real-time quotes for multiple symbols
func (s *RealtimeService) FetchMultipleQuotes(ctx context.Context, symbols []string) ([]*RealtimeQuote, error) {
	if len(symbols) == 0 {
		return []*RealtimeQuote{}, nil
	}

	// Build ex_ch parameter for multiple stocks
	var exChParts []string
	for _, symbol := range symbols {
		// Check market
		var market string
		exchange := "tse"
		err := s.db.QueryRow("SELECT market FROM taiwan_stocks WHERE symbol = $1", symbol).Scan(&market)
		if err == nil && market == "OTC" {
			exchange = "otc"
		}
		exChParts = append(exChParts, fmt.Sprintf("%s_%s.tw", exchange, symbol))
	}

	url := fmt.Sprintf("https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=%s", strings.Join(exChParts, "|"))
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://mis.twse.com.tw/stock/fibest.jsp")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		MsgArray []struct {
			Symbol    string `json:"c"`
			Name      string `json:"n"`
			Price     string `json:"z"`
			Open      string `json:"o"`
			High      string `json:"h"`
			Low       string `json:"l"`
			PrevClose string `json:"y"`
			Volume    string `json:"v"`
			BidPrice  string `json:"b"`
			AskPrice  string `json:"a"`
			LimitUp   string `json:"u"`
			LimitDown string `json:"w"`
		} `json:"msgArray"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	marketStatus := s.GetMarketStatus()
	var quotes []*RealtimeQuote
	
	for _, data := range result.MsgArray {
		quote := &RealtimeQuote{
			Symbol:    data.Symbol,
			Name:      data.Name,
			UpdatedAt: time.Now(),
			IsOpen:    marketStatus.IsOpen,
		}

		if data.Price != "" && data.Price != "-" {
			quote.Price, _ = decimal.NewFromString(data.Price)
		}
		if data.PrevClose != "" && data.PrevClose != "-" {
			quote.PrevClose, _ = decimal.NewFromString(data.PrevClose)
		}
		if data.Open != "" && data.Open != "-" {
			quote.Open, _ = decimal.NewFromString(data.Open)
		}
		if data.High != "" && data.High != "-" {
			quote.High, _ = decimal.NewFromString(data.High)
		}
		if data.Low != "" && data.Low != "-" {
			quote.Low, _ = decimal.NewFromString(data.Low)
		}
		if data.LimitUp != "" && data.LimitUp != "-" {
			quote.LimitUp, _ = decimal.NewFromString(data.LimitUp)
		}
		if data.LimitDown != "" && data.LimitDown != "-" {
			quote.LimitDown, _ = decimal.NewFromString(data.LimitDown)
		}

		if data.Volume != "" && data.Volume != "-" {
			var vol int64
			fmt.Sscanf(data.Volume, "%d", &vol)
			quote.Volume = vol * 1000
		}

		if data.BidPrice != "" && data.BidPrice != "-" {
			bids := strings.Split(data.BidPrice, "_")
			if len(bids) > 0 && bids[0] != "" {
				quote.BidPrice, _ = decimal.NewFromString(bids[0])
			}
		}
		if data.AskPrice != "" && data.AskPrice != "-" {
			asks := strings.Split(data.AskPrice, "_")
			if len(asks) > 0 && asks[0] != "" {
				quote.AskPrice, _ = decimal.NewFromString(asks[0])
			}
		}

		if !quote.Price.IsZero() && !quote.PrevClose.IsZero() {
			quote.Change = quote.Price.Sub(quote.PrevClose)
			quote.ChangePercent = quote.Change.Div(quote.PrevClose).Mul(decimal.NewFromInt(100)).Round(2)
		}

		quotes = append(quotes, quote)
	}

	return quotes, nil
}

// Subscribe adds a subscriber for real-time updates on a symbol
func (s *RealtimeService) Subscribe(symbol string, ch chan *RealtimeQuote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subscribers[symbol] == nil {
		s.subscribers[symbol] = make(map[chan *RealtimeQuote]bool)
	}
	s.subscribers[symbol][ch] = true
}

// Unsubscribe removes a subscriber
func (s *RealtimeService) Unsubscribe(symbol string, ch chan *RealtimeQuote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subscribers[symbol] != nil {
		delete(s.subscribers[symbol], ch)
		if len(s.subscribers[symbol]) == 0 {
			delete(s.subscribers, symbol)
		}
	}
}

// GetSubscribedSymbols returns all symbols with active subscribers
func (s *RealtimeService) GetSubscribedSymbols() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var symbols []string
	for symbol := range s.subscribers {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// Broadcast sends quote updates to all subscribers of a symbol
func (s *RealtimeService) Broadcast(quote *RealtimeQuote) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if channels, ok := s.subscribers[quote.Symbol]; ok {
		for ch := range channels {
			select {
			case ch <- quote:
			default:
				// Channel full, skip
			}
		}
	}
}

// parseOrderBook parses the 5-level bid/ask order book from TWSE data
func parseOrderBook(bidPrices, askPrices, bidVolumes, askVolumes string) *OrderBook {
	ob := &OrderBook{
		Bids: make([]OrderBookLevel, 0, 5),
		Asks: make([]OrderBookLevel, 0, 5),
	}

	// Parse bid prices and volumes (underscore separated)
	if bidPrices != "" && bidPrices != "-" && bidVolumes != "" && bidVolumes != "-" {
		prices := strings.Split(bidPrices, "_")
		volumes := strings.Split(bidVolumes, "_")
		
		for i := 0; i < len(prices) && i < len(volumes) && i < 5; i++ {
			if prices[i] == "" || volumes[i] == "" {
				continue
			}
			price, err := decimal.NewFromString(prices[i])
			if err != nil {
				continue
			}
			var vol int64
			fmt.Sscanf(volumes[i], "%d", &vol)
			ob.Bids = append(ob.Bids, OrderBookLevel{
				Price:  price,
				Volume: vol,
			})
		}
	}

	// Parse ask prices and volumes
	if askPrices != "" && askPrices != "-" && askVolumes != "" && askVolumes != "-" {
		prices := strings.Split(askPrices, "_")
		volumes := strings.Split(askVolumes, "_")
		
		for i := 0; i < len(prices) && i < len(volumes) && i < 5; i++ {
			if prices[i] == "" || volumes[i] == "" {
				continue
			}
			price, err := decimal.NewFromString(prices[i])
			if err != nil {
				continue
			}
			var vol int64
			fmt.Sscanf(volumes[i], "%d", &vol)
			ob.Asks = append(ob.Asks, OrderBookLevel{
				Price:  price,
				Volume: vol,
			})
		}
	}

	if len(ob.Bids) == 0 && len(ob.Asks) == 0 {
		return nil
	}

	return ob
}
