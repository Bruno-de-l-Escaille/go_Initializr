package service

import (
	"context"
	"go_Initializr/models"
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

