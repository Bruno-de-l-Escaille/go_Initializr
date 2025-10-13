package events

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// EventHandler handles incoming events
type EventHandler interface {
	HandleEvent(event Event) error
}

// EventSubscriber subscribes to events
type EventSubscriber interface {
	Subscribe(handler EventHandler)
	Close()
}

// RedisEventSubscriber implements EventSubscriber using Redis
type RedisEventSubscriber struct {
	redisClient   *redis.Client
	streamName    string
	consumerGroup string
	consumerName  string
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewRedisEventSubscriber creates a new Redis event subscriber
func NewRedisEventSubscriber(client *redis.Client, stream string, group string, consumer string) EventSubscriber {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisEventSubscriber{
		redisClient:   client,
		streamName:    stream,
		consumerGroup: group,
		consumerName:  consumer,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// ensureConsumerGroup ensures the consumer group exists
func (s *RedisEventSubscriber) ensureConsumerGroup() error {
	// Try to create the consumer group, ignoring if it already exists
	err := s.redisClient.XGroupCreateMkStream(s.ctx, s.streamName, s.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

// Subscribe subscribes to events and processes them with the given handler
func (s *RedisEventSubscriber) Subscribe(handler EventHandler) {
	// Ensure consumer group exists
	if err := s.ensureConsumerGroup(); err != nil {
		log.Error().Err(err).Msg("Failed to create consumer group")
		return
	}

	// Start subscription in a goroutine
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				// Read from the stream
				streams, err := s.redisClient.XReadGroup(s.ctx, &redis.XReadGroupArgs{
					Group:    s.consumerGroup,
					Consumer: s.consumerName,
					Streams:  []string{s.streamName, ">"},
					Count:    10,
					Block:    time.Second,
				}).Result()

				if err != nil && err != redis.Nil {
					log.Error().Err(err).Msg("Failed to read from stream")
					time.Sleep(time.Second)
					continue
				}

				// Process messages
				for _, stream := range streams {
					for _, message := range stream.Messages {
						uuid, ok1 := message.Values["uuid"].(string)
						eventType, ok2 := message.Values["type"].(string)
						entityUUID, ok3 := message.Values["entity_uuid"].(string)
						if !ok1 || !ok2 || !ok3 {
							log.Warn().Msg("Invalid message format")
							continue
						}
						// Create event with just UUID and type since that's all we store in the stream
						event := BaseEvent{
							UUID:       uuid,
							EntityUUID: entityUUID,
							Type:       eventType,
						}
						// Handle the event
						if err := handler.HandleEvent(event); err != nil {
							log.Error().Err(err).Msg("Failed to handle event")
						} else {
							// Acknowledge the message
							s.redisClient.XAck(s.ctx, s.streamName, s.consumerGroup, message.ID)
							// Delete the message from the stream in production mode
							// log.Info().Str("app_env", os.Getenv("APP_ENV")).Msg("APP_ENV")
							if os.Getenv("APP_ENV") == "prod" {
								log.Info().Msg("Deleting message from stream")
								if err := s.redisClient.XDel(s.ctx, s.streamName, message.ID).Err(); err != nil {
									log.Error().Err(err).Msg("Failed to delete message from stream")
								}
							}
						}
					}
				}
			}
		}
	}()
}

// Close closes the subscription
func (s *RedisEventSubscriber) Close() {
	s.cancel()
}
