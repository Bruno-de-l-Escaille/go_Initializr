package example_Entity

import (
	"database/sql"
	"go_Initializr/pkg/nosql"
	"go_Initializr/repository"
)

type ExampleEntityRepository struct {
	repository.BaseRepository
}

func NewExampleEntityRepository(db *sql.DB, mongoClient *nosql.MongoDB) *ExampleEntityRepository {
	return &ExampleEntityRepository{
		BaseRepository: repository.BaseRepository{
			DB:          db,
			MongoClient: mongoClient,
		},
	}
}
func (r *ExampleEntityRepository) GetBaseRepository() *repository.BaseRepository {
	return &r.BaseRepository
}
