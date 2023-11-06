package resource

import (
	"api/database"
	"context"

	"github.com/doug-martin/goqu/v9"
)

type Repository interface {
	// Returns the list of registered resources
	FetchResources(ctx context.Context) ([]*Resource, error)

	// Get a resource by id, returns nil if it can't be found
	GetById(ctx context.Context, id uint64) (*Resource, error)

	GetRequirements(ctx context.Context, resourceId uint64) ([]*Item, error)

	// Creates a resource
	SaveResource(ctx context.Context, resource *Resource) (*Resource, error)

	// Updates a resource
	UpdateResource(ctx context.Context, resource *Resource) (*Resource, error)
}

type goquRepository struct {
	builder *goqu.Database
}

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) FetchResources(ctx context.Context) ([]*Resource, error) {
	resources := make([]*Resource, 0)

	err := r.builder.
		Select(
			goqu.I("r.*"),
			goqu.I("c.id").As(goqu.C("category.id")),
			goqu.I("c.name").As(goqu.C("category.name")),
		).
		From(goqu.T("resources").As("r")).
		InnerJoin(goqu.T("categories").As("c"), goqu.On(goqu.I("r.category_id").Eq(goqu.I("c.id")))).
		ScanStructsContext(ctx, &resources)

	if err != nil {
		return nil, err
	}

	for _, resource := range resources {
		requirements, err := r.GetRequirements(ctx, resource.Id)
		if err != nil {
			return nil, err
		}
		resource.Requirements = requirements
	}

	return resources, nil
}

func (r *goquRepository) GetById(ctx context.Context, id uint64) (*Resource, error) {
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
		ScanStructContext(ctx, resource)

	if err != nil || !found {
		return nil, err
	}

	requirements, err := r.GetRequirements(ctx, resource.Id)
	if err != nil {
		return nil, err
	}

	resource.Requirements = requirements
	return resource, err
}

func (r *goquRepository) GetRequirements(ctx context.Context, resourceId uint64) ([]*Item, error) {
	requirements := make([]*Item, 0)

	err := r.builder.
		Select(
			goqu.I("req.quality"),
			goqu.I("req.qty").As("quantity"),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
			goqu.I("c.id").As(goqu.C("resource.category.id")),
			goqu.I("c.name").As(goqu.C("resource.category.name")),
		).
		From(goqu.T("resources_requirements").As("req")).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("req.requirement_id").Eq(goqu.I("r.id"))),
		).
		InnerJoin(
			goqu.T("categories").As("c"),
			goqu.On(
				goqu.And(
					goqu.I("r.category_id").Eq(goqu.I("c.id")),
					goqu.I("c.deleted_at").IsNull(),
				),
			),
		).
		Where(goqu.I("req.resource_id").Eq(resourceId)).
		ScanStructsContext(ctx, &requirements)

	return requirements, err
}

func (r *goquRepository) SaveResource(ctx context.Context, resource *Resource) (*Resource, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	result, err := tx.
		Insert("resources").
		Rows(goqu.Record{
			"name":        resource.Name,
			"image":       resource.Image,
			"category_id": resource.CategoryId,
		}).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err := r.saveRequirements(tx, id, resource.Requirements); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetById(ctx, uint64(id))
}

func (r *goquRepository) UpdateResource(ctx context.Context, resource *Resource) (*Resource, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	record := goqu.Record{
		"name":        resource.Name,
		"image":       resource.Image,
		"category_id": resource.CategoryId,
	}

	_, err = tx.
		Update(goqu.T("resources")).
		Set(record).
		Where(goqu.I("id").Eq(resource.Id)).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	if err := r.saveRequirements(tx, int64(resource.Id), resource.Requirements); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return resource, nil
}

func (r *goquRepository) saveRequirements(tx *goqu.TxDatabase, id int64, requirements []*Item) error {
	_, err := tx.Delete(goqu.T("resources_requirements")).
		Where(goqu.I("resource_id").Eq(id)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	if len(requirements) > 0 {
		reqs := []goqu.Record{}

		for _, requirement := range requirements {
			reqs = append(reqs, goqu.Record{
				"qty":            requirement.Qty,
				"quality":        requirement.Quality,
				"requirement_id": requirement.ResourceId,
				"resource_id":    id,
			})
		}

		_, err := tx.
			Insert(goqu.T("resources_requirements")).
			Rows(reqs).
			Executor().
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
}
