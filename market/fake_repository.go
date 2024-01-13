package market

import (
	"api/company"
	"api/resource"
	"api/server"
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
			Company:      &company.Company{Id: 1},
		},
		2: {
			Id:           2,
			Price:        1823,
			Quality:      0,
			Quantity:     150,
			CompanyId:    2,
			ResourceId:   2,
			SourcingCost: 1553,
			TransportFee: 1164,
			MarketFee:    8203,
			Resource:     &resource.Resource{Id: 2},
			Company:      &company.Company{Id: 2},
		},
	}

	return &fakeRepository{
		lastId: 2,
		orders: orders,
	}
}

func (r *fakeRepository) GetById(ctx context.Context, orderId uint64) (*Order, error) {
	return r.orders[orderId], nil
}

func (r *fakeRepository) GetByResource(ctx context.Context, resourceId uint64, quality uint8) ([]*Order, error) {
	orders := make([]*Order, 0)

	for _, order := range r.orders {
		if order.ResourceId == resourceId && order.Quality >= quality {
			orders = append(orders, order)
		}
	}

	return orders, nil
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

func (r *fakeRepository) Purchase(ctx context.Context, purchase *Purchase, companyId uint64) ([]*warehouse.StockItem, []*Order, error) {
	orders, err := r.GetByResource(ctx, purchase.ResourceId, purchase.Quality)
	if err != nil {
		return nil, nil, err
	}

	remaining := purchase.Quantity
	purchasedOrders := make([]*Order, 0)
	items := make([]*warehouse.StockItem, 0)

	for _, order := range orders {
		if order.Quantity > remaining {
			items = append(items, &warehouse.StockItem{
				Cost: order.SourcingCost,
				Item: &resource.Item{
					Qty:        remaining,
					Quality:    order.Quality,
					Resource:   order.Resource,
					ResourceId: order.ResourceId,
				},
			})
			remaining = 0
			purchasedOrders = append(purchasedOrders, order)
			break
		} else {
			remaining -= order.Quantity
			items = append(items, &warehouse.StockItem{
				Cost: order.SourcingCost,
				Item: &resource.Item{
					Qty:        order.Quantity,
					Quality:    order.Quality,
					Resource:   order.Resource,
					ResourceId: order.ResourceId,
				},
			})
			purchasedOrders = append(purchasedOrders, order)
		}
	}

	if remaining > 0 {
		return nil, nil, server.NewBusinessRuleError("nope")
	}

	return items, purchasedOrders, nil
}
