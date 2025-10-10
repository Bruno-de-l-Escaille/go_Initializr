package injection

import (
	"go_Initializr/handler"
	"go_Initializr/repository"
	"go_Initializr/service"
)

// exampleEntityComponents holds ExampleEntity-related components
type exampleEntityComponents struct {
	repository repository.ExampleEntityRepositoryInterface
	service    service.ExampleEntityServiceInterface
	handler    *handler.ExampleEntityHandler
}

// initializeExampleEntityComponents initializes ExampleEntity components
func initializeExampleEntityComponents(core *coreComponents) *exampleEntityComponents {
	// Initialize repository
	repo := repository.NewExampleEntityRepository(core.db)

	// Initialize service
	svc := service.NewExampleEntityService(repo)

	// Initialize handler
	hdlr := handler.NewExampleEntityHandler(svc)

	return &exampleEntityComponents{
		repository: repo,
		service:    svc,
		handler:    hdlr,
	}
}

// GetRepository returns the ExampleEntity repository
func (c *exampleEntityComponents) GetRepository() repository.ExampleEntityRepositoryInterface {
	return c.repository
}

// GetService returns the ExampleEntity service
func (c *exampleEntityComponents) GetService() service.ExampleEntityServiceInterface {
	return c.service
}

// GetHandler returns the ExampleEntity handler
func (c *exampleEntityComponents) GetHandler() *handler.ExampleEntityHandler {
	return c.handler
}
