package market

import (
	"api/company"
	"api/database"
	"api/warehouse"
	"context"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		PlaceOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) (*Order, error)
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
		}).Executor().Exec()

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
