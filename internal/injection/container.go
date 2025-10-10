package injection

import (
	"database/sql"
	"go_Initializr/handler"
)

// Container holds all application dependencies.
type Container struct {
	// Core components
	core *coreComponents

	// Business components
	exampleEntityComponents *exampleEntityComponents
}

// NewContainer creates and initializes the dependency injection container.
func NewContainer(db *sql.DB) (*Container, error) {
	// Initialize core components first
	core, err := initializeCoreComponents(db)
	if err != nil {
		return nil, err
	}

	// Initialize business components
	exampleEntityComps := initializeExampleEntityComponents(core)

	return &Container{
		core:                    core,
		exampleEntityComponents: exampleEntityComps,
	}, nil
}

// GetExampleEntityHandler returns the ExampleEntity handler
func (c *Container) GetExampleEntityHandler() *handler.ExampleEntityHandler {
	return c.exampleEntityComponents.handler
}

// GetConfig returns the application configuration
func (c *Container) GetConfig() *AppConfig {
	return c.core.GetConfig()
}

// GetDB returns the database connection
func (c *Container) GetDB() *sql.DB {
	return c.core.GetDB()
}

// GetExampleEntityComponents returns ExampleEntity components
func (c *Container) GetExampleEntityComponents() *exampleEntityComponents {
	return c.exampleEntityComponents
}
