package market

import (
	"api/company"
	"api/database"
	"api/warehouse"
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		GetById(ctx context.Context, orderId uint64) (*Order, error)
		PlaceOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) (*Order, error)
		CancelOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) error
	}

	goquRepository struct {
		builder       *goqu.Database
		companyRepo   company.Repository
		warehouseRepo warehouse.Repository
	}
)

func NewRepository(conn *database.Connection, companyRepo company.Repository, warehouseRepo warehouse.Repository) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder, companyRepo, warehouseRepo}
}

func (r *goquRepository) GetById(ctx context.Context, orderId uint64) (*Order, error) {
	order := new(Order)

	found, err := r.builder.
		Select().
		From(goqu.T("orders")).
		Where(goqu.And()).
		ScanStructContext(ctx, order)

	if err != nil || !found {
		return nil, err
	}

	return order, nil
}

func (r *goquRepository) PlaceOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) (*Order, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	dbTx := &database.DB{TxDatabase: tx}
	if err := r.warehouseRepo.UpdateInventory(dbTx, inventory); err != nil {
		return nil, err
	}

	if err := r.companyRepo.RegisterTransaction(
		dbTx,
		order.CompanyId,
		company.TRANSPORT_FEE,
		int(order.TransportFee)*-1,
		"Market transport fee",
	); err != nil {
		return nil, err
	}

	result, err := tx.
		Insert(goqu.T("orders")).
		Rows(goqu.Record{
			"price":         order.Price,
			"quality":       order.Quality,
			"quantity":      order.Quantity,
			"company_id":    order.CompanyId,
			"resource_id":   order.ResourceId,
			"market_fee":    order.MarketFee,
			"transport_fee": order.TransportFee,
			"sourcing_cost": order.SourcingCost,
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

	order.Id = uint64(id)
	return order, nil
}

func (r *goquRepository) CancelOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	dbTx := &database.DB{TxDatabase: tx}
	if err := r.warehouseRepo.UpdateInventory(dbTx, inventory); err != nil {
		return err
	}

	if err := r.companyRepo.RegisterTransaction(
		dbTx,
		order.CompanyId,
		company.REFUNDS,
		int(order.TransportFee),
		"Market transport fee refund",
	); err != nil {
		return err
	}

	_, err = tx.
		Update(goqu.T("orders")).
		Set(goqu.Record{
			"canceled_at": time.Now(),
		}).
		Where(goqu.And(
			goqu.I("id").Eq(order.Id),
			goqu.I("company_id").Eq(order.CompanyId),
		)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	return tx.Commit()
}
