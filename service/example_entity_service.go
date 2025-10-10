package service

import (
	"context"
	"fmt"
	"go_Initializr/models"
	"go_Initializr/pkg/uuidv7"
	"go_Initializr/repository"
	"time"

	"github.com/rs/zerolog/log"
)

// ExampleEntityService implements ExampleEntityServiceInterface
type ExampleEntityService struct {
	repo repository.ExampleEntityRepositoryInterface
}

// NewExampleEntityService creates a new ExampleEntity service
func NewExampleEntityService(repo repository.ExampleEntityRepositoryInterface) *ExampleEntityService {
	return &ExampleEntityService{
		repo: repo,
	}
}

// CreateEntity creates a new ExampleEntity
func (s *ExampleEntityService) CreateEntity(ctx context.Context, req *models.ExampleEntityCreateRequest) (*models.ExampleEntity, error) {
	// Validate request
	if validationErrors := req.ValidateCreate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", validationErrors)
	}

	// Check if entity with email already exists
	existingEntity, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existingEntity != nil {
		return nil, fmt.Errorf("entity with email %s already exists", req.Email)
	}

	// Create new entity with base62 UUID
	base62UUID, err := uuidv7.GenerateB62()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate base62 UUID")
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	entity := &models.ExampleEntity{
		UUID:        base62UUID,
		Name:        req.Name,
		Description: req.Description,
		Email:       req.Email,
		Status:      req.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	entity.SetDefaults()

	if err := s.repo.Create(ctx, entity); err != nil {
		log.Error().Err(err).Msg("Failed to create ExampleEntity")
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	log.Info().Str("uuid", entity.UUID).Str("email", entity.Email).Msg("ExampleEntity created successfully")
	return entity, nil
}

// GetEntityByUUID retrieves an ExampleEntity by UUID
func (s *ExampleEntityService) GetEntityByUUID(ctx context.Context, uuid string) (*models.ExampleEntity, error) {
	entity, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to get ExampleEntity by UUID")
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return entity, nil
}

// GetEntityByEmail retrieves an ExampleEntity by email
func (s *ExampleEntityService) GetEntityByEmail(ctx context.Context, email string) (*models.ExampleEntity, error) {
	entity, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("Failed to get ExampleEntity by email")
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return entity, nil
}

// UpdateEntity updates an ExampleEntity
func (s *ExampleEntityService) UpdateEntity(ctx context.Context, uuid string, req *models.ExampleEntityUpdateRequest) (*models.ExampleEntity, error) {
	// Check if entity exists
	existingEntity, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Email != nil {
		// Check if new email is already used by another entity
		if *req.Email != existingEntity.Email {
			emailExists, err := s.repo.GetByEmail(ctx, *req.Email)
			if err == nil && emailExists != nil {
				return nil, fmt.Errorf("email %s is already in use", *req.Email)
			}
		}
		updates["email"] = *req.Email
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return existingEntity, nil // No updates needed
	}

	// Perform update
	if err := s.repo.Update(ctx, uuid, updates); err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to update ExampleEntity")
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// Get updated entity
	updatedEntity, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to get updated ExampleEntity")
		return nil, fmt.Errorf("failed to get updated entity: %w", err)
	}

	log.Info().Str("uuid", uuid).Msg("ExampleEntity updated successfully")
	return updatedEntity, nil
}

// DeleteEntity deletes an ExampleEntity
func (s *ExampleEntityService) DeleteEntity(ctx context.Context, uuid string) error {
	// Check if entity exists
	_, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("entity not found: %w", err)
	}

	if err := s.repo.Delete(ctx, uuid); err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to delete ExampleEntity")
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	log.Info().Str("uuid", uuid).Msg("ExampleEntity deleted successfully")
	return nil
}

// GetAllEntities retrieves all ExampleEntities with filtering and pagination
func (s *ExampleEntityService) GetAllEntities(ctx context.Context, filter string, offset, limit int) ([]*models.ExampleEntity, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	entities, totalCount, err := s.repo.GetAll(ctx, filter, offset, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all ExampleEntities")
		return nil, 0, fmt.Errorf("failed to get entities: %w", err)
	}

	return entities, totalCount, nil
}
