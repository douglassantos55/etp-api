package resource

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type Repository interface {
	// Returns the list of registered resources
	FetchResources() ([]*Resource, error)

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

func (r *goquRepository) FetchResources() ([]*Resource, error) {
	resources := make([]*Resource, 0)
	err := r.builder.From(goqu.T("resources")).ScanStructs(&resources)
	if err != nil {
		return nil, err
	}
	return resources, nil
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
