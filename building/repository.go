package building

import (
	"api/database"
	"api/resource"
	"context"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		// Lists all registered buildings
		GetAll(ctx context.Context) ([]*Building, error)

		// Get building by ID
		GetById(ctx context.Context, id uint64) (*Building, error)
	}

	goquRepository struct {
		builder   *goqu.Database
		resources resource.Repository
	}
)

func NewRepository(conn *database.Connection, resources resource.Repository) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder, resources}
}

func (r *goquRepository) GetAll(ctx context.Context) ([]*Building, error) {
	buildings := make([]*Building, 0)

	err := r.builder.
		Select(
			goqu.I("id"),
			goqu.I("name"),
			goqu.I("wages_per_hour"),
			goqu.I("admin_per_hour"),
			goqu.I("maintenance_per_hour"),
		).
		From(goqu.T("buildings")).
		Where(goqu.I("deleted_at").IsNull()).
		ScanStructsContext(ctx, &buildings)

	if err != nil {
		return nil, err
	}

	for _, building := range buildings {
		requirements, err := r.GetRequirements(ctx, building.Id)
		if err != nil {
			return nil, err
		}

		resources, err := r.GetResources(ctx, building.Id)
		if err != nil {
			return nil, err
		}

		building.Resources = resources
		building.Requirements = requirements
	}

	return buildings, nil
}

func (r *goquRepository) GetById(ctx context.Context, id uint64) (*Building, error) {
	building := new(Building)

	found, err := r.builder.
		Select(
			goqu.I("id"),
			goqu.I("name"),
			goqu.I("wages_per_hour"),
			goqu.I("admin_per_hour"),
			goqu.I("maintenance_per_hour"),
		).
		From(goqu.T("buildings")).
		Where(goqu.And(
			goqu.I("id").Eq(id),
			goqu.I("deleted_at").IsNull(),
		)).
		ScanStructContext(ctx, building)

	if err != nil || !found {
		return nil, err
	}

	requirements, err := r.GetRequirements(ctx, id)
	if err != nil {
		return nil, err
	}

	resources, err := r.GetResources(ctx, id)
	if err != nil {
		return nil, err
	}

	building.Resources = resources
	building.Requirements = requirements

	return building, err
}

// Get the resources available for production in a given building
func (r *goquRepository) GetResources(ctx context.Context, buildingId uint64) ([]*BuildingResource, error) {
	resources := make([]*BuildingResource, 0)

	err := r.builder.
		Select(
			goqu.I("br.qty_per_hour"),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
			goqu.I("r.id").As(goqu.C("resource.id")),
		).
		From(goqu.T("buildings_resources").As("br")).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("br.resource_id").Eq(goqu.I("r.id"))),
		).
		Where(goqu.I("br.building_id").Eq(buildingId)).
		ScanStructsContext(ctx, &resources)

	for _, resource := range resources {
		requirements, err := r.resources.GetRequirements(ctx, resource.Resource.Id)
		if err != nil {
			return nil, err
		}
		resource.Resource.Requirements = requirements
	}

	return resources, err
}

// Get the resources required to construct a given building
func (r *goquRepository) GetRequirements(ctx context.Context, buildingId uint64) ([]*resource.Item, error) {
	requirements := make([]*resource.Item, 0)

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
		From(goqu.T("buildings_requirements").As("req")).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("req.resource_id").Eq(goqu.I("r.id"))),
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
		Where(goqu.I("req.building_id").Eq(buildingId)).
		ScanStructsContext(ctx, &requirements)

	return requirements, err
}
