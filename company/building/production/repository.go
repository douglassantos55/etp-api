package production

import (
	"api/company"
	"api/company/building"
	"api/database"
	"api/warehouse"
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
)

type (
	ProductionRepository interface {
		GetProduction(ctx context.Context, id, buildingId, companyId uint64) (*Production, error)
		SaveProduction(ctx context.Context, production *Production, inventory *warehouse.Inventory, companyId uint64) (*Production, error)
		CancelProduction(ctx context.Context, production *Production, inventory *warehouse.Inventory) error
		CollectResource(ctx context.Context, production *Production, inventory *warehouse.Inventory) error
	}

	productionRepository struct {
		builder   *goqu.Database
		company   company.Repository
		building  building.BuildingRepository
		warehouse warehouse.Repository
	}
)

func NewProductionRepository(conn *database.Connection, company company.Repository, building building.BuildingRepository, warehouse warehouse.Repository) ProductionRepository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &productionRepository{builder, company, building, warehouse}
}

func (r *productionRepository) SaveProduction(ctx context.Context, production *Production, inventory *warehouse.Inventory, companyId uint64) (*Production, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	dbTx := &database.DB{TxDatabase: tx}
	if err := r.warehouse.UpdateInventory(dbTx, inventory); err != nil {
		return nil, err
	}

	if err := r.company.RegisterTransaction(
		dbTx,
		companyId,
		company.WAGES,
		int(-production.ProductionCost),
		fmt.Sprintf("Production of %s", production.Resource.Name),
	); err != nil {
		return nil, err
	}

	result, err := tx.
		Insert(goqu.T("productions")).
		Rows(goqu.Record{
			"qty":           production.Qty,
			"sourcing_cost": production.CalculateSourcingCost(),
			"quality":       production.Quality,
			"building_id":   production.Building.Id,
			"resource_id":   production.Resource.Id,
			"finishes_at":   production.FinishesAt,
		}).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	production.Id = uint64(id)
	production.SourcingCost = production.CalculateSourcingCost()

	return production, nil
}

func (r *productionRepository) GetProduction(ctx context.Context, id, buildingId, companyId uint64) (*Production, error) {
	production := new(Production)

	found, err := r.builder.
		Select(
			goqu.I("p.id"),
			goqu.I("p.quality"),
			goqu.I("p.finishes_at"),
			goqu.I("p.created_at"),
			goqu.I("p.collected_at"),
			goqu.I("p.canceled_at"),
			goqu.I("p.sourcing_cost"),
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
		InnerJoin(
			goqu.T("companies_buildings").As("cb"),
			goqu.On(goqu.And(
				goqu.I("p.building_id").Eq(goqu.I("cb.id")),
				goqu.I("cb.company_id").Eq(companyId),
			)),
		).
		Where(goqu.And(
			goqu.I("p.id").Eq(id)),
			goqu.I("p.building_id").Eq(buildingId),
		).
		ScanStructContext(ctx, production)

	if err != nil || !found {
		return nil, err
	}

	building, err := r.building.GetById(ctx, buildingId, companyId)
	if err != nil {
		return nil, err
	}

	production.Building = building
	return production, nil
}

func (r *productionRepository) CancelProduction(ctx context.Context, production *Production, inventory *warehouse.Inventory) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = r.warehouse.UpdateInventory(&database.DB{TxDatabase: tx}, inventory)
	if err != nil {
		return err
	}

	_, err = tx.Update(goqu.T("productions")).
		Set(goqu.Record{
			"canceled_at":  production.CanceledAt,
			"collected_at": production.CanceledAt,
		}).
		Where(goqu.I("id").Eq(production.Id)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *productionRepository) CollectResource(ctx context.Context, production *Production, inventory *warehouse.Inventory) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = r.warehouse.UpdateInventory(&database.DB{TxDatabase: tx}, inventory)
	if err != nil {
		return err
	}

	_, err = tx.Update(goqu.T("productions")).
		Set(goqu.Record{"collected_at": production.LastCollection}).
		Where(goqu.I("id").Eq(production.Id)).
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
