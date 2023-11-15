package building

import (
	"api/building"
	"api/database"
	"api/resource"
	"api/warehouse"
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type (
	BuildingRepository interface {
		GetAll(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error)
		GetById(ctx context.Context, buildingId, companyId uint64) (*CompanyBuilding, error)
		AddBuilding(ctx context.Context, companyId uint64, inventory *warehouse.Inventory, building *building.Building, position uint8) (*CompanyBuilding, error)
		Demolish(ctx context.Context, companyId, building uint64) error
		Upgrade(ctx context.Context, inventory *warehouse.Inventory, building *CompanyBuilding) error
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
		Where(r.getSelectConditions(companyId)).
		ScanStructsContext(ctx, &buildings)

	if err != nil {
		return nil, err
	}

	for _, building := range buildings {
		resources, err := r.getResources(ctx, building.Id)
		if err != nil {
			return nil, err
		}

		requirements, err := r.getRequirements(ctx, building.Id)
		if err != nil {
			return nil, err
		}

		building.Resources = resources
		building.Requirements = requirements
	}

	return buildings, nil
}

func (r *buildingRepository) GetById(ctx context.Context, id, companyId uint64) (*CompanyBuilding, error) {
	companyBuilding := new(CompanyBuilding)

	found, err := r.getSelectDataset().
		Where(goqu.And(
			r.getSelectConditions(companyId).Append(
				goqu.I("cb.id").Eq(id),
			),
		)).
		ScanStructContext(ctx, companyBuilding)

	if err != nil || !found {
		return nil, err
	}

	resources, err := r.getResources(ctx, companyBuilding.Id)
	if err != nil {
		return nil, err
	}

	requirements, err := r.getRequirements(ctx, companyBuilding.Id)
	if err != nil {
		return nil, err
	}

	companyBuilding.Resources = resources
	companyBuilding.Requirements = requirements

	return companyBuilding, nil
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
			"position":     position,
			"company_id":   companyId,
			"building_id":  buildingToConstruct.Id,
			"name":         buildingToConstruct.Name,
			"completes_at": time.Now().Add(time.Minute * time.Duration(*buildingToConstruct.Downtime)),
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

func (r *buildingRepository) Demolish(ctx context.Context, companyId, buildingId uint64) error {
	_, err := r.builder.
		Update(goqu.T("companies_buildings")).
		Set(goqu.Record{"demolished_at": time.Now()}).
		Where(goqu.And(
			goqu.I("id").Eq(buildingId),
			goqu.I("company_id").Eq(companyId),
		)).
		Executor().
		Exec()

	return err
}

func (r *buildingRepository) Upgrade(ctx context.Context, inventory *warehouse.Inventory, companyBuilding *CompanyBuilding) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = r.warehouse.UpdateInventory(&database.DB{TxDatabase: tx}, inventory)
	if err != nil {
		return err
	}

	_, err = tx.Update(goqu.T("companies_buildings")).
		Set(goqu.Record{
			"level":        companyBuilding.Level,
			"completes_at": *companyBuilding.CompletesAt,
		}).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
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

func (r *buildingRepository) getRequirements(ctx context.Context, buildingId uint64) ([]*resource.Item, error) {
	requirements := make([]*resource.Item, 0)

	err := r.builder.
		Select(
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
			goqu.L("? - 1", goqu.I("cb.level")).As("quality"),
			goqu.L("? * ?", goqu.I("req.qty"), goqu.I("cb.level")).As("quantity"),
		).
		From(goqu.T("buildings_requirements").As("req")).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("req.resource_id").Eq(goqu.I("r.id"))),
		).
		InnerJoin(
			goqu.T("companies_buildings").As("cb"),
			goqu.On(goqu.I("cb.building_id").Eq(goqu.I("req.building_id"))),
		).
		Where(goqu.I("cb.id").Eq(buildingId)).
		ScanStructsContext(ctx, &requirements)

	return requirements, err
}

func (r *buildingRepository) getSelectDataset() *goqu.SelectDataset {
	return r.builder.
		Select(
			// building generic information
			goqu.I("cb.id"),
			goqu.I("cb.name"),
			goqu.L("? * ?", goqu.I("b.downtime"), goqu.I("cb.level")).As("downtime"),
			goqu.L("? * ?", goqu.I("b.wages_per_hour"), goqu.I("cb.level")).As("wages_per_hour"),
			goqu.L("? * ?", goqu.I("b.admin_per_hour"), goqu.I("cb.level")).As("admin_per_hour"),
			goqu.L("? * ?", goqu.I("b.maintenance_per_hour"), goqu.I("cb.level")).As("maintenance_per_hour"),

			// company specific information
			goqu.I("cb.level"),
			goqu.I("cb.position"),
			goqu.I("cb.completes_at"),
			goqu.I("bp.finishes_at").As("busy_until"),
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

func (r *buildingRepository) getSelectConditions(companyId uint64) exp.ExpressionList {
	return goqu.And(
		goqu.I("cb.company_id").Eq(companyId),
		goqu.I("cb.demolished_at").IsNull(),
	)
}
