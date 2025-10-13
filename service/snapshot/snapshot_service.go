package snapshot

import (
	"context"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"

	"go_Initializr/models"
	"go_Initializr/repository"
	events "go_Initializr/repository/queue"
	"go_Initializr/service"
)

// SnapshotService implements SnapshotServiceInterface
type SnapshotService struct {
	eventRepo       repository.EventRepositoryInterface
	entityRepo      repository.BaseRepository
	eventSubscriber events.EventSubscriber
}

func NewSnapshotService(
	eventRepo repository.EventRepositoryInterface,
	subscriber events.EventSubscriber,
	entityRepository repository.BaseRepository) service.SnapshotServiceInterface {
	service := &SnapshotService{
		eventRepo:       eventRepo,
		eventSubscriber: subscriber,
		entityRepo:      entityRepository,
	}
	return service
}

// StartEventListener starts listening for events from Redis Streams
func (s *SnapshotService) StartEventListener() {
	// Create a handler for events
	handler := &EventHandler{snapshotService: s}
	// Start subscription using the event subscriber
	s.eventSubscriber.Subscribe(handler)
	log.Info().Msg("Event queue listener started")
}

// EventHandler implements events.EventHandler interface
type EventHandler struct {
	snapshotService *SnapshotService
}

// HandleEvent processes received events from Redis stream
func (h *EventHandler) HandleEvent(event events.Event) error {
	return h.snapshotService.ProcessEvent(event)
}

// ProcessEvent processes an incoming event
func (s *SnapshotService) ProcessEvent(event events.Event) error {
	// Process different event types
	eventType := event.GetType()
	// log.Info().Str("event_type", eventType).Str("uuid", event.GetUUID()).Msg("Processing event")

	// Extract entity type from event type
	// Support entity types with underscores (like community_role_CREATE)
	lastUnderscoreIndex := strings.LastIndex(eventType, "_")
	if lastUnderscoreIndex == -1 || lastUnderscoreIndex == 0 || lastUnderscoreIndex == len(eventType)-1 {
		log.Warn().Str("event_type", eventType).Msg("Invalid event type format")
		return nil
	}

	entityType := eventType[:lastUnderscoreIndex]
	operation := eventType[lastUnderscoreIndex+1:]

	// For our stream-based approach, we need to fetch the full event data using the UUID
	// The stream only contains UUID and type
	dbEvent, err := s.eventRepo.GetEventByUUIDAndType(event.GetUUID(), entityType)
	// log.Info().Str("uuid", event.GetUUID()).Msg("Event uuid")
	if err != nil {
		log.Error().Err(err).Str("uuid", event.GetUUID()).Msg("Failed to get event data for UUID")
		return err
	}

	ctx := context.Background()

	switch operation {
	case "CREATE", "UPDATE":
		// Unmarshal the payload into a generic map
		var entityData map[string]interface{}
		if err := sonic.Unmarshal([]byte(dbEvent.Payload), &entityData); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal event payload")
			return err
		}

		// Add entity UUID from the event
		entityData["uuid"] = dbEvent.EntityUUID
		err = s.UpdateSnapshot(ctx, entityData, entityType)
		if err == nil { // if no error, update the last snapshot event UUID
			s.entityRepo.UpdateLastSnapshotUUID(ctx, entityType, dbEvent.UUID)
		}
		return err
	case "DELETE":
		err = s.DeleteSnapshot(ctx, event.GetEntityUUID(), entityType)
		if err == nil {
			s.entityRepo.UpdateLastSnapshotUUID(ctx, entityType, dbEvent.UUID)
		}
		return err
	default:
		log.Warn().Str("operation", operation).Msg("Unknown operation")
		return nil
	}
}

// UpdateSnapshot updates a user snapshot
func (s *SnapshotService) UpdateSnapshot(ctx context.Context, entityData interface{}, entityType string) error {

	// log.Info().Interface("entity_data", entityData).Msg("Updating snapshot in MongoDB")
	return s.entityRepo.UpdateEntitySnapshot(ctx, entityData, entityType)
}

// DeleteSnapshot deletes a user snapshot
func (s *SnapshotService) DeleteSnapshot(ctx context.Context, entityUUID string, entityType string) error {
	log.Info().Str("entity_uuid", entityUUID).Msg("Deleting user snapshot in MongoDB")
	return s.entityRepo.DeleteEntitySnapshot(ctx, entityUUID, entityType)
}

// InitializeSnapshots initializes all entity snapshots
func (s *SnapshotService) InitializeAllSnapshots(ctx context.Context) (map[string]int, error) {
	entityMap, err := s.entityRepo.GetAllEntityEvents(ctx, "example_entity")

	if err != nil {
		return nil, err
	}

	// Use a copy of entityMap to avoid any potential race conditions
	entityMapCopy := make(map[string]*models.ExampleEntity, len(entityMap))
	for k, v := range entityMap {
		entityMapCopy[k] = v.(*models.ExampleEntity)
	}

	// Count of successfully updated snapshots
	updateCount := 0

	// Update snapshots for all entities
	for _, entity := range entityMapCopy {
		if err := s.entityRepo.UpdateEntitySnapshot(ctx, entity, "example_entity"); err != nil {
			log.Error().Err(err).Msg("Failed to update entity snapshot during initialization")
		} else {
			updateCount++
		}
	}

	// Update the last snapshotted UUID with the most recent event
	if len(entityMapCopy) > 0 {
		lastUUID, err := s.entityRepo.GetLastEventUUID(ctx, "example_entity")
		if err == nil && lastUUID != "" {
			if err := s.entityRepo.UpdateLastSnapshotUUID(ctx, "example_entity", lastUUID); err != nil {
				log.Error().Err(err).Msg("Failed to update last snapshot UUID during initialization")
			}
		}
	}

	result := map[string]int{"example_entity": updateCount}
	return result, nil
}
