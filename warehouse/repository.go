package warehouse

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

type Repository interface {
	// Fetches the inventory of a company
	FetchInventory(companyId uint64) ([]*StockItem, error)
}

type goquRepository struct {
	builder *goqu.Database
}

// Creates warehouse repository
func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) FetchInventory(companyId uint64) ([]*StockItem, error) {
	items := make([]*StockItem, 0)

	err := r.builder.
		Select(
			goqu.I("r.id").As(goqu.C("resource.id")),
			goqu.I("r.name").As(goqu.C("resource.name")),
			goqu.I("r.image").As(goqu.C("resource.image")),
			goqu.I("c.id").As(goqu.C("resource.category.id")),
			goqu.I("c.name").As(goqu.C("resource.category.name")),
			goqu.SUM("i.quantity").As("quantity"),
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
		GroupBy(goqu.I("r.id")).
		ScanStructs(&items)

	if err != nil {
		return nil, err
	}

	return items, nil
}

// Get the stock of a company's resource, grouping by quality
func (r *goquRepository) FetchStock(companyId, resourceId uint64) ([]any, error) {
	return nil, nil
}
