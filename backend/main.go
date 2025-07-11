package main

import (
	"log"
	"net/http"
	"os"

	"github.com/finsights-ai/backend/packages/dotenv"
	"github.com/finsights-ai/backend/packages/eodhd"
	httphandlers "github.com/finsights-ai/backend/packages/http"
)

func main() {
	dotenv.Load()

	EOD_HISTORICAL_DATA_API_SECRET := os.Getenv("EOD_HISTORICAL_DATA_API_SECRET")
	if EOD_HISTORICAL_DATA_API_SECRET == "" {
		log.Fatal("EOD_HISTORICAL_DATA_API_SECRET environment variable is required")
	}

	// Initialize EODHD client
	client, err := eodhd.NewClient(EOD_HISTORICAL_DATA_API_SECRET, "./eodhd-cache")
	if err != nil {
		log.Fatal("Failed to initialize EODHD client:", err)
	}

	// Setup HTTP handlers
	screenerHandler := httphandlers.NewScreenerHandler(client)

	// Setup routes
	http.HandleFunc("/api/screener", screenerHandler.GetScreenerData)

	// Start server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
