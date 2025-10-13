package injection

import (
	"crypto/ecdsa"
	"database/sql"
	"go_Initializr/handler"
	"go_Initializr/service"
	"go_Initializr/pkg/nosql"
)

// Container holds all application dependencies.
type Container struct {
	// Core components
	core *coreComponents

	// Business components
	exampleEntityComponents *exampleEntityComponents
	snapshotComponents      *snapshotComponents
}

// NewContainer creates and initializes the dependency injection container.
func NewContainer(db *sql.DB, mongoClient *nosql.MongoDB) (*Container, error) {
	// Initialize core components first
		core, err := initializeCoreComponents(db, mongoClient)

	if err != nil {
		return nil, err
	}

	// Initialize business components
	exampleEntityComps := initializeExampleEntityComponents(core)
	snapshotComps := initializeSnapshotComponents(core)

	return &Container{
		core:                    core,
		exampleEntityComponents: exampleEntityComps,
		snapshotComponents:      snapshotComps,
	}, nil
}

// GetExampleEntityHandler returns the ExampleEntity handler
func (c *Container) GetExampleEntityHandler() *handler.ExampleEntityHandler {
	return c.exampleEntityComponents.handler
}

// GetExampleEntityComponents returns ExampleEntity components
func (c *Container) GetExampleEntityComponents() *exampleEntityComponents {
	return c.exampleEntityComponents
}

// GetPublicKey returns the ECDSA public key for JWT validation
func (c *Container) GetPublicKey() *ecdsa.PublicKey {
	return c.core.GetPublicKey()
}

// GetSnapshotService returns the snapshot service
func (c *Container) GetSnapshotService() service.SnapshotServiceInterface {
	return c.snapshotComponents.snapshotService
}
