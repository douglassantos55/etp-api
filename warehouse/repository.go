package warehouse

import "github.com/doug-martin/goqu/v9"

type Repository interface {
	FetchInventory(companyId uint64) ([]*Resource, error)
}

type goquRepository struct {
	builder *goqu.Database
}

func NewGoquRepository(builder *goqu.Database) (*goquRepository, error) {
	return &goquRepository{builder}, nil
}

// Fetches the inventory of a company
func (r *goquRepository) FetchInventory(companyId uint64) ([]*Resource, error) {
	var resources []*Resource

	err := r.builder.
		Select("r.id", "r.name", "r.image", goqu.SUM("i.quantity").As("quantity")).
		Select(goqu.L("? / ?", goqu.SUM(goqu.I("i.sourcing_cost")), goqu.SUM(goqu.I("i.quantity")).As("sourcing_cost"))).
		From(goqu.T("inventories").As("i")).
		InnerJoin(goqu.T("resources").As("r"), goqu.On(goqu.I("i.resource_id").Eq(goqu.I("r.id")))).
		Where(goqu.I("i.company_id").Eq(companyId)).
		GroupBy(goqu.I("r.id")).
		ScanStructs(&resources)

	if err != nil {
		return nil, err
	}

	return resources, nil
}
