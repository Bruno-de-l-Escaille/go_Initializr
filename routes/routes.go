package routes

import (
	"go_Initializr/handler/middleware"
	"go_Initializr/internal/injection"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all application routes

func SetupRoutes(router *gin.Engine, container *injection.Container) {
	// Apply global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS())

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is running",
		})
	})

	// Swagger documentation (no auth required)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup entity routes (add more setup functions for other entities)
	setupExampleEntityRoutes(router, container)
}

// setupExampleEntityRoutes configures routes for ExampleEntity
func setupExampleEntityRoutes(router *gin.Engine, container *injection.Container) {
	config := container.GetConfig()
	handler := container.GetExampleEntityHandler()

	// Public API routes (no auth required)
	publicAPI := router.Group("/api/public")
	{
		publicAPI.GET("/example-entities", handler.GetAllEntities)
		publicAPI.GET("/example-entities/:uuid", handler.GetEntity)
	}

	// Protected API routes (JWT auth required)
	protectedAPI := router.Group("/api")
	protectedAPI.Use(middleware.JWTAuth(config.JWTSecret))
	{
		exampleEntities := protectedAPI.Group("/example-entities")
		{
			exampleEntities.POST("", handler.CreateEntity)
			exampleEntities.PUT("/:uuid", handler.UpdateEntity)
			exampleEntities.DELETE("/:uuid", handler.DeleteEntity)
		}
	}
}
