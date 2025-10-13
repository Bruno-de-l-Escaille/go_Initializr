package base

import (
	"context"
	"go_Initializr/models"
	"go_Initializr/models/apperrors"
	"go_Initializr/repository"
	"go_Initializr/service"
	"go_Initializr/service/utils"
	"reflect"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/bytedance/sonic"
	"go.mongodb.org/mongo-driver/bson"
)

// BaseService provides a generic implementation of CRUD operations for any entity type
type BaseService[T any] struct {
	Repository     *repository.BaseRepository
	EventService   service.EventServiceInterface
	CollectionName string
}

// NewBaseService creates a new base service with the given repository
func NewBaseService[T any](repo *repository.BaseRepository, eventService service.EventServiceInterface, collectionName string) *BaseService[T] {
	if repo == nil {
		log.Fatal().Msg("FATAL: BaseRepository cannot be nil in NewBaseService")
	}
	if eventService == nil {
		log.Fatal().Msg("FATAL: EventServiceInterface cannot be nil in NewBaseService")
	}
	if collectionName == "" {
		log.Fatal().Msg("FATAL: CollectionName cannot be empty in NewBaseService")
	}

	return &BaseService[T]{
		Repository:     repo,
		EventService:   eventService,
		CollectionName: collectionName,
	}
}

// CreateEntity creates a new Entity in the system
func (s *BaseService[T]) CreateEntity(ctx context.Context, entity T) (*T, error) {
	payload, err := utils.OmitFields(entity, "uuid")
	if err != nil {
		return nil, apperrors.NewInternal()
	}
	// log.Info().Str("payload", string(payload)).Msg("Creating Entity")
	userID := ""
	userID, err = utils.GetAuthenticatedUser(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error getting authenticated user")
		return nil, apperrors.NewAuthorization("Missing user ID in context")
	}
	entityEvent, err := s.EventService.CreateAndProcessEvent(s.CollectionName, "CREATE", string(payload), "", userID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating event")
		return nil, apperrors.NewInternal()
	}
	// Use reflection to set fields on the entity
	entityValue := reflect.ValueOf(&entity).Elem()

	// Try to set UUID field
	uuidField := entityValue.FieldByName("UUID")
	if uuidField.IsValid() && uuidField.CanSet() {
		uuidField.SetString(entityEvent.EntityUUID)
	} else {
		log.Warn().Str("entity_type", reflect.TypeOf(entity).String()).Msg("UUID field not found or not settable in entity")
	}
	// Return the created entity
	return &entity, nil
}

// CreateEntitiesBulk creates multiple entities in a single operation
// It will return error if no entities are provided or if the context is invalid
func (s *BaseService[T]) CreateEntitiesBulk(ctx context.Context, entities []T) ([]T, error) {
	if ctx == nil {
		return nil, apperrors.NewBadRequest("Context is required")
	}

	userID, err := utils.GetAuthenticatedUser(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error getting authenticated user")
		return nil, apperrors.NewAuthorization("Missing user ID in context")
	}
	if len(entities) == 0 {
		return []T{}, nil
	}

	// Extract payloads for all entities
	payloads := make([]string, len(entities))
	refs := make([]string, len(entities)) // Empty refs for creation
	operations := make([]string, len(entities))

	for i := range entities {
		payload, err := utils.OmitFields(entities[i], "uuid")
		if err != nil {
			return nil, apperrors.NewInternal()
		}
		payloads[i] = string(payload)
		operations[i] = "CREATE"
	}

	// Create events in bulk
	events, err := s.EventService.CreateAndProcessEventsBulk(s.CollectionName, operations, payloads, refs, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating bulk events")
		return nil, apperrors.NewInternal()
	}

	// Update entity UUIDs from generated events
	for i, event := range events {
		// Use reflection to set fields on the entity
		entityValue := reflect.ValueOf(&entities[i]).Elem()

		// Try to set UUID field
		uuidField := entityValue.FieldByName("UUID")
		if uuidField.IsValid() && uuidField.CanSet() {
			uuidField.SetString(event.EntityUUID)
		} else {
			log.Warn().Str("entity_type", reflect.TypeOf(entities[i]).String()).Msg("UUID field not found or not settable in entity")
		}
	}

	return entities, nil
}

