package building

import (
	"api/database"
	"api/resource"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		GetAll() ([]*Building, error)
		GetById(uint64) (*Building, error)
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) GetAll() ([]*Building, error) {
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
		ScanStructs(&buildings)

	if err != nil {
		return nil, err
	}

	for _, building := range buildings {
		requirements, err := r.GetRequirements(building.Id)
		if err != nil {
			return nil, err
		}
		building.Requirements = requirements
	}

	return buildings, nil
}

func (r *goquRepository) GetById(id uint64) (*Building, error) {
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
		ScanStruct(building)

	if err != nil || !found {
		return nil, err
	}

	requirements, err := r.GetRequirements(id)
	if err != nil {
		return nil, err
	}

	building.Requirements = requirements

	return building, err
}

func (r *goquRepository) GetRequirements(buildingId uint64) ([]*resource.Item, error) {
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
		ScanStructs(&requirements)

	return requirements, err
}
