package service

import (
	"context"
	"go_Initializr/models"
	events "go_Initializr/repository/queue"
)

// EventServiceInterface defines operations for event creation and processing
type EventServiceInterface interface {
	CreateAndProcessEvent(entity, operation string, payload string, ref string, actorUUID string) (*models.Event, error)
	CreateAndProcessEventsBulk(entity string, operations []string, payloads []string, refs []string, actorUUID string) ([]*models.Event, error)
}

// SnapshotServiceInterface defines operations for snapshot management
type SnapshotServiceInterface interface {
	StartEventListener()
	ProcessEvent(event events.Event) error
}

// BaseServiceInterface defines common event sourcing operations
type BaseServiceInterface[T any] interface {
	CreateEntity(ctx context.Context, entity T) (*T, error)
	UpdateEntityByRef(ctx context.Context, ref string, entityData map[string]interface{}) (map[string]interface{}, error)
	DeleteEntityByRef(ctx context.Context, ref string) error
	GetAllEntities(ctx context.Context, filter string, offset, limit int) ([]T, int64, int64, error)
	GetEntityByRef(ctx context.Context, ref string) (*T, error)
	GetEntityHistoryByRef(ctx context.Context, ref string) ([]models.Event, error)
	CreateEntitiesBulk(ctx context.Context, entities []T) ([]T, error)
	InitializeSnapshots(ctx context.Context) (map[string]int, error)
	CheckSnapshotAndEventStoreLastUUIDs(ctx context.Context) (map[string]int, error)
	CheckAllSnapshotAndEventStoreLastUUIDs(ctx context.Context, entities []string) error
}

// ExampleEntityServiceInterface defines operations for ExampleEntity with event sourcing
type ExampleEntityServiceInterface interface {
	BaseServiceInterface[models.ExampleEntity]

}
