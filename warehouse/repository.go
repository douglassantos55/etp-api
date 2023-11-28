package warehouse

import (
	"api/database"
	"context"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

type Repository interface {
	// Fetches the inventory of a company
	FetchInventory(ctx context.Context, companyId uint64) (*Inventory, error)

	// Updates the inventory
	UpdateInventory(db *database.DB, inventory *Inventory) error
}

type goquRepository struct {
	builder *goqu.Database
}

// Creates warehouse repository
func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) FetchInventory(ctx context.Context, companyId uint64) (*Inventory, error) {
	items := make([]*StockItem, 0)

	err := r.builder.
		Select(
			goqu.I("i.quality").As("quality"),
			goqu.SUM("i.quantity").As("quantity"),
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
			goqu.I("c.id").As(goqu.C("resource.category.id")),
			goqu.I("c.name").As(goqu.C("resource.category.name")),
			goqu.L("? / ?", goqu.SUM(goqu.L("? * ?", goqu.I("i.sourcing_cost"), goqu.I("i.quantity"))), goqu.SUM(goqu.I("i.quantity"))).As("sourcing_cost"),
		).
		From(goqu.T("inventories").As("i")).
		InnerJoin(goqu.T("resources").As("r"), goqu.On(goqu.I("i.resource_id").Eq(goqu.I("r.id")))).
		InnerJoin(goqu.T("categories").As("c"), goqu.On(
			goqu.And(
				goqu.I("r.category_id").Eq(goqu.I("c.id")),
				goqu.I("c.deleted_at").IsNull(),
			),
		)).
		Where(goqu.I("i.company_id").Eq(companyId)).
		GroupBy(goqu.I("r.id"), goqu.I("i.quality")).
		// make sure q0 comes before q1 so that it's consumed first
		Order(goqu.I("i.quality").Asc()).
		ScanStructsContext(ctx, &items)

	if err != nil {
		return nil, err
	}

	return &Inventory{companyId, items}, nil
}

func (r *goquRepository) UpdateInventory(db *database.DB, inventory *Inventory) error {
	for _, item := range inventory.Items {
		if item.Qty == 0 {
			if err := r.removeStock(db, inventory.CompanyId, item); err != nil {
				return err
			}
		} else {
			if updated, err := r.updateStock(db, inventory.CompanyId, item); err != nil || updated == 0 {
				if err := r.insertStock(db, inventory.CompanyId, item); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *goquRepository) removeStock(tx *database.DB, companyId uint64, item *StockItem) error {
	_, err := tx.
		Delete(goqu.T("inventories")).
		Where(goqu.And(
			goqu.I("quality").Eq(item.Quality),
			goqu.I("company_id").Eq(companyId),
			goqu.I("resource_id").Eq(item.Resource.Id),
		)).
		Executor().
		Exec()

	return err
}

func (r *goquRepository) updateStock(tx *database.DB, companyId uint64, item *StockItem) (int64, error) {
	result, err := tx.
		Update(goqu.T("inventories")).
		Set(goqu.Record{"quantity": item.Qty}).
		Where(goqu.And(
			goqu.I("quality").Eq(item.Quality),
			goqu.I("company_id").Eq(companyId),
			goqu.I("resource_id").Eq(item.Resource.Id),
		)).
		Executor().
		Exec()

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *goquRepository) insertStock(tx *database.DB, companyId uint64, item *StockItem) error {
	_, err := tx.
		Insert(goqu.T("inventories")).
		Rows(goqu.Record{
			"sourcing_cost": item.Cost,
			"quantity":      item.Qty,
			"quality":       item.Quality,
			"company_id":    companyId,
			"resource_id":   item.Resource.Id,
		}).
		Executor().
		Exec()

	return err
}
