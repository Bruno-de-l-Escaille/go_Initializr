package nosql

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConfig contains the configuration for MongoDB
type MongoDBConfig struct {
	URI      string
	Database string
	Username string
	Password string
	Timeout  time.Duration
}

// MongoDB struct to encapsulate MongoDB connection
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoDB initializes a connection to MongoDB
func NewMongoDB(config MongoDBConfig) (*MongoDB, error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Connection options
	clientOptions := options.Client().ApplyURI(config.URI)

	// Add authentication if username and password are provided
	if config.Username != "" && config.Password != "" {
		credential := options.Credential{
			Username: config.Username,
			Password: config.Password,
		}
		clientOptions.SetAuth(credential)
	}

	// Connect to MongoDB client
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	// Get database reference
	database := client.Database(config.Database)

	log.Info().Str("database", config.Database).Msg("Successfully connected to MongoDB")

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.Client.Disconnect(ctx)
}

// GetCollection returns a collection from the database
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
