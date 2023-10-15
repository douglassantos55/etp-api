package resource

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type Repository interface {
	// Creates a resource
	SaveResource(resource *Resource) (*Resource, error)
}

type goquRepository struct {
	builder *goqu.Database
}

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) SaveResource(resource *Resource) (*Resource, error) {
	result, err := r.builder.Insert("resources").Rows(resource).Executor().Exec()
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	resource.Id = uint64(id)

	return resource, nil
}
