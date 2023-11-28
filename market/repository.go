package market

import (
	"api/company"
	"api/database"
	"api/resource"
	"api/server"
	"api/warehouse"
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		GetById(ctx context.Context, orderId uint64) (*Order, error)
		GetByResource(ctx context.Context, resourceId uint64, quality uint8) ([]*Order, error)
		PlaceOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) (*Order, error)
		CancelOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) error
		Purchase(ctx context.Context, purchase *Purchase, companyId uint64) ([]*warehouse.StockItem, error)
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
		Select(
			goqu.I("o.id"),
			goqu.I("o.price"),
			goqu.I("o.quality"),
			goqu.I("o.quantity"),
			goqu.I("o.market_fee"),
			goqu.I("o.sourcing_cost"),
			goqu.I("o.purchased_at").As("last_purchase"),
			goqu.I("c.id").As(goqu.C("company.id")),
			goqu.I("c.name").As(goqu.C("company.name")),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
		).
		From(goqu.T("orders").As("o")).
		InnerJoin(
			goqu.T("companies").As("c"),
			goqu.On(goqu.I("o.company_id").Eq(goqu.I("c.id"))),
		).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("o.resource_id").Eq(goqu.I("r.id"))),
		).
		Where(goqu.And(
			goqu.I("o.id").Eq(orderId),
			goqu.I("o.quantity").Gt(0),
			goqu.I("o.canceled_at").IsNull(),
		)).
		ScanStructContext(ctx, order)

	if err != nil || !found {
		return nil, err
	}

	return order, nil
}

func (r *goquRepository) GetByResource(ctx context.Context, resourceId uint64, quality uint8) ([]*Order, error) {
	orders := make([]*Order, 0)

	err := r.builder.
		Select(
			goqu.I("o.id"),
			goqu.I("o.price"),
			goqu.I("o.quality"),
			goqu.I("o.quantity"),
			goqu.I("o.market_fee"),
			goqu.I("o.sourcing_cost"),
			goqu.I("o.purchased_at").As("last_purchase"),
			goqu.I("c.id").As(goqu.C("company.id")),
			goqu.I("c.name").As(goqu.C("company.name")),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
		).
		From(goqu.T("orders").As("o")).
		InnerJoin(
			goqu.T("companies").As("c"),
			goqu.On(goqu.I("o.company_id").Eq(goqu.I("c.id"))),
		).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("o.resource_id").Eq(goqu.I("r.id"))),
		).
		Where(goqu.And(
			goqu.I("o.canceled_at").IsNull(),
			goqu.I("o.quantity").Gt(0),
			goqu.I("o.quality").Gte(quality),
			goqu.I("o.resource_id").Eq(resourceId),
		)).
		Order(goqu.I("o.price").Asc()).
		ScanStructsContext(ctx, &orders)

	if err != nil {
		return nil, err
	}

	return orders, nil
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
		order.Company.Id,
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
			goqu.I("company_id").Eq(order.Company.Id),
		)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *goquRepository) Purchase(ctx context.Context, purchase *Purchase, companyId uint64) ([]*warehouse.StockItem, error) {
	orders, err := r.GetByResource(ctx, purchase.ResourceId, purchase.Quality)
	if err != nil {
		return nil, err
	}

	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	total := 0
	remaining := purchase.Quantity
	purchasedItems := make([]*warehouse.StockItem, 0)

	for _, order := range orders {
		if order.Quantity >= remaining {
			item, err := r.partialPurchase(tx, order, remaining, companyId)
			if err != nil {
				return nil, err
			}

			remaining = 0
			total += int(order.Price) * int(remaining)
			purchasedItems = append(purchasedItems, item)

			break
		} else {
			remaining -= order.Quantity
			total += int(order.Price) * int(order.Quantity)

			item, err := r.fullPurchase(tx, order, companyId)
			if err != nil {
				return nil, err
			}

			purchasedItems = append(purchasedItems, item)
		}
	}

	if remaining > 0 {
		return nil, server.NewBusinessRuleError("not enough market orders")
	}

	company, err := r.companyRepo.GetById(ctx, companyId)
	if err != nil {
		return nil, err
	}

	if company.AvailableCash < total {
		return nil, server.NewBusinessRuleError("not enough cash")
	}

	inventory, err := r.warehouseRepo.FetchInventory(ctx, companyId)
	if err != nil {
		return nil, err
	}

	inventory.IncrementStock(purchasedItems)

	dbTx := &database.DB{TxDatabase: tx}
	if err := r.warehouseRepo.UpdateInventory(dbTx, inventory); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return purchasedItems, nil
}

func (r *goquRepository) fullPurchase(tx *goqu.TxDatabase, order *Order, companyId uint64) (*warehouse.StockItem, error) {
	item := &warehouse.StockItem{
		Cost: order.SourcingCost,
		Item: &resource.Item{
			Qty:      order.Quantity,
			Quality:  order.Quality,
			Resource: order.Resource,
		},
	}

	if err := r.registerPurchaseTransactions(tx, order, order.Quantity, companyId); err != nil {
		return nil, err
	}

	order.Quantity = 0
	if err := r.updateOrder(tx, order); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *goquRepository) partialPurchase(tx *goqu.TxDatabase, order *Order, quantity, companyId uint64) (*warehouse.StockItem, error) {
	if err := r.registerPurchaseTransactions(tx, order, quantity, companyId); err != nil {
		return nil, err
	}

	order.Quantity -= quantity
	if err := r.updateOrder(tx, order); err != nil {
		return nil, err
	}

	return &warehouse.StockItem{
		Cost: order.SourcingCost,
		Item: &resource.Item{
			Qty:      quantity,
			Quality:  order.Quality,
			Resource: order.Resource,
		},
	}, nil
}

func (r *goquRepository) registerPurchaseTransactions(tx *goqu.TxDatabase, order *Order, quantity, companyId uint64) error {
	total := int(order.Price) * int(quantity)

	if order.LastPurchase == nil {
		total -= int(order.MarketFee)
	}

	if err := r.companyRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		companyId,
		company.MARKET_PURCHASE,
		total*-1,
		fmt.Sprintf("Purchase of %dx %s on market", quantity, order.Resource.Name),
	); err != nil {
		return err
	}

	if err := r.companyRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		order.Company.Id,
		company.MARKET_SALE,
		total,
		fmt.Sprintf("Sale of %dx %s on market", quantity, order.Resource.Name),
	); err != nil {
		return err
	}

	return nil
}

func (r *goquRepository) updateOrder(tx *goqu.TxDatabase, order *Order) error {
	_, err := tx.
		Update(goqu.T("orders")).
		Set(goqu.Record{
			"quantity":     order.Quantity,
			"purchased_at": time.Now(),
		}).
		Where(goqu.I("id").Eq(order.Id)).
		Executor().
		Exec()

	return err
}
