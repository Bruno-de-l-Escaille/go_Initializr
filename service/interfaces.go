package service

import (
	"context"
	"go_Initializr/models"
	"go_Initializr/repository"
)

// ExampleEntityServiceInterface defines the service interface for ExampleEntity
type ExampleEntityServiceInterface interface {
	CreateEntity(ctx context.Context, req *models.ExampleEntityCreateRequest) (*models.ExampleEntity, error)
	GetEntityByUUID(ctx context.Context, uuid string) (*models.ExampleEntity, error)
	GetEntityByEmail(ctx context.Context, email string) (*models.ExampleEntity, error)
	UpdateEntity(ctx context.Context, uuid string, req *models.ExampleEntityUpdateRequest) (*models.ExampleEntity, error)
	DeleteEntity(ctx context.Context, uuid string) error
	GetAllEntities(ctx context.Context, filter string, offset, limit int) ([]*models.ExampleEntity, int64, error)
}

// EventServiceInterface defines the service interface for events
type EventServiceInterface interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	GetEventsByEntityUUID(ctx context.Context, entityUUID string) ([]*models.Event, error)
}

// BaseServiceInterface defines common service operations
type BaseServiceInterface[T any] interface {
	CreateEntity(ctx context.Context, entity T) (*T, error)
	UpdateEntityByRef(ctx context.Context, ref string, entityData map[string]interface{}) (map[string]interface{}, error)
	DeleteEntityByRef(ctx context.Context, ref string) error
	GetAllEntities(ctx context.Context, filter string, offset, limit int) ([]T, int64, int64, error)
	GetEntityByRef(ctx context.Context, ref string) (*T, error)
}

// NewEventService creates a new event service
func NewEventService(repo repository.EventRepositoryInterface) EventServiceInterface {
	return &EventService{repo: repo}
}

// EventService implements EventServiceInterface
type EventService struct {
	repo repository.EventRepositoryInterface
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) error {
	return s.repo.CreateEvent(ctx, event)
}

// GetEventsByEntityUUID gets events by entity UUID
func (s *EventService) GetEventsByEntityUUID(ctx context.Context, entityUUID string) ([]*models.Event, error) {
	return s.repo.GetEventsByEntityUUID(ctx, entityUUID)
}
