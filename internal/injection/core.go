package injection

import (
	"crypto/ecdsa"
	"database/sql"
	"fmt"
	jwtutil "go_Initializr/pkg/jwt"
	"go_Initializr/repository"

	"github.com/rs/zerolog/log"
)

// coreComponents holds fundamental shared dependencies.
type coreComponents struct {
	config         *AppConfig
	db             *sql.DB
	baseRepository *repository.BaseRepository
	publicKey      *ecdsa.PublicKey
}

// initializeCoreComponents creates the essential shared components.
func initializeCoreComponents(db *sql.DB) (*coreComponents, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	// Load ECDSA public key for JWT validation (ES256)
	publicKey, err := jwtutil.LoadPublicKeyFromFile(config.JWTPublicKeyPath)
	if err != nil {
		log.Error().Err(err).Str("path", config.JWTPublicKeyPath).Msg("Failed to load ECDSA public key")
		return nil, fmt.Errorf("failed to load ECDSA public key: %w", err)
	}
	log.Info().Str("path", config.JWTPublicKeyPath).Str("algorithm", "ES256").Msg("ECDSA public key loaded successfully")

	baseRepo := &repository.BaseRepository{
		DB: db,
	}

	return &coreComponents{
		config:         config,
		db:             db,
		baseRepository: baseRepo,
		publicKey:      publicKey,
	}, nil
}

// GetConfig returns the application configuration
func (c *coreComponents) GetConfig() *AppConfig {
	return c.config
}

// GetDB returns the database connection
func (c *coreComponents) GetDB() *sql.DB {
	return c.db
}

// GetBaseRepository returns the base repository
func (c *coreComponents) GetBaseRepository() *repository.BaseRepository {
	return c.baseRepository
}

// GetPublicKey returns the ECDSA public key for JWT validation (ES256)
func (c *coreComponents) GetPublicKey() *ecdsa.PublicKey {
	return c.publicKey
}
