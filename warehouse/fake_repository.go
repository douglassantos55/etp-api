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
			{Cost: 153553, Item: &resource.Item{Quality: 2, Qty: 1700, Resource: &resource.Resource{Id: 4}}},
		}},
		2: {Items: []*StockItem{
			{Cost: 525, Item: &resource.Item{Quality: 1, Qty: 50, Resource: &resource.Resource{Id: 1}}},
			{Cost: 1553, Item: &resource.Item{Quality: 0, Qty: 700, Resource: &resource.Resource{Id: 2}}},
		}},
	}
	return &fakeRepository{data}
}

func (r *fakeRepository) FetchInventory(ctx context.Context, companyId uint64) (*Inventory, error) {
	inventory, ok := r.data[companyId]
	if !ok {
		return nil, errors.New("inventory not found")
	}
	inventory.CompanyId = companyId
	return inventory, nil
}

func (r *fakeRepository) UpdateInventory(db *database.DB, inventory *Inventory) error {
	r.data[inventory.CompanyId] = inventory
	return nil
}
