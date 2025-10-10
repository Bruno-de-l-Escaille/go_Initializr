package repository

import (
	"context"
	"database/sql"
	"go_Initializr/models"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	DB *sql.DB
}
// ExampleEntityRepositoryInterface defines operations for ExampleEntity data access
type ExampleEntityRepositoryInterface interface {
	Create(ctx context.Context, entity *models.ExampleEntity) error
	GetByUUID(ctx context.Context, uuid string) (*models.ExampleEntity, error)
	Update(ctx context.Context, uuid string, updates map[string]interface{}) error
	Delete(ctx context.Context, uuid string) error
	GetAll(ctx context.Context, filter string, offset, limit int) ([]*models.ExampleEntity, int64, error)
	GetByEmail(ctx context.Context, email string) (*models.ExampleEntity, error)
}
