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
			goqu.I("resource.id").As(goqu.C("resource.id")),
			goqu.I("resource.name").As(goqu.C("resource.name")),
			goqu.I("resource.image").As(goqu.C("resource.image")),
			goqu.SUM("i.quantity").As("quantity"),
			goqu.L("? / ?", goqu.SUM(goqu.L("? * ?", goqu.I("i.sourcing_cost"), goqu.I("i.quantity"))), goqu.SUM(goqu.I("i.quantity"))).As("sourcing_cost"),
		).
		From(goqu.T("inventories").As("i")).
		InnerJoin(goqu.T("resources").As("resource"), goqu.On(goqu.I("i.resource_id").Eq(goqu.I("resource.id")))).
		Where(goqu.I("i.company_id").Eq(companyId)).
		GroupBy(goqu.I("resource.id")).
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
