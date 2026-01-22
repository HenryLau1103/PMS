package main

import (
	"log"
	"os"
	"psm-backend/internal/database"
	"psm-backend/internal/handlers"
	"psm-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Get configuration
	databaseURL := getEnv("DATABASE_URL", "postgres://psm_user:psm_password@localhost:5432/portfolio_db?sslmode=disable&client_encoding=UTF8")
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	port := getEnv("PORT", "8080")

	// Connect to database
	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Connect to Redis (optional, gracefully skip if unavailable)
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	defer redisClient.Close()

	// Initialize services
	ledgerService := services.NewLedgerService(db)
	stockService := services.NewStockService(db)
	stockSyncService := services.NewStockSyncService(db)
	marketDataService := services.NewMarketDataService(db)
	taService := services.NewTechnicalAnalysisService(db, redisClient)

	// Initialize handlers
	ledgerHandler := handlers.NewLedgerHandler(ledgerService)
	stockHandler := handlers.NewStockHandler(stockService)
	stockSyncHandler := handlers.NewStockSyncHandler(stockSyncService)
	marketDataHandler := handlers.NewMarketDataHandler(marketDataService)
	indicatorHandler := handlers.NewIndicatorHandler(taService)
	bulkSyncHandler := handlers.NewBulkSyncHandler(marketDataService, db)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "PSM Backend API",
		ServerHeader: "PSM",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Health(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})

	// API v1 routes
	api := app.Group("/api/v1")

	// Transaction/Event routes
	api.Post("/events", ledgerHandler.CreateEvent)
	api.Get("/portfolios/:portfolio_id/events", ledgerHandler.GetEvents)
	api.Get("/portfolios/:portfolio_id/events/:symbol", ledgerHandler.GetEventsBySymbol)

	// Position routes
	api.Get("/portfolios/:portfolio_id/positions", ledgerHandler.GetPositions)
	api.Get("/portfolios/:portfolio_id/positions/:symbol", ledgerHandler.GetPosition)
	api.Get("/portfolios/:portfolio_id/positions/:symbol/pnl", ledgerHandler.CalculateUnrealizedPnL)

	// Portfolio routes
	api.Get("/portfolios/:portfolio_id", ledgerHandler.GetPortfolio)
	api.Get("/portfolios", ledgerHandler.GetUserPortfolios)

	// Stock routes
	api.Get("/stocks/search", stockHandler.SearchStocks)
	api.Get("/stocks/:symbol", stockHandler.GetStock)
	api.Post("/stocks/sync", stockSyncHandler.SyncStocks)

	// Market data routes (Phase 2.1)
	api.Get("/stocks/:symbol/ohlcv", marketDataHandler.GetOHLCV)
	api.Post("/market/sync", marketDataHandler.SyncMarketData)
	api.Post("/market/refresh-aggregates", marketDataHandler.RefreshAggregates)

	// Technical indicator routes (Phase 2.2)
	api.Get("/indicators/:symbol/ma", indicatorHandler.GetMA)
	api.Get("/indicators/:symbol/rsi", indicatorHandler.GetRSI)
	api.Get("/indicators/:symbol/macd", indicatorHandler.GetMACD)
	api.Get("/indicators/:symbol/bb", indicatorHandler.GetBollingerBands)
	api.Get("/indicators/:symbol/kdj", indicatorHandler.GetKDJ)
	api.Post("/indicators/:symbol/batch", indicatorHandler.GetBatchIndicators)

	// Bulk sync routes (Phase 2.5)
	api.Get("/market/bulk-sync/status", bulkSyncHandler.GetSyncStatus)
	api.Post("/market/bulk-sync/start", bulkSyncHandler.StartBulkSync)
	api.Post("/market/bulk-sync/stop", bulkSyncHandler.StopBulkSync)

	// Start server
	log.Printf("🚀 Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
