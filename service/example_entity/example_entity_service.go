package service

import (
	"fmt"
	"go_Initializr/models"
	"go_Initializr/repository"
	"go_Initializr/service/base"
	"go_Initializr/service"
)

// ExampleEntityService implements ExampleEntityServiceInterface (following gopeople pattern)
type ExampleEntityService struct {
	Repository repository.ExampleEntityRepositoryInterface
	*base.BaseService[models.ExampleEntity]
}

// Compile-time check to ensure ExampleEntityService implements ExampleEntityServiceInterface
var _ service.ExampleEntityServiceInterface = (*ExampleEntityService)(nil)

// NewExampleEntityService creates a new ExampleEntity service
func NewExampleEntityService(
	repo repository.ExampleEntityRepositoryInterface,
	eventService service.EventServiceInterface,
) service.ExampleEntityServiceInterface {
	// Cast to concrete type to access GetBaseRepository
	
	baseRepo := repo.GetBaseRepository()
	if baseRepo == nil {
		panic(fmt.Sprintf("GetBaseRepository returned nil for %T", repo))
	}
	return &ExampleEntityService{
		Repository: repo,
		BaseService:         base.NewBaseService[models.ExampleEntity](baseRepo, eventService, "example_entity"),
	}
}
