package injection

import (
	"go_Initializr/handler"
	"go_Initializr/models"
	"go_Initializr/repository"
	"go_Initializr/repository/example_Entity"
	"go_Initializr/service/base"
	example_EntityService "go_Initializr/service/example_Entity"
	"go_Initializr/service"
)

// exampleEntityComponents holds ExampleEntity-related components
type exampleEntityComponents struct {
	repository               repository.ExampleEntityRepositoryInterface
	service                  service.ExampleEntityServiceInterface
	exampleEntityBaseService  *base.BaseService[models.ExampleEntity]
	handler                  *handler.ExampleEntityHandler
}

// initializeExampleEntityComponents initializes ExampleEntity components
func initializeExampleEntityComponents(core *coreComponents) *exampleEntityComponents {
	// Initialize repository
	exampleEntityRepository := &example_Entity.ExampleEntityRepository{
		BaseRepository: *core.baseRepository,
	}

	exampleEntityBaseService := base.NewBaseService[models.ExampleEntity](
		core.baseRepository,
		core.eventService,
		"example_entity",
	)
	

	// Initialize service
	exampleEntityService := example_EntityService.NewExampleEntityService(exampleEntityRepository, core.eventService)

	// Initialize handler
	example_EntityHandler := &handler.ExampleEntityHandler{
		ExampleEntityService: exampleEntityService,
	}

	return &exampleEntityComponents{
		repository: exampleEntityRepository,
		service:    exampleEntityService,
		exampleEntityBaseService: exampleEntityBaseService,
		handler:    example_EntityHandler,
	}
}
