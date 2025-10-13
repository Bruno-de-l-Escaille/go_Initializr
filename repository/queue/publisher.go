package events

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// EventPublisher is responsible for publishing events
type EventPublisher interface {
	Publish(event Event) error
}

// RedisEventPublisher implements EventPublisher using Redis
type RedisEventPublisher struct {
	redisClient *redis.Client
	streamName  string
}

// NewRedisEventPublisher creates a new Redis event publisher
func NewRedisEventPublisher(client *redis.Client, stream string) EventPublisher {
	return &RedisEventPublisher{
		redisClient: client,
		streamName:  stream,
	}
}

// Publish publishes an event to Redis stream
func (p *RedisEventPublisher) Publish(event Event) error {
	// Create values for the stream entry
	// Only store UUID and entity type in the stream as requested
	values := map[string]interface{}{
		"uuid":        event.GetUUID(),
		"entity_uuid": event.GetEntityUUID(),
		"type":        event.GetType(),
	}

	ctx := context.Background()

	// Add to Redis stream
	_, err := p.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		Values: values,
	}).Result()

	if err != nil {
		log.Error().Err(err).Msg("Failed to publish event to stream")
		return err
	}

	return nil
}
