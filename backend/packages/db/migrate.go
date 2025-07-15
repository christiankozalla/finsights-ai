package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func MigrateDatabaseFromFile(db *sql.DB, schemaFilePath string) error {
	log.Println("Running database migrations from schema.sql...")

	schemaBytes, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	schemaSQL := string(schemaBytes)

	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
