package injection

import (
	"os"
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

	// JWT configuration (validation only - external service)
	JWTPublicKeyPath string

	// MongoDB configuration
	MongoHost     string
	MongoPort     string
	MongoUsername string
	MongoPassword string
	MongoDB       string

	// Redis configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisStreamName   string
	RedisConsumerGroup string
	RedisConsumerName  string


	// Logging configuration
	LogLevel string

	// Environment
	Environment string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*AppConfig, error) {
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

		// JWT configuration (validation only - external service)
		JWTPublicKeyPath: getEnvOrDefault("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),

		// MongoDB configuration
		MongoHost:     getEnvOrDefault("MONGO_HOST", "localhost"),
		MongoPort:     getEnvOrDefault("MONGO_PORT", "27017"),
		MongoUsername: getEnvOrDefault("MONGO_USERNAME", ""),
		MongoPassword: getEnvOrDefault("MONGO_PASSWORD", ""),
		MongoDB:       getEnvOrDefault("MONGO_DB", "goinitializr"),

		// Redis configuration
		RedisHost:          getEnvOrDefault("REDIS_HOST", "localhost"),
		RedisPort:          getEnvOrDefault("REDIS_PORT", "6379"),
		RedisPassword:      getEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:            0, // Default DB
		RedisStreamName:    getEnvOrDefault("REDIS_STREAM_NAME", "go_init_events"),
		RedisConsumerGroup: getEnvOrDefault("REDIS_CONSUMER_GROUP", "go_init_consumers"),
		RedisConsumerName:  getEnvOrDefault("REDIS_CONSUMER_NAME", "consumer_1"),

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
