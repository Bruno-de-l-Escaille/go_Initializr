package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the zerolog logger with appropriate configuration
func InitLogger() {
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	env := os.Getenv("APP_ENV")

	// Default to info level if not set or invalid
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil || logLevel == "" {
		level = zerolog.InfoLevel
	}

	// Create logs directory if it doesn't exist
	err = os.MkdirAll("var/logs", 0755)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create logs directory")
	}

	// Get current date for filename
	currentDate := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("var/logs/%s.json", currentDate)

	// Clear log file in development/local environment
	var openFlags int
	if env == "development" || env == "local" || env == "" {
		openFlags = os.O_RDWR | os.O_CREATE | os.O_TRUNC // Truncate (clear) the file
	} else {
		openFlags = os.O_RDWR | os.O_CREATE | os.O_APPEND // Append in production
	}

	logFile, err := os.OpenFile(logFileName, openFlags, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}

	jsonLogWriter := NewJSONLogWriter(logFile)

	// Only write to file, no console output
	log.Logger = zerolog.New(jsonLogWriter).With().
		Timestamp().
		Logger()

	zerolog.SetGlobalLevel(level)
	log.Info().Msgf("Logger configured with level '%s' for '%s' environment", level, env)
}
