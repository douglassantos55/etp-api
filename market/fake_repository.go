package market

import (
	"api/resource"
	"api/warehouse"
	"context"
)

type fakeRepository struct {
	lastId uint64
	orders map[uint64]*Order
}

func NewFakeRepository() Repository {
	orders := map[uint64]*Order{
		1: {
			Id:           1,
			Price:        1823,
			Quality:      0,
			Quantity:     150,
			CompanyId:    1,
			ResourceId:   2,
			SourcingCost: 1553,
			TransportFee: 1164,
			MarketFee:    8203,
			Resource:     &resource.Resource{Id: 2},
		},
	}

	return &fakeRepository{
		lastId: 1,
		orders: orders,
	}
}

func (r *fakeRepository) GetById(ctx context.Context, orderId uint64) (*Order, error) {
	return r.orders[orderId], nil
}

func (r *fakeRepository) PlaceOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) (*Order, error) {
	r.lastId++

	order.Id = r.lastId
	r.orders[r.lastId] = order

	return order, nil
}

func (r *fakeRepository) CancelOrder(ctx context.Context, order *Order, inventory *warehouse.Inventory) error {
	return nil
}
