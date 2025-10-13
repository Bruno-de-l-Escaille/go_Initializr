package injection

import (
	"database/sql"
	"go_Initializr/pkg/nosql"
	"go_Initializr/repository"
	"go_Initializr/pkg/initializer"
	events "go_Initializr/repository/queue"
	"go_Initializr/service"
	"go_Initializr/service/event"
	"github.com/redis/go-redis/v9"
	"crypto/ecdsa"
	jwtutil "go_Initializr/pkg/jwt"
	"fmt"
	"github.com/rs/zerolog/log"
)

// coreComponents holds fundamental shared dependencies.
type coreComponents struct {
	config          *AppConfig // Assumes AppConfig is defined in config.go
	db              *sql.DB
	mongoClient     *nosql.MongoDB
	redisClient     *redis.Client
	baseRepository  *repository.BaseRepository
	eventRepository repository.EventRepositoryInterface
	eventPublisher  events.EventPublisher
	eventSubscriber events.EventSubscriber
	eventService    service.EventServiceInterface
	publicKey      *ecdsa.PublicKey
}

// initializeCoreComponents creates the essential shared components.
func initializeCoreComponents(db *sql.DB, mongoClient *nosql.MongoDB) (*coreComponents, error) {
	config, err := LoadConfig() // Assumes LoadConfig is defined in config.go
	if err != nil {
		return nil, err // Handle config loading error if necessary
	}
	// Load ECDSA public key for JWT validation (ES256)
	publicKey, err := jwtutil.LoadPublicKeyFromFile(config.JWTPublicKeyPath)
	if err != nil {
		log.Error().Err(err).Str("path", config.JWTPublicKeyPath).Msg("Failed to load ECDSA public key")
		return nil, fmt.Errorf("failed to load ECDSA public key: %w", err)
	}
	log.Info().Str("path", config.JWTPublicKeyPath).Str("algorithm", "ES256").Msg("ECDSA public key loaded successfully")

	// Assuming these are initialized elsewhere and passed in or globally available
	// For clarity, they should ideally be passed into NewContainer
	redisClient := initializer.RedisClient

	baseRepo := &repository.BaseRepository{
		DB:          db,
		MongoClient: mongoClient,
	}

	eventRepo := &repository.EventRepository{DB: db}

	eventPublisher := events.NewRedisEventPublisher(redisClient, config.RedisStreamName)
	eventSubscriber := events.NewRedisEventSubscriber(redisClient, config.RedisStreamName, config.RedisConsumerGroup, config.RedisConsumerName)

	eventService := &event.EventService{
		Repository:       eventRepo,
		EntityRepository: *baseRepo,
		EventPublisher:   eventPublisher,
	}

	return &coreComponents{
		config:          config,
		db:              db,
		mongoClient:     mongoClient,
		redisClient:     redisClient,
		baseRepository:  baseRepo,
		eventRepository: eventRepo,
		eventPublisher:  eventPublisher,
		eventSubscriber: eventSubscriber,
		eventService:    eventService,
		publicKey:      publicKey,
	}, nil
}

// GetPublicKey returns the ECDSA public key for JWT validation (ES256)
func (c *coreComponents) GetPublicKey() *ecdsa.PublicKey {
	return c.publicKey
}
