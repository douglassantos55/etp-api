package company

import (
	"api/building"
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		Register(registration *Registration) (*Company, error)

		GetById(id uint64) (*Company, error)

		GetByEmail(email string) (*Company, error)

		GetBuildings(companyId uint64) ([]*CompanyBuilding, error)

		AddBuilding(companyId uint64, building *building.Building, position uint8) (*CompanyBuilding, error)
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) GetById(id uint64) (*Company, error) {
	company := new(Company)

	found, err := r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
		).
		From(goqu.T("companies").As("c")).
		Where(
			goqu.And(
				goqu.I("c.id").Eq(id),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		ScanStruct(company)

	if err != nil || !found {
		return nil, err
	}

	return company, err
}

func (r *goquRepository) GetByEmail(email string) (*Company, error) {
	company := new(Company)

	found, err := r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
		).
		From(goqu.T("companies").As("c")).
		Where(
			goqu.And(
				goqu.I("email").Eq(email),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		ScanStruct(company)

	if err != nil || !found {
		return nil, err
	}

	return company, nil
}

func (r *goquRepository) Register(registration *Registration) (*Company, error) {
	record := goqu.Record{
		"name":     registration.Name,
		"email":    registration.Email,
		"password": registration.Password,
	}

	result, err := r.builder.
		Insert(goqu.T("companies")).
		Rows(record).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetById(uint64(id))
}

func (r *goquRepository) GetBuildings(companyId uint64) ([]*CompanyBuilding, error) {
	buildings := make([]*CompanyBuilding, 0)

	err := r.builder.
		Select(
			goqu.I("cb.id"),
			goqu.I("cb.name"),
			goqu.L("? * ?", goqu.I("b.wages_per_hour"), goqu.I("cb.level")).As("wages_per_hour"),
			goqu.L("? * ?", goqu.I("b.admin_per_hour"), goqu.I("cb.level")).As("admin_per_hour"),
			goqu.L("? * ?", goqu.I("b.maintenance_per_hour"), goqu.I("cb.level")).As("maintenance_per_hour"),
			goqu.I("cb.level"),
			goqu.I("cb.position"),
		).
		From(goqu.T("companies_buildings").As("cb")).
		InnerJoin(
			goqu.T("buildings").As("b"),
			goqu.On(
				goqu.And(
					goqu.I("b.id").Eq(goqu.I("cb.building_id")),
					goqu.I("b.deleted_at").IsNull(),
				),
			),
		).
		Where(goqu.And(
			goqu.I("cb.company_id").Eq(companyId),
			goqu.I("cb.demolished_at").IsNull(),
		)).
		ScanStructs(&buildings)

	if err != nil {
		return nil, err
	}

	return buildings, nil
}

func (r *goquRepository) GetBuildingById(id uint64) (*CompanyBuilding, error) {
	building := new(CompanyBuilding)

	found, err := r.builder.
		Select(
			goqu.I("cb.id"),
			goqu.I("cb.name"),
			goqu.L("? * ?", goqu.I("b.wages_per_hour"), goqu.I("cb.level")).As("wages_per_hour"),
			goqu.L("? * ?", goqu.I("b.admin_per_hour"), goqu.I("cb.level")).As("admin_per_hour"),
			goqu.L("? * ?", goqu.I("b.maintenance_per_hour"), goqu.I("cb.level")).As("maintenance_per_hour"),
			goqu.I("cb.level"),
			goqu.I("cb.position"),
		).
		From(goqu.T("companies_buildings").As("cb")).
		InnerJoin(
			goqu.T("buildings").As("b"),
			goqu.On(
				goqu.And(
					goqu.I("b.id").Eq(goqu.I("cb.building_id")),
					goqu.I("b.deleted_at").IsNull(),
				),
			),
		).
		Where(goqu.And(
			goqu.I("cb.id").Eq(id),
			goqu.I("cb.demolished_at").IsNull(),
		)).
		ScanStruct(building)

	if err != nil || !found {
		return nil, err
	}

	return building, nil
}

func (r *goquRepository) AddBuilding(companyId uint64, building *building.Building, position uint8) (*CompanyBuilding, error) {
	result, err := r.builder.
		Insert(goqu.T("companies_buildings")).
		Rows(goqu.Record{
			"name":        building.Name,
			"company_id":  companyId,
			"building_id": building.Id,
			"position":    position,
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

	return r.GetBuildingById(uint64(id))
}
