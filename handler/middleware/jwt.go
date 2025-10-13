package middleware

import (
	"crypto/ecdsa"
	jwtutil "go_Initializr/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// JWTAuth returns a middleware that validates JWT tokens using ECDSA public key (ES256)
// This is for external service validation - tokens are signed by gopeople with private key
func JWTAuth(publicKey *ecdsa.PublicKey) gin.HandlerFunc {
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

		// Validate the token using ECDSA public key (ES256)
		token, err := jwtutil.ValidateTokenWithPublicKey(tokenString, publicKey)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate JWT token with ECDSA public key (ES256)")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Extract user_id or user_uuid from claims
		var userUUID string
		if uuid, ok := claims["user_uuid"].(string); ok {
			userUUID = uuid
			log.Info().Str("source", "user_uuid").Str("user_id", userUUID).Msg("Extracted user ID from JWT claims")
		} else if userID, ok := claims["user_id"].(string); ok {
			userUUID = userID
			log.Info().Str("source", "user_id").Str("user_id", userUUID).Msg("Extracted user ID from JWT claims")
		}

		if userUUID == "" {
			log.Error().Interface("claims", claims).Msg("No user_uuid or user_id found in token claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims: missing user identifier"})
			c.Abort()
			return
		}

		// Set user UUID in context for use in handlers
		c.Set("user_uuid", userUUID)
		c.Set("jwt_claims", claims)

		log.Info().Str("user_id", userUUID).Msg("Successfully set user UUID in Gin context")

		// Continue to the next middleware/handler
		c.Next()
	}
}

// OptionalJWTAuth returns a middleware that validates JWT tokens but doesn't require them using ECDSA public key (ES256)
func OptionalJWTAuth(publicKey *ecdsa.PublicKey) gin.HandlerFunc {
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

		// Validate the token using ECDSA public key (ES256)
		token, err := jwtutil.ValidateTokenWithPublicKey(tokenString, publicKey)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to validate JWT token with ES256 in optional auth")
			// Continue without authentication
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			// Continue without authentication
			c.Next()
			return
		}

		// Extract user_id or user_uuid from claims
		var userUUID string
		if uuid, ok := claims["user_uuid"].(string); ok {
			userUUID = uuid
		} else if userID, ok := claims["user_id"].(string); ok {
			userUUID = userID
		}

		if userUUID != "" {
			// Set user UUID in context for use in handlers
			c.Set("user_uuid", userUUID)
			c.Set("jwt_claims", claims)
		}

		// Continue to the next middleware/handler
		c.Next()
	}
}
