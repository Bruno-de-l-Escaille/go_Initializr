package injection

import (
	"os"
	"strconv"
)

// AppConfig holds application configuration values derived from environment variables.
type AppConfig struct {
	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Server configuration
	Port    string
	GinMode string

	// JWT configuration
	JWTSecret      string
	JWTExpiryHours int

	// Logging configuration
	LogLevel string

	// Environment
	Environment string
}

// loadConfig loads configuration from environment variables.
func loadConfig() (*AppConfig, error) {
	jwtExpiryHoursStr := os.Getenv("JWT_EXPIRY_HOURS")
	jwtExpiryHours, err := strconv.Atoi(jwtExpiryHoursStr)
	if err != nil {
		jwtExpiryHours = 24 // Default to 24 hours
	}

	return &AppConfig{
		// Database configuration
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "postgres"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "password"),
		DBName:     getEnvOrDefault("DB_NAME", "go_Initializr"),
		DBSSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),

		// Server configuration
		Port:    getEnvOrDefault("PORT", "8080"),
		GinMode: getEnvOrDefault("GIN_MODE", "debug"),

		// JWT configuration
		JWTSecret:      getEnvOrDefault("JWT_SECRET", "your-secret-key"),
		JWTExpiryHours: jwtExpiryHours,

		// Logging configuration
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),

		// Environment
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
	}, nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
