package warehouse

import (
	"api/database"
	"api/resource"
	"context"
	"errors"
)

type fakeRepository struct {
	data map[uint64]*Inventory
}

func NewFakeRepository() Repository {
	data := map[uint64]*Inventory{
		1: {Items: []*StockItem{
			{Cost: 137, Item: &resource.Item{Quality: 0, Qty: 100, Resource: &resource.Resource{Id: 1}}},
			{Cost: 47, Item: &resource.Item{Quality: 1, Qty: 1000, Resource: &resource.Resource{Id: 3}}},
			{Cost: 1553, Item: &resource.Item{Quality: 0, Qty: 700, Resource: &resource.Resource{Id: 2}}},
		}},
		2: {Items: []*StockItem{
			{Cost: 525, Item: &resource.Item{Quality: 1, Qty: 50, Resource: &resource.Resource{Id: 1}}},
		}},
	}
	return &fakeRepository{data}
}

func (r *fakeRepository) FetchInventory(ctx context.Context, companyId uint64) (*Inventory, error) {
	inventory, ok := r.data[companyId]
	if !ok {
		return nil, errors.New("inventory not found")
	}
	return inventory, nil
}

func (r *fakeRepository) ReduceStock(tx *database.DB, companyId uint64, inventory *Inventory, items []*resource.Item) error {
	for _, item := range items {
		for _, inv := range inventory.Items {
			isResource := item.Resource.Id == inv.Resource.Id
			isQuality := item.Quality == inv.Quality

			if isResource && isQuality {
				inv.Qty -= item.Qty
			}
		}
	}
	return nil
}

func (r *fakeRepository) IncrementStock(tx *database.DB, companyId uint64, resources []*resource.Item) error {
	return nil
}
