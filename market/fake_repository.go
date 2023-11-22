package market

import (
	"api/warehouse"
	"context"
)

type fakeRepository struct {
	lastId uint64
	orders map[uint64]*Order
}

func NewFakeRepository() Repository {
	return &fakeRepository{
		lastId: 0,
		orders: make(map[uint64]*Order),
	}
}

func (r *fakeRepository) PlaceOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) (*Order, error) {
	r.lastId++

	order.Id = r.lastId
	r.orders[r.lastId] = order

	return order, nil
}
