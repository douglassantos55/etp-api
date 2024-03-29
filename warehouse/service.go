package warehouse

import (
	"api/resource"
	"context"
)

type (
	Inventory struct {
		CompanyId uint64
		Items     []*StockItem
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

func (i *Inventory) IncrementStock(resources []*StockItem) {
outer:
	for _, resource := range resources {
		for _, item := range i.Items {
			isResource := item.Resource.Id == resource.Resource.Id
			isQuality := item.Quality == resource.Quality

			if isResource && isQuality {
				item.Cost = ((item.Cost * item.Qty) + (resource.Cost * resource.Qty)) / (item.Qty + resource.Qty)
				item.Qty += resource.Qty

				continue outer
			}
		}
		// If not found, append it
		i.Items = append(i.Items, resource)
	}
}

func (i *Inventory) ReduceStock(resources []*resource.Item) uint64 {
	var totalQty uint64
	var sourcingCost uint64

	if len(resources) == 0 {
		return 0
	}

	for _, resource := range resources {
		totalQty += resource.Qty
		remaining := resource.Qty

		for _, item := range i.Items {
			isResource := item.Resource.Id == resource.Resource.Id
			hasSufficientQuality := item.Quality >= resource.Quality

			if remaining > 0 && isResource && hasSufficientQuality {
				if item.Qty > remaining {
					item.Qty -= remaining
					sourcingCost += item.Cost * remaining
				} else {
					remaining -= item.Qty
					sourcingCost += item.Cost * item.Qty
					item.Qty = 0
				}
			}
		}
	}

	return sourcingCost / totalQty
}

func (i *Inventory) HasResources(resources []*resource.Item) bool {
	if len(resources) == 0 {
		return true
	}
	for _, resource := range resources {
		var count uint64
		for _, item := range i.Items {
			isResource := item.Resource.Id == resource.Resource.Id
			hasSufficientQuality := item.Quality >= resource.Quality

			if isResource && hasSufficientQuality {
				count += item.Qty
				if count >= resource.Qty {
					return true
				}
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
