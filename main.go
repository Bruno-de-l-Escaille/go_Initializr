package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_Initializr/docs"
	"go_Initializr/internal/injection"
	"go_Initializr/internal/logging"
	"go_Initializr/pkg/initializer"
	"go_Initializr/routes"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// @title Go Initializr API
// @version 1.0
// @description This is a sample Go Initializr API server with JWT authentication.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize logger
	logging.InitLogger()

	// Load environment variables
	if err := initializer.LoadEnvVariables(); err != nil {
		log.Fatal().Err(err).Msg("Failed to load environment variables")
	}

	// Initialize database
	db, err := initializer.InitDatabase()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize dependency container
	container, err := injection.NewContainer(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize dependency container")
	}

	// Get configuration from container
	config := container.GetConfig()

	// Configure Swagger documentation dynamically
	swaggerHost := os.Getenv("SWAGGER_HOST")
	if swaggerHost == "" {
		swaggerHost = "localhost:" + config.Port
	}

	// Determine schemes based on environment
	schemes := []string{"http"}
	if config.GinMode == "release" {
		schemes = []string{"https", "http"}
	}

	// Update swagger configuration dynamically
	docs.SwaggerInfo.Host = swaggerHost
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = schemes

	log.Info().Str("host", swaggerHost).Msg("Swagger documentation configured")

	// Set Gin mode
	gin.SetMode(config.GinMode)

	// Create Gin router
	router := gin.New()

	// Setup routes with container (following gopeople DI pattern)
	routes.SetupRoutes(router, container)

	// Get server port from config
	port := config.Port

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting HTTP server on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
