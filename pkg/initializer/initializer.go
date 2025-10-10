package initializer

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// LoadEnvVariables loads environment variables from .env file
func LoadEnvVariables() error {
	if err := godotenv.Load(); err != nil {
		// It's okay if .env file doesn't exist in production
		log.Warn().Err(err).Msg("No .env file found, using system environment variables")
	}
	return nil
}

// InitDatabase initializes the database connection and runs migrations
func InitDatabase() (*sql.DB, error) {
	// Get database configuration from environment variables
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "password")
	dbname := getEnvOrDefault("DB_NAME", "go_Initializr")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open database connection")
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Error().Err(err).Msg("Failed to ping database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Info().
		Str("host", host).
		Str("port", port).
		Str("user", user).
		Str("database", dbname).
		Msg("Database connection established")

	return db, nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