// UpdateUserByRef updates a user by reference (UUID)
func (s *BaseService[T]) UpdateEntityByRef(ctx context.Context, ref string, entityData map[string]interface{}) (map[string]interface{}, error) {
	userID, err := utils.GetAuthenticatedUser(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error getting authenticated user")
		return nil, apperrors.NewAuthorization("Missing user ID in context")
	}

	if ref == "" {
		return nil, apperrors.NewBadRequest("Reference cannot be empty")
	}
	if len(entityData) == 0 {
		return nil, apperrors.NewBadRequest("No update data provided")
	}
	payload, err := sonic.Marshal(entityData)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling update data")
		return nil, apperrors.NewInternal()
	}
	_, err = s.EventService.CreateAndProcessEvent(s.CollectionName, "UPDATE", string(payload), ref, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating update event")
		return nil, apperrors.NewInternal()
	}

	return entityData, nil
}

// DeleteEntityByRef deletes an entity by its reference (UUID)
func (s *BaseService[T]) DeleteEntityByRef(ctx context.Context, ref string) error {
	userID, err := utils.GetAuthenticatedUser(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error getting authenticated user")
		return apperrors.NewAuthorization("Missing user ID in context")
	}
	if ref == "" {
		return apperrors.NewBadRequest("Reference cannot be empty")
	}
	_, err = s.EventService.CreateAndProcessEvent(s.CollectionName, "DELETE", "", ref, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating delete event")
		return apperrors.NewInternal()
	}
	return nil
}

// GetAllEntities retrieves all entities of type T with filtering and optional pagination.
// Parameters:
//   - ctx: context.Context for the request
//   - filterStr: JSON string containing MongoDB filter criteria
//   - offset: Number of documents to skip (0 or negative values disable pagination)
//   - limit: Maximum number of documents to return (0 or negative values disable pagination)
//
// Returns:
//   - []T: Slice of entity objects
//   - int64: Total count of documents matching the filter (regardless of pagination)
//   - error: Any error that occurred during the operation
func (s *BaseService[T]) GetAllEntities(ctx context.Context, filterStr string, offset, limit int) ([]T, int64, int64, error) {
	startTime := time.Now()
	// Check if the snapshot is up to date
	s.CheckSnapshotAndEventStoreLastUUIDs(ctx)

	// Parse filter string to bson.M
	var filter bson.M
	if filterStr != "" {
		if err := sonic.Unmarshal([]byte(filterStr), &filter); err != nil {
			log.Error().Err(err).Msg("Error parsing filter JSON")
			return nil, 0, 0, apperrors.NewBadRequest("Invalid filter format: must be valid JSON")
		}
	} else {
		filter = bson.M{}
	}
	// Validate pagination parameters
	if offset != -1 && offset < 0 {
		return nil, 0, 0, apperrors.NewBadRequest("Offset cannot be negative")
	}
	if limit != -1 && limit < 0 {
		return nil, 0, 0, apperrors.NewBadRequest("Limit cannot be negative")
	}
	// Call the repository's GetAllEntities method
	entities, resultCount, totalCount, err := repository.GetAllEntities[T](ctx, s.Repository, s.CollectionName, filter, &offset, &limit)
	if err != nil {
		if appErr, ok := err.(*apperrors.Error); ok {
			return nil, 0, 0, appErr
		}
		log.Error().Err(err).Msg("Error getting all entities from repository")
		return nil, 0, 0, apperrors.NewInternal()
	}

	elapsed := time.Since(startTime)
	log.Info().Int("count", len(entities)).Int64("total", totalCount).Dur("duration", elapsed).Msg("Retrieved entities")

	return entities, resultCount, totalCount, nil
}

// GetEntityByRef retrieves an entity by its reference (UUID)
func (s *BaseService[T]) GetEntityByRef(ctx context.Context, ref string) (*T, error) {
	if ref == "" {
		return nil, apperrors.NewBadRequest("Reference cannot be empty")
	}
	// Check if the snapshot is up to date
	go s.CheckSnapshotAndEventStoreLastUUIDs(ctx)
	entity, err := repository.GetEntityByRef[T](ctx, s.Repository, ref, s.CollectionName)
	if err != nil {
		log.Warn().Err(err).Str("ref", ref).Msg("Entity not found for reference")
		if err.Error() == "mongo: no documents in result" {
			return nil, apperrors.NewNotFound("entity", ref)
		}
		return nil, apperrors.NewInternal()
	}
	return entity, nil
}

// GetEntityHistoryByRef retrieves the history of an entity by its reference
func (s *BaseService[T]) GetEntityHistoryByRef(ctx context.Context, ref string) ([]models.Event, error) {
	if ref == "" {
		return nil, apperrors.NewBadRequest("Reference cannot be empty")
	}
	events, err := s.Repository.GetEntityHistoryByRef(ctx, ref, s.CollectionName)
	if err != nil {
		log.Error().Err(err).Str("ref", ref).Msg("Error getting entity history by ref")
		return nil, apperrors.NewInternal()
	}
	return events, nil
}

