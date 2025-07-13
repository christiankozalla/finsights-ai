package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/finsights-ai/backend/packages/dotenv"
	httphandlers "github.com/finsights-ai/backend/packages/http"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dotenv.Load()

	// Initialize SQLite database
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./screener.db" // Default database path
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize database screener client
	screenerClient := httphandlers.NewDatabaseScreenerClient(db)

	// Setup HTTP handlers
	screenerHandler := httphandlers.NewScreenerHandler(screenerClient)

	// Setup routes
	http.HandleFunc("/api/screener", screenerHandler.GetScreenerData)

	// Start server
	log.Println("Starting server on :8080")
	log.Println("Using database:", dbPath)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
