package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/finsights-ai/backend/packages/db"
	"github.com/finsights-ai/backend/packages/dotenv"
	httphandlers "github.com/finsights-ai/backend/packages/http"
	_ "github.com/mattn/go-sqlite3"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	dotenv.Load()

	// Initialize SQLite database
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./screener.db" // Default database path
	}

	dbConn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer dbConn.Close()

	// Test database connection
	if err := dbConn.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.MigrateDatabaseFromFile(dbConn, filepath.Join(".", "packages", "screener", "schema.sql")); err != nil {
		log.Fatal("Migration failed:", err)
	}

	// Insert sample data if database is empty
	if err := db.InsertSampleData(dbConn); err != nil {
		log.Fatal("Failed to insert sample data:", err)
	}

	// Initialize database screener client
	screenerClient := httphandlers.NewDatabaseScreenerClient(dbConn)

	// Setup HTTP handlers
	screenerHandler := httphandlers.NewScreenerHandler(screenerClient)

	// TODO: Only in development: Setup routes with CORS middleware
	http.HandleFunc("/api/screener", corsMiddleware(screenerHandler.GetScreenerData))

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting server on :%s", port)
	log.Println("Using database:", dbPath)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
