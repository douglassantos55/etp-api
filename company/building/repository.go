package building

import (
	"api/building"
	"api/database"
	"api/resource"
	"api/warehouse"
	"context"

	"github.com/doug-martin/goqu/v9"
)

type (
	BuildingRepository interface {
		GetAll(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error)
		GetById(ctx context.Context, buildingId, companyId uint64) (*CompanyBuilding, error)
		AddBuilding(ctx context.Context, companyId uint64, inventory *warehouse.Inventory, building *building.Building, position uint8) (*CompanyBuilding, error)
	}

	buildingRepository struct {
		builder   *goqu.Database
		resources resource.Repository
		warehouse warehouse.Repository
	}
)

func NewBuildingRepository(conn *database.Connection, resources resource.Repository, warehouse warehouse.Repository) BuildingRepository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &buildingRepository{builder, resources, warehouse}
}

func (r *buildingRepository) GetAll(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error) {
	buildings := make([]*CompanyBuilding, 0)

	err := r.getSelectDataset().
		Where(goqu.And(
			goqu.I("cb.company_id").Eq(companyId),
			goqu.I("cb.demolished_at").IsNull(),
		)).
		ScanStructsContext(ctx, &buildings)

	if err != nil {
		return nil, err
	}

	for _, building := range buildings {
		resources, err := r.getResources(ctx, building.Id)
		if err != nil {
			return nil, err
		}
		building.Resources = resources
	}

	return buildings, nil
}

func (r *buildingRepository) GetById(ctx context.Context, id, companyId uint64) (*CompanyBuilding, error) {
	building := new(CompanyBuilding)

	found, err := r.getSelectDataset().
		Where(goqu.And(
			goqu.I("cb.id").Eq(id),
			goqu.I("cb.company_id").Eq(companyId),
			goqu.I("cb.demolished_at").IsNull(),
		)).
		ScanStructContext(ctx, building)

	if err != nil || !found {
		return nil, err
	}

	resources, err := r.getResources(ctx, building.Id)
	if err != nil {
		return nil, err
	}
	building.Resources = resources

	return building, nil
}

func (r *buildingRepository) AddBuilding(ctx context.Context, companyId uint64, inventory *warehouse.Inventory, buildingToConstruct *building.Building, position uint8) (*CompanyBuilding, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = r.warehouse.UpdateInventory(&database.DB{TxDatabase: tx}, inventory)
	if err != nil {
		return nil, err
	}

	result, err := tx.
		Insert(goqu.T("companies_buildings")).
		Rows(goqu.Record{
			"position":    position,
			"company_id":  companyId,
			"building_id": buildingToConstruct.Id,
			"name":        buildingToConstruct.Name,
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

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetById(ctx, uint64(id), companyId)
}

func (r *buildingRepository) getResources(ctx context.Context, buildingId uint64) ([]*building.BuildingResource, error) {
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

func (r *buildingRepository) getSelectDataset() *goqu.SelectDataset {
	return r.builder.
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
					goqu.I("bp.canceled_at").IsNull(),
					goqu.I("bp.building_id").Eq(goqu.I("cb.id")),
					goqu.I("bp.finishes_at").Gt(goqu.L("CURRENT_TIMESTAMP")),
				),
			),
		)
}
