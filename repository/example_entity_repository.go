package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go_Initializr/models"
	"go_Initializr/pkg/uuidv7"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ExampleEntityRepository implements ExampleEntityRepositoryInterface
type ExampleEntityRepository struct {
	*BaseRepository
}

// NewExampleEntityRepository creates a new ExampleEntity repository
func NewExampleEntityRepository(db *sql.DB) *ExampleEntityRepository {
	return &ExampleEntityRepository{
		BaseRepository: &BaseRepository{DB: db},
	}
}

// Create creates a new ExampleEntity in the database
func (r *ExampleEntityRepository) Create(ctx context.Context, entity *models.ExampleEntity) error {
	if entity.UUID == "" {
		// Generate base62 encoded UUID v7 (like gopeople)
		base62UUID, err := uuidv7.GenerateB62()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate base62 UUID")
			return fmt.Errorf("failed to generate UUID: %w", err)
		}
		entity.UUID = base62UUID
	}

	entity.SetDefaults()

	query := `
		INSERT INTO example_entities (uuid, name, description, email, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.ExecContext(ctx, query,
		entity.UUID,
		entity.Name,
		entity.Description,
		entity.Email,
		entity.Status,
		entity.CreatedAt,
		entity.UpdatedAt,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create ExampleEntity")
		return err
	}

	return nil
}

// GetByUUID retrieves an ExampleEntity by UUID
func (r *ExampleEntityRepository) GetByUUID(ctx context.Context, uuid string) (*models.ExampleEntity, error) {
	query := `
		SELECT uuid, name, description, email, status, created_at, updated_at 
		FROM example_entities 
		WHERE uuid = $1`

	var entity models.ExampleEntity
	err := r.DB.QueryRowContext(ctx, query, uuid).Scan(
		&entity.UUID,
		&entity.Name,
		&entity.Description,
		&entity.Email,
		&entity.Status,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ExampleEntity with UUID %s not found", uuid)
		}
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to get ExampleEntity by UUID")
		return nil, err
	}

	return &entity, nil
}

// GetByEmail retrieves an ExampleEntity by email
func (r *ExampleEntityRepository) GetByEmail(ctx context.Context, email string) (*models.ExampleEntity, error) {
	query := `
		SELECT uuid, name, description, email, status, created_at, updated_at 
		FROM example_entities 
		WHERE email = $1`

	var entity models.ExampleEntity
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&entity.UUID,
		&entity.Name,
		&entity.Description,
		&entity.Email,
		&entity.Status,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ExampleEntity with email %s not found", email)
		}
		log.Error().Err(err).Str("email", email).Msg("Failed to get ExampleEntity by email")
		return nil, err
	}

	return &entity, nil
}

// Update updates an ExampleEntity in the database
func (r *ExampleEntityRepository) Update(ctx context.Context, uuid string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	// Add updated_at to updates
	updates["updated_at"] = time.Now()

	// Build dynamic query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	args = append(args, uuid) // Add UUID as the last argument

	query := fmt.Sprintf(`
		UPDATE example_entities 
		SET %s 
		WHERE uuid = $%d`,
		strings.Join(setParts, ", "),
		argIndex,
	)

	result, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to update ExampleEntity")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ExampleEntity with UUID %s not found", uuid)
	}

	return nil
}

// Delete deletes an ExampleEntity from the database
func (r *ExampleEntityRepository) Delete(ctx context.Context, uuid string) error {
	query := `DELETE FROM example_entities WHERE uuid = $1`

	result, err := r.DB.ExecContext(ctx, query, uuid)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to delete ExampleEntity")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ExampleEntity with UUID %s not found", uuid)
	}

	return nil
}

// GetAll retrieves all ExampleEntities with optional filtering and pagination
func (r *ExampleEntityRepository) GetAll(ctx context.Context, filter string, offset, limit int) ([]*models.ExampleEntity, int64, error) {
	var whereClause string
	var args []interface{}
	argIndex := 1

	// Build WHERE clause if filter is provided
	if filter != "" {
		whereClause = "WHERE (name ILIKE $1 OR email ILIKE $1 OR description ILIKE $1)"
		args = append(args, "%"+filter+"%")
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM example_entities %s", whereClause)
	var totalCount int64
	err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count ExampleEntities")
		return nil, 0, err
	}

	// Main query with pagination
	query := fmt.Sprintf(`
		SELECT uuid, name, description, email, status, created_at, updated_at 
		FROM example_entities %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get ExampleEntities")
		return nil, 0, err
	}
	defer rows.Close()

	var entities []*models.ExampleEntity
	for rows.Next() {
		var entity models.ExampleEntity
		err := rows.Scan(
			&entity.UUID,
			&entity.Name,
			&entity.Description,
			&entity.Email,
			&entity.Status,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan ExampleEntity")
			return nil, 0, err
		}
		entities = append(entities, &entity)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("Error iterating over ExampleEntity rows")
		return nil, 0, err
	}

	return entities, totalCount, nil
}
