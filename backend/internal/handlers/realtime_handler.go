package handlers

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"psm-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// RealtimeHandler handles real-time stock data endpoints
type RealtimeHandler struct {
	realtimeService *services.RealtimeService
	clients         map[*websocket.Conn]*clientInfo
	mu              sync.RWMutex
}

type clientInfo struct {
	symbols map[string]bool
}

// WebSocket message types
type wsMessage struct {
	Action  string   `json:"action"`  // "subscribe", "unsubscribe"
	Symbols []string `json:"symbols"` // Stock symbols
}

type wsResponse struct {
	Type    string      `json:"type"` // "quote", "status", "error", "subscribed"
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

func NewRealtimeHandler(realtimeService *services.RealtimeService) *RealtimeHandler {
	h := &RealtimeHandler{
		realtimeService: realtimeService,
		clients:         make(map[*websocket.Conn]*clientInfo),
	}
	
	// Start the quote broadcaster
	go h.runQuoteBroadcaster()
	
	return h
}

// GetMarketStatus returns current market status
func (h *RealtimeHandler) GetMarketStatus(c *fiber.Ctx) error {
	status := h.realtimeService.GetMarketStatus()
	return c.JSON(fiber.Map{
		"success": true,
		"data":    status,
	})
}

// GetRealtimeQuote returns real-time quote for a single stock
func (h *RealtimeHandler) GetRealtimeQuote(c *fiber.Ctx) error {
	symbol := strings.ToUpper(c.Params("symbol"))
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Symbol is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	quote, err := h.realtimeService.FetchRealtimeQuote(ctx, symbol)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    quote,
	})
}

// GetBatchQuotes returns real-time quotes for multiple stocks
func (h *RealtimeHandler) GetBatchQuotes(c *fiber.Ctx) error {
	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Symbols parameter is required",
		})
	}

	symbols := strings.Split(strings.ToUpper(symbolsParam), ",")
	if len(symbols) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "At least one symbol is required",
		})
	}

	// Limit to 20 symbols per request
	if len(symbols) > 20 {
		symbols = symbols[:20]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	quotes, err := h.realtimeService.FetchMultipleQuotes(ctx, symbols)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    quotes,
		"count":   len(quotes),
	})
}

// WebSocketUpgrade middleware to allow WebSocket connections
func (h *RealtimeHandler) WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// HandleWebSocket handles WebSocket connections for real-time updates
func (h *RealtimeHandler) HandleWebSocket(c *websocket.Conn) {
	// Register client
	h.mu.Lock()
	h.clients[c] = &clientInfo{
		symbols: make(map[string]bool),
	}
	h.mu.Unlock()

	log.Printf("WebSocket client connected: %s", c.RemoteAddr())

	// Send initial market status
	status := h.realtimeService.GetMarketStatus()
	h.sendMessage(c, wsResponse{
		Type: "status",
		Data: status,
	})

	defer func() {
		// Unregister client
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		c.Close()
		log.Printf("WebSocket client disconnected: %s", c.RemoteAddr())
	}()

	// Handle incoming messages
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg wsMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			h.sendMessage(c, wsResponse{
				Type:    "error",
				Message: "Invalid message format",
			})
			continue
		}

		switch wsMsg.Action {
		case "subscribe":
			h.handleSubscribe(c, wsMsg.Symbols)
		case "unsubscribe":
			h.handleUnsubscribe(c, wsMsg.Symbols)
		default:
			h.sendMessage(c, wsResponse{
				Type:    "error",
				Message: "Unknown action: " + wsMsg.Action,
			})
		}
	}
}

func (h *RealtimeHandler) handleSubscribe(c *websocket.Conn, symbols []string) {
	h.mu.Lock()
	client, ok := h.clients[c]
	if !ok {
		h.mu.Unlock()
		return
	}

	for _, symbol := range symbols {
		symbol = strings.ToUpper(symbol)
		client.symbols[symbol] = true
	}
	h.mu.Unlock()

	// Send confirmation
	h.sendMessage(c, wsResponse{
		Type:    "subscribed",
		Data:    symbols,
		Message: "Successfully subscribed to symbols",
	})

	// Immediately fetch and send current quotes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	quotes, err := h.realtimeService.FetchMultipleQuotes(ctx, symbols)
	if err != nil {
		log.Printf("Error fetching initial quotes: %v", err)
		return
	}

	for _, quote := range quotes {
		h.sendMessage(c, wsResponse{
			Type: "quote",
			Data: quote,
		})
	}
}

func (h *RealtimeHandler) handleUnsubscribe(c *websocket.Conn, symbols []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.clients[c]
	if !ok {
		return
	}

	for _, symbol := range symbols {
		symbol = strings.ToUpper(symbol)
		delete(client.symbols, symbol)
	}

	h.sendMessage(c, wsResponse{
		Type:    "unsubscribed",
		Data:    symbols,
		Message: "Successfully unsubscribed from symbols",
	})
}

func (h *RealtimeHandler) sendMessage(c *websocket.Conn, msg wsResponse) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// runQuoteBroadcaster periodically fetches and broadcasts quotes to subscribers
func (h *RealtimeHandler) runQuoteBroadcaster() {
	ticker := time.NewTicker(5 * time.Second) // Update every 5 seconds
	defer ticker.Stop()

	for range ticker.C {
		h.broadcastQuotes()
	}
}

func (h *RealtimeHandler) broadcastQuotes() {
	// Check market status first
	status := h.realtimeService.GetMarketStatus()
	
	// Collect all subscribed symbols
	h.mu.RLock()
	symbolSet := make(map[string]bool)
	for _, client := range h.clients {
		for symbol := range client.symbols {
			symbolSet[symbol] = true
		}
	}
	h.mu.RUnlock()

	if len(symbolSet) == 0 {
		return
	}

	// Convert to slice
	symbols := make([]string, 0, len(symbolSet))
	for symbol := range symbolSet {
		symbols = append(symbols, symbol)
	}

	// Fetch quotes (only during market hours or slightly after for data consistency)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	quotes, err := h.realtimeService.FetchMultipleQuotes(ctx, symbols)
	if err != nil {
		log.Printf("Error fetching quotes for broadcast: %v", err)
		return
	}

	// Create quote map for quick lookup
	quoteMap := make(map[string]*services.RealtimeQuote)
	for _, quote := range quotes {
		quoteMap[quote.Symbol] = quote
	}

	// Broadcast to each client their subscribed symbols
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn, client := range h.clients {
		// Send market status periodically
		h.sendMessage(conn, wsResponse{
			Type: "status",
			Data: status,
		})

		// Send quotes for subscribed symbols
		for symbol := range client.symbols {
			if quote, ok := quoteMap[symbol]; ok {
				h.sendMessage(conn, wsResponse{
					Type: "quote",
					Data: quote,
				})
			}
		}
	}
}
