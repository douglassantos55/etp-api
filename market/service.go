package market

import (
	"api/company"
	"api/notification"
	"api/resource"
	"api/server"
	"api/warehouse"
	"context"
	"log"
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

	Purchase struct {
		ResourceId uint64 `json:"resource_id" validate:"required"`
		Quantity   uint64 `json:"quantity" validate:"required"`
		Quality    uint8  `json:"quality" validate:"gte=0"`
	}

	Service interface {
		GetById(ctx context.Context, orderId uint64) (*Order, error)
		GetByResource(ctx context.Context, resourceId, quality uint64) ([]*Order, error)
		PlaceOrder(ctx context.Context, order *Order) (*Order, error)
		CancelOrder(ctx context.Context, order *Order) error
		Purchase(ctx context.Context, purchase *Purchase, companyId uint64) ([]*warehouse.StockItem, error)
	}

	service struct {
		repository   Repository
		companySvc   company.Service
		warehouseSvc warehouse.Service
		notifier     notification.Notifier
		logger       *log.Logger
	}
)

func NewService(repository Repository, companySvc company.Service, warehouseSvc warehouse.Service, notifier notification.Notifier, logger *log.Logger) Service {
	return &service{repository, companySvc, warehouseSvc, notifier, logger}
}

func (s *service) GetById(ctx context.Context, orderId uint64) (*Order, error) {
	return s.repository.GetById(ctx, orderId)
}

func (s *service) GetByResource(ctx context.Context, resourceId, quality uint64) ([]*Order, error) {
	return s.repository.GetByResource(ctx, resourceId, uint8(quality))
}

func (s *service) Purchase(ctx context.Context, purchase *Purchase, companyId uint64) ([]*warehouse.StockItem, error) {
	stockItem, orders, err := s.repository.Purchase(ctx, purchase, companyId)
	if err != nil {
		return nil, err
	}

	event := notification.Event{
		Type:    notification.OrderPurchased,
		Payload: orders,
	}

	if err := s.notifier.Broadcast(ctx, event); err != nil {
		s.logger.Printf("error broadcasting purchase event: %s", err)
	}

	return stockItem, nil
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

	newOrder, err := s.repository.PlaceOrder(ctx, order, inventory)
	if err != nil {
		return nil, err
	}

	event := notification.Event{
		Type:    notification.OrderPlaced,
		Payload: newOrder,
	}

	if err := s.notifier.Broadcast(ctx, event); err != nil {
		s.logger.Printf("error broadcasting order placed event: %s", err)
	}

	return newOrder, err
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

	if err := s.repository.CancelOrder(ctx, order, inventory); err != nil {
		return err
	}

	event := notification.Event{
		Type:    notification.OrderCanceled,
		Payload: order.Id,
	}

	if err := s.notifier.Broadcast(ctx, event); err != nil {
		s.logger.Printf("error broadcasting order canceled event: %s", err)
	}

	return nil
}
