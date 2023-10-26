package resource

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type Repository interface {
	// Returns the list of registered resources
	FetchResources() ([]*Resource, error)

	// Get a resource by id, returns nil if it can't be found
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

	err := r.builder.
		Select(
			goqu.I("r.*"),
			goqu.I("c.id").As(goqu.C("category.id")),
			goqu.I("c.name").As(goqu.C("category.name")),
		).
		From(goqu.T("resources").As("r")).
		InnerJoin(goqu.T("categories").As("c"), goqu.On(goqu.I("r.category_id").Eq(goqu.I("c.id")))).
		ScanStructs(&resources)

	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *goquRepository) GetById(id uint64) (*Resource, error) {
	resource := new(Resource)

	found, err := r.builder.
		Select(
			goqu.I("r.*"),
			goqu.I("c.id").As(goqu.C("category.id")),
			goqu.I("c.name").As(goqu.C("category.name")),
		).
		From(goqu.T("resources").As("r")).
		InnerJoin(
			goqu.T("categories").As("c"),
			goqu.On(
				goqu.And(
					goqu.I("r.category_id").Eq(goqu.I("c.id")),
					goqu.I("c.deleted_at").IsNull(),
				),
			),
		).
		Where(goqu.I("r.id").Eq(id)).
		ScanStruct(resource)

	if err != nil || !found {
		return nil, err
	}

	return resource, err
}

func (r *goquRepository) SaveResource(resource *Resource) (*Resource, error) {
	record := goqu.Record{
		"name":        resource.Name,
		"image":       resource.Image,
		"category_id": resource.CategoryId,
	}

	result, err := r.builder.Insert("resources").Rows(record).Executor().Exec()
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	resource.Id = uint64(id)

	return resource, nil
}

func (r *goquRepository) UpdateResource(resource *Resource) (*Resource, error) {
	record := goqu.Record{
		"name":        resource.Name,
		"image":       resource.Image,
		"category_id": resource.CategoryId,
	}

	_, err := r.builder.Update(goqu.T("resources")).
		Set(record).
		Where(goqu.I("id").Eq(resource.Id)).
		Executor().Exec()

	if err != nil {
		return nil, err
	}

	return resource, nil
}
