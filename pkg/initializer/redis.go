package initializer

import (
	"fmt"
	"os"
	"strconv"

	"context"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var RedisClient *redis.Client

func ConnectToRedis() {
	// Get Redis configuration from environment variables
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := os.Getenv("REDIS_DB")

	// Parse Redis DB number
	db, err := strconv.Atoi(redisDB)
	if err != nil {
		db = 0 // Default to DB 0 if there's an error
	}

	// Format Redis URL
	redisURL := fmt.Sprintf("%s:%s", redisHost, redisPort)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPassword,
		DB:       db,
	})

	// Test the connection
	ctx := context.Background()
	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	log.Info().Str("redis_url", redisURL).Msg("Connected to Redis")
}

func NewRedisClientWithDB(db int) *redis.Client {
	// Get Redis configuration from environment variables
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Format Redis URL
	redisURL := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPassword,
		DB:       db,
	})

	// Test the connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	log.Info().Str("redis_url", redisURL).Int("db", db).Msg("Connected to Redis")
	return client
}
