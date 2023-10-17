package resource

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type Repository interface {
	// Returns the list of registered resources
	FetchResources() ([]*Resource, error)

	// Get a resource by id
	GetById(id uint64) (*Resource, error)

	// Creates a resource
	SaveResource(resource *Resource) (*Resource, error)

	// Updates a resource
	UpdateResource(resource *Resource) (*Resource, error)
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

func (r *goquRepository) GetById(id uint64) (*Resource, error) {
	resource := new(Resource)

	found, err := r.builder.From(goqu.T("resources")).
		Where(goqu.I("id").Eq(id)).
		ScanStruct(resource)

	if !found {
		return nil, err
	}

	return resource, err
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

func (r *goquRepository) UpdateResource(resource *Resource) (*Resource, error) {
	_, err := r.builder.Update(goqu.T("resources")).
		Set(resource).
		Where(goqu.C("id").Eq(resource.Id)).
		Executor().Exec()

	if err != nil {
		return nil, err
	}

	return resource, nil
}