// CheckSnapshotAndEventStoreLastUUIDs checks if the snapshot is up to date and updates it if necessary
func (s *BaseService[T]) CheckSnapshotAndEventStoreLastUUIDs(ctx context.Context) (map[string]int, error) {
	snapshotIsUpToDate, lastSnapshotUuid, err := s.Repository.CheckSnapshotAndEventStoreLastUUIDs(ctx, s.CollectionName)
	if err != nil {
		log.Error().Err(err).Msg("Error checking snapshot and event store last UUIDs")
		return nil, err
	}
	if lastSnapshotUuid == "" {
		log.Info().Str("collection_name", s.CollectionName).Msg("No snapshot found, creating snapshot")
		_, err = s.Repository.InitializeAllSnapshots(ctx, s.CollectionName)
		if err != nil {
			log.Error().Err(err).Msg("Error creating initial snapshot")
			return nil, err
		}
		log.Info().Msg("Initial snapshot created successfully")
		return nil, nil
	}
	if !snapshotIsUpToDate && lastSnapshotUuid != "" {
		log.Info().Msg("Snapshot is not up to date, creating snapshot")
		err = s.Repository.UpdateSnapshotSinceEventCuttOff(ctx, s.CollectionName, lastSnapshotUuid)
		if err != nil {
			log.Error().Err(err).Msg("Error creating snapshot")
			return nil, err
		}
		log.Info().Msg("Snapshot updated successfully")
	}
	log.Info().Msg("Snapshot is up to date")
	return nil, nil
}

// CheckAllSnapshotAndEventStoreLastUUIDs checks if snapshots are up to date for multiple entities and updates them if necessary
func (s *BaseService[T]) CheckAllSnapshotAndEventStoreLastUUIDs(ctx context.Context, entities []string) error {
	if len(entities) == 0 {
		return nil // No entities to process
	}

	log.Info().Int("entity_count", len(entities)).Msg("Starting CheckAllSnapshotAndEventStoreLastUUIDs")

	for _, entityType := range entities {
		log.Info().Str("entity_type", entityType).Msg("Checking snapshots for entity")

		// Call repository method for this specific entity
		snapshotIsUpToDate, lastSnapshotUuid, err := s.Repository.CheckSnapshotAndEventStoreLastUUIDs(ctx, entityType)
		if err != nil {
			log.Error().Err(err).Str("entity_type", entityType).Msg("Error checking snapshot")
			return err // Stop on first error
		}

		// Handle missing snapshot
		if lastSnapshotUuid == "" {
			log.Info().Str("entity_type", entityType).Msg("No snapshot found, creating snapshot")
			_, err = s.Repository.InitializeAllSnapshots(ctx, entityType)
			if err != nil {
				log.Error().Err(err).Str("entity_type", entityType).Msg("Error creating initial snapshot")
				return err
			}
			log.Info().Str("entity_type", entityType).Msg("Initial snapshot created successfully")
			continue
		}

		// Handle out-of-date snapshot
		if !snapshotIsUpToDate {
			log.Info().Str("entity_type", entityType).Msg("Snapshot is not up to date, updating snapshot")
			err = s.Repository.UpdateSnapshotSinceEventCuttOff(ctx, entityType, lastSnapshotUuid)
			if err != nil {
				log.Error().Err(err).Str("entity_type", entityType).Msg("Error updating snapshot")
				return err
			}
			log.Info().Str("entity_type", entityType).Msg("Snapshot updated successfully")
		} else {
			log.Info().Str("entity_type", entityType).Msg("Snapshot is up to date")
		}
	}

	log.Info().Int("entity_count", len(entities)).Msg("Successfully completed CheckAllSnapshotAndEventStoreLastUUIDs")
	return nil
}

// InitializeSnapshots initializes all snapshots for the entity type
func (s *BaseService[T]) InitializeSnapshots(ctx context.Context) (map[string]int, error) {
	log.Info().Str("collection_name", s.CollectionName).Msg("Initializing snapshots for entity type")
	entityMap, err := s.Repository.InitializeAllSnapshots(ctx, s.CollectionName)
	if err != nil {
		log.Error().Err(err).Str("collection_name", s.CollectionName).Msg("Error initializing snapshots")
		return nil, err
	}
	count := len(entityMap)
	log.Info().Int("count", count).Str("collection_name", s.CollectionName).Msg("Successfully initialized snapshots")
	result := map[string]int{s.CollectionName: count}
	return result, nil
}
