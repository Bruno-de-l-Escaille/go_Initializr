package middleware

import (
	"go_Initializr/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// JWTAuth returns a middleware that validates JWT tokens
func JWTAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format. Use 'Bearer <token>'"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := authHeader[7:] // Remove "Bearer " prefix
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			log.Error().Err(err).Msg("Failed to parse JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user UUID in context for use in handlers
		c.Set("user_uuid", claims.UserUUID)
		c.Set("jwt_claims", claims)

		// Continue to the next middleware/handler
		c.Next()
	}
}

// OptionalJWTAuth returns a middleware that validates JWT tokens but doesn't require them
func OptionalJWTAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		// Extract the token
		tokenString := authHeader[7:] // Remove "Bearer " prefix
		if tokenString == "" {
			// No token, continue without authentication
			c.Next()
			return
		}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			log.Warn().Err(err).Msg("Failed to parse JWT token in optional auth")
			// Continue without authentication
			c.Next()
			return
		}

		// Check if token is valid
		if !token.Valid {
			// Continue without authentication
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			// Continue without authentication
			c.Next()
			return
		}

		// Set user UUID in context for use in handlers
		c.Set("user_uuid", claims.UserUUID)
		c.Set("jwt_claims", claims)

		// Continue to the next middleware/handler
		c.Next()
	}
}
