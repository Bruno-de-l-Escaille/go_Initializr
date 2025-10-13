package repository

import (
	"context"
	"go_Initializr/models"
)

// EventRepositoryInterface defines operations for event data access
type EventRepositoryInterface interface {
	GetEventByUUIDAndType(uuid string, entityType string) (*models.Event, error)
}


type EntityRepositoryInterface interface {
	GetLastSnapshotUUID(ctx context.Context, entityType string) (string, error)
	UpdateLastSnapshotUUID(ctx context.Context, entityType string, uuid string) error
	UpdateEntitySnapshot(ctx context.Context, entityData interface{}, entityType string) error
	DeleteEntitySnapshot(ctx context.Context, uuid string, entityType string) error
	ApplyEventsToEntityMap(ctx context.Context, entityMap map[string]interface{}, lastSnapshottedUUID string, entityType string) (int, error)
	GetEntityHistoryByRef(ctx context.Context, ref string, entityType string) ([]models.Event, error)
	GetAllEntityEvents(ctx context.Context, entityType string) (map[string]interface{}, error)
	GetLastEventUUID(ctx context.Context, entityType string) (string, error)
	SaveEntityEvent(event *models.Event, entityType string) error
	GetEntitySnapshotByUUID(ctx context.Context, uuid string, entityType string, result interface{}) error
	InitializeAllSnapshots(ctx context.Context, entityType string) (map[string]interface{}, error)
	CheckSnapshotAndEventStoreLastUUIDs(ctx context.Context, entityType string) (bool, string, error)
	UpdateSnapshotSinceEventCuttOff(ctx context.Context, entityType string, eventCutOffUuid string) error
}
// ExampleEntityRepositoryInterface defines operations for ExampleEntity data access
type ExampleEntityRepositoryInterface interface {
	EntityRepositoryInterface
	GetBaseRepository() *BaseRepository
}