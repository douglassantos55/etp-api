package market

import (
	"api/company"
	"api/resource"
	"api/server"
	"api/warehouse"
	"context"
	"time"
)

const TRANSPORT_FEE_PERCENTAGE = 0.005

type (
	Order struct {
		Id           uint64     `db:"id" json:"id" goqu:"skipinsert,skipupdate" validate:"-"`
		Price        uint64     `db:"price" json:"price" validate:"required,gt=0"`
		Quality      uint8      `db:"quality" json:"quality" validate:"gte=0"`
		Quantity     uint64     `db:"quantity" json:"quantity" validate:"required,gte=1"`
		MarketFee    uint64     `db:"market_fee" json:"market_fee" validate:"-"`
		TransportFee uint64     `db:"transport_fee" json:"transport_fee" validate:"-"`
		CompanyId    uint64     `db:"company_id" json:"company_id,omitempty" validate:"-"`
		ResourceId   uint64     `db:"resource_id" json:"resource_id,omitempty" validate:"required"`
		SourcingCost uint64     `db:"sourcing_cost" json:"-" validate:"-"`
		LastPurchase *time.Time `db:"last_purchase" json:"-" validate:"-"`

		Resource *resource.Resource `db:"resource" json:"resource" validate:"-"`
		Company  *company.Company   `db:"company" json:"company" validate:"-"`
	}

	Service interface {
		GetById(ctx context.Context, orderId uint64) (*Order, error)
		PlaceOrder(ctx context.Context, order *Order) (*Order, error)
		CancelOrder(ctx context.Context, order *Order) error
	}

	service struct {
		repository   Repository
		companySvc   company.Service
		warehouseSvc warehouse.Service
	}
)

func NewService(repository Repository, companySvc company.Service, warehouseSvc warehouse.Service) Service {
	return &service{repository, companySvc, warehouseSvc}
}

func (s *service) GetById(ctx context.Context, orderId uint64) (*Order, error) {
	return s.repository.GetById(ctx, orderId)
}

func (s *service) PlaceOrder(ctx context.Context, order *Order) (*Order, error) {
	inventory, err := s.warehouseSvc.GetInventory(ctx, order.CompanyId)
	if err != nil {
		return nil, err
	}

	orderItem := []*resource.Item{{
		Qty:      order.Quantity,
		Quality:  order.Quality,
		Resource: &resource.Resource{Id: order.ResourceId},
	}}

	if !inventory.HasResources(orderItem) {
		return nil, server.NewBusinessRuleError("not enough resources")
	}

	orderCompany, err := s.companySvc.GetById(ctx, order.CompanyId)
	if err != nil {
		return nil, err
	}

	sourcingCost := inventory.ReduceStock(orderItem)

	order.SourcingCost = sourcingCost
	order.TransportFee = uint64(float64(sourcingCost*order.Quantity) * TRANSPORT_FEE_PERCENTAGE)

	if orderCompany.AvailableCash < int(order.TransportFee) {
		return nil, server.NewBusinessRuleError("not enough cash to pay transport fee")
	}

	return s.repository.PlaceOrder(ctx, order, inventory)
}

func (s *service) CancelOrder(ctx context.Context, order *Order) error {
	inventory, err := s.warehouseSvc.GetInventory(ctx, order.CompanyId)
	if err != nil {
		return err
	}

	inventory.IncrementStock([]*warehouse.StockItem{
		{
			Item: &resource.Item{
				Qty:      order.Quantity,
				Quality:  order.Quality,
				Resource: &resource.Resource{Id: order.Resource.Id},
			},
			Cost: order.SourcingCost,
		},
	})

	return s.repository.CancelOrder(ctx, order, inventory)
}
