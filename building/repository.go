package building

import (
	"api/database"

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

	return building, err
}
