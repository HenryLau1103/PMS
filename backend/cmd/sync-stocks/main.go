package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"psm-backend/internal/database"
	"psm-backend/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Get configuration
	databaseURL := getEnv("DATABASE_URL", "postgres://psm_user:psm_password@localhost:5432/portfolio_db?sslmode=disable&client_encoding=UTF8")

	// Connect to database
	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create sync service
	syncService := services.NewStockSyncService(db)

	fmt.Println("🔄 Starting stock synchronization from TWSE/TPEx Open API...")
	fmt.Println("")

	// Sync all stocks
	result, err := syncService.SyncAll(context.Background())
	if err != nil {
		log.Fatalf("❌ Synchronization failed: %v", err)
	}

	// Display results
	fmt.Println("✅ Synchronization completed successfully!")
	fmt.Println("")
	fmt.Printf("📊 Results:\n")
	fmt.Printf("   - TSE (上市): %d stocks\n", result["tse"])
	fmt.Printf("   - OTC (上櫃): %d stocks\n", result["otc"])
	fmt.Printf("   - Total:      %d stocks\n", result["total"])
	fmt.Println("")
	fmt.Println("✨ Stock database is now up to date!")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
