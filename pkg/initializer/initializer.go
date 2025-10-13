package initializer

import (
	"database/sql"
	"fmt"
	"os"

	"go_Initializr/pkg/nosql"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"time"
	"go_Initializr/models/apperrors"
)

var MongoClient *nosql.MongoDB

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

// InitMongoDB initializes the MongoDB connection
func ConnectionToMongoDB() error {
	mongoHost := os.Getenv("MONGO_HOST")
	mongoPort := os.Getenv("MONGO_PORT")
	mongoUsername := os.Getenv("MONGO_USERNAME")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDatabase := os.Getenv("MONGO_DB")

	// Build MongoDB URI with optional authentication
	var mongoURI string
	if mongoUsername != "" && mongoPassword != "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s", mongoUsername, mongoPassword, mongoHost, mongoPort)
	} else {
		mongoURI = fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)
	}

	mongoConfig := nosql.MongoDBConfig{
		URI:      mongoURI,
		Database: mongoDatabase,
		Username: mongoUsername,
		Password: mongoPassword,
		Timeout:  10 * time.Second,
	}

	client, err := nosql.NewMongoDB(mongoConfig)
	if err != nil {
		return apperrors.NewServiceUnavailable()
	}

	MongoClient = client
	return nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
