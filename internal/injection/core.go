package injection

import (
	"database/sql"
	"go_Initializr/repository"
)

// coreComponents holds fundamental shared dependencies.
type coreComponents struct {
	config         *AppConfig
	db             *sql.DB
	baseRepository *repository.BaseRepository
}

// initializeCoreComponents creates the essential shared components.
func initializeCoreComponents(db *sql.DB) (*coreComponents, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	baseRepo := &repository.BaseRepository{
		DB: db,
	}

	return &coreComponents{
		config:         config,
		db:             db,
		baseRepository: baseRepo,
	}, nil
}

// GetConfig returns the application configuration
func (c *coreComponents) GetConfig() *AppConfig {
	return c.config
}

// GetDB returns the database connection
func (c *coreComponents) GetDB() *sql.DB {
	return c.db
}

// GetBaseRepository returns the base repository
func (c *coreComponents) GetBaseRepository() *repository.BaseRepository {
	return c.baseRepository
}
