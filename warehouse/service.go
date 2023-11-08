package warehouse

import (
	"api/resource"
	"context"
)

type (
	Inventory struct {
		Items []*StockItem
	}

	StockItem struct {
		*resource.Item
		Cost uint64 `db:"sourcing_cost" json:"cost"`
	}

	Service interface {
		GetInventory(ctx context.Context, companyId uint64) (*Inventory, error)
	}

	service struct {
		repository Repository
	}
)

func (i *Inventory) GetStock(resourceId uint64, quality uint8) uint64 {
	for _, item := range i.Items {
		if item.Quality == quality && item.Resource.Id == resourceId {
			return item.Qty
		}
	}
	return 0
}

func (i *Inventory) HasResources(resources []*resource.Item) bool {
	if len(resources) == 0 {
		return true
	}
	for _, resource := range resources {
		for _, item := range i.Items {
			isResource := item.Resource.Id == resource.Resource.Id
			isSameQuality := item.Quality == resource.Quality
			hasEnoughQty := item.Qty >= resource.Qty

			if isResource && isSameQuality && hasEnoughQty {
				return true
			}
		}
	}
	return false
}

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetInventory(ctx context.Context, companyId uint64) (*Inventory, error) {
	return s.repository.FetchInventory(ctx, companyId)
}
