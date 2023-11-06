package warehouse

import (
	"api/database"
	"api/resource"
	"context"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

type Repository interface {
	// Fetches the inventory of a company
	FetchInventory(ctx context.Context, companyId uint64) (*Inventory, error)

	// Reduces the stock for the given resources
	ReduceStock(db *database.DB, companyId uint64, inventory *Inventory, resources []*resource.Item) error
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
		ScanStructsContext(ctx, &items)

	if err != nil {
		return nil, err
	}

	return &Inventory{items}, nil
}

func (r *goquRepository) ReduceStock(db *database.DB, companyId uint64, inventory *Inventory, resources []*resource.Item) error {
	for _, resource := range resources {
		for _, item := range inventory.Items {
			if item.Resource.Id == resource.Resource.Id && item.Quality == resource.Quality {
				item.Qty -= resource.Qty
				if item.Qty == 0 {
					if err := r.removeStock(db, companyId, item); err != nil {
						return err
					}
				} else {
					if err := r.updateStock(db, companyId, item); err != nil {
						return err
					}
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

func (r *goquRepository) updateStock(tx *database.DB, companyId uint64, item *StockItem) error {
	_, err := tx.
		Update(goqu.T("inventories")).
		Set(goqu.Record{"quantity": item.Qty}).
		Where(goqu.And(
			goqu.I("quality").Eq(item.Quality),
			goqu.I("company_id").Eq(companyId),
			goqu.I("resource_id").Eq(item.Resource.Id),
		)).
		Executor().
		Exec()

	return err
}

// Get the stock of a company's resource, grouping by quality
func (r *goquRepository) FetchStock(companyId, resourceId uint64) ([]any, error) {
	return nil, nil
}
