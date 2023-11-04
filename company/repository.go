package company

import (
	"api/building"
	"api/database"
	"api/resource"
	"time"

	"github.com/doug-martin/goqu/v9"
)

const (
	WAGES          = 1
	SOCIAL_CAPITAL = 2
)

type (
	Repository interface {
		Register(registration *Registration) (*Company, error)

		GetById(id uint64) (*Company, error)

		GetByEmail(email string) (*Company, error)

		GetBuildings(companyId uint64) ([]*CompanyBuilding, error)

		GetBuilding(buildingId, companyId uint64) (*CompanyBuilding, error)

		AddBuilding(companyId uint64, building *building.Building, position uint8) (*CompanyBuilding, error)

		Produce(companyId uint64, building *CompanyBuilding, item *resource.Item) (*Production, error)

		RegisterTransaction(companyId, classificationId uint64, amount int, description string) error
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
			goqu.COALESCE(goqu.SUM("t.value"), 0).As("cash"),
		).
		From(goqu.T("companies").As("c")).
		LeftJoin(
			goqu.T("transactions").As("t"),
			goqu.On(goqu.I("t.company_id").Eq(goqu.I("c.id"))),
		).
		Where(
			goqu.And(
				goqu.I("c.id").Eq(id),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		GroupBy(goqu.I("c.id")).
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

	if err = r.RegisterTransaction(
		uint64(id),
		SOCIAL_CAPITAL,
		1_000_000,
		"Initial capital",
	); err != nil {
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
			goqu.I("bp.finishes_at").As("busy_until"),
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
		LeftJoin(
			goqu.T("productions").As("bp"),
			goqu.On(
				goqu.And(
					goqu.I("bp.building_id").Eq(goqu.I("cb.id")),
					goqu.I("bp.finishes_at").Gt(goqu.L("CURRENT_TIMESTAMP")),
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

	for _, building := range buildings {
		resources, err := r.GetResources(building.Id)
		if err != nil {
			return nil, err
		}
		building.Resources = resources
	}

	return buildings, nil
}

func (r *goquRepository) GetBuilding(id, companyId uint64) (*CompanyBuilding, error) {
	building := new(CompanyBuilding)

	found, err := r.builder.
		Select(
			goqu.I("cb.id"),
			goqu.I("cb.name"),
			goqu.I("bp.finishes_at").As("busy_until"),
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
		LeftJoin(
			goqu.T("productions").As("bp"),
			goqu.On(
				goqu.And(
					goqu.I("bp.building_id").Eq(goqu.I("cb.id")),
					goqu.I("bp.finishes_at").Gt(goqu.L("CURRENT_TIMESTAMP")),
				),
			),
		).
		Where(goqu.And(
			goqu.I("cb.id").Eq(id),
			goqu.I("cb.company_id").Eq(companyId),
			goqu.I("cb.demolished_at").IsNull(),
		)).
		ScanStruct(building)

	if err != nil || !found {
		return nil, err
	}

	resources, err := r.GetResources(building.Id)
	if err != nil {
		return nil, err
	}
	building.Resources = resources

	return building, nil
}

func (r *goquRepository) GetResources(buildingId uint64) ([]*building.BuildingResource, error) {
	resources := make([]*building.BuildingResource, 0)

	err := r.builder.
		Select(
			goqu.L("? * ?", goqu.I("cb.level"), goqu.I("br.qty_per_hour")).As("qty_per_hour"),
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
		InnerJoin(
			goqu.T("companies_buildings").As("cb"),
			goqu.On(goqu.I("br.building_id").Eq(goqu.I("cb.building_id"))),
		).
		Where(goqu.I("cb.id").Eq(buildingId)).
		ScanStructs(&resources)

	for _, resource := range resources {
		requirements, err := r.resources.GetRequirements(resource.Resource.Id)
		if err != nil {
			return nil, err
		}
		resource.Resource.Requirements = requirements
	}

	return resources, err
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

	return r.GetBuilding(uint64(id), companyId)
}

func (r *goquRepository) Produce(companyId uint64, building *CompanyBuilding, item *resource.Item) (*Production, error) {
	resourceToProduce, err := building.GetResource(item.ResourceId)
	if err != nil {
		return nil, err
	}

	timeToProduce := float64(item.Qty) / (float64(resourceToProduce.QtyPerHours) / 60.0)
	finishesAt := time.Now().Add(time.Second * time.Duration(timeToProduce*60))

	result, err := r.builder.
		Insert(goqu.T("productions")).
		Rows(goqu.Record{
			"qty":         item.Qty,
			"quality":     item.Quality,
			"building_id": building.Id,
			"resource_id": item.ResourceId,
			"finishes_at": finishesAt,
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

	return r.GetProduction(uint64(id))
}

func (r *goquRepository) GetProduction(id uint64) (*Production, error) {
	production := new(Production)

	found, err := r.builder.
		Select(
			goqu.I("p.id"),
			goqu.I("p.quality"),
			goqu.I("p.finishes_at"),
			goqu.I("p.qty").As("quantity"),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
			goqu.I("r.id").As(goqu.C("resource.id")),
		).
		From(goqu.T("productions").As("p")).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("p.resource_id").Eq(goqu.I("r.id"))),
		).
		Where(goqu.I("p.id").Eq(id)).
		ScanStruct(production)

	if err != nil || !found {
		return nil, err
	}

	return production, nil
}

func (r *goquRepository) RegisterTransaction(companyId, classificationId uint64, amount int, description string) error {
	_, err := r.builder.
		Insert(goqu.T("transactions")).
		Rows(goqu.Record{
			"company_id":        companyId,
			"classification_id": classificationId,
			"description":       description,
			"value":             amount,
		}).
		Executor().
		Exec()

	return err
}
