package production

import (
	"api/building"
	companyBuilding "api/company/building"
	"api/resource"
	"api/warehouse"
	"context"
	"time"
)

type fakeProductionRepository struct {
	lastId uint64
	data   map[uint64]map[uint64]*Production
}

func NewFakeProductionRepository() ProductionRepository {
	data := map[uint64]map[uint64]*Production{
		1: {
			1: {
				Id: 1,
				Item: &resource.Item{
					Quality:  0,
					Qty:      2000,
					Resource: &resource.Resource{Id: 6},
				},
				ResourcesCost:  70500,
				FinishesAt:     time.Now().Add(time.Hour),
				StartedAt:      time.Now().Add(-time.Hour),
				ProductionCost: 60,
				SourcingCost:   4704,
				Building: &companyBuilding.CompanyBuilding{
					Id:        4,
					Name:      "Factory",
					Level:     1,
					WagesHour: 10,
					AdminHour: 50,
					Resources: []*building.BuildingResource{
						{
							QtyPerHours: 1000,
							Resource: &resource.Resource{
								Id:   6,
								Name: "Iron bar",
								Requirements: []*resource.Item{
									{Qty: 1500, Quality: 0, Resource: &resource.Resource{Id: 3}},
								},
							},
						},
					},
				},
			},
			2: {
				Id: 2,
				Item: &resource.Item{
					Quality:  0,
					Qty:      2000,
					Resource: &resource.Resource{Id: 6},
				},
				ResourcesCost:  70500,
				FinishesAt:     time.Now().Add(time.Hour),
				StartedAt:      time.Now().Add(-time.Hour),
				ProductionCost: 60,
				SourcingCost:   4704,
				Building: &companyBuilding.CompanyBuilding{
					Id:        4,
					Name:      "Factory",
					Level:     1,
					WagesHour: 10,
					AdminHour: 50,
					Resources: []*building.BuildingResource{
						{
							QtyPerHours: 1000,
							Resource: &resource.Resource{
								Id:   6,
								Name: "Iron bar",
								Requirements: []*resource.Item{
									{Qty: 1500, Quality: 0, Resource: &resource.Resource{Id: 3}},
								},
							},
						},
					},
				},
			},
		},
	}
	return &fakeProductionRepository{2, data}
}

func (r *fakeProductionRepository) GetProduction(ctx context.Context, productionId, buildingId, companyId uint64) (*Production, error) {
	return r.data[companyId][productionId], nil
}

func (r *fakeProductionRepository) SaveProduction(ctx context.Context, production *Production, inventory *warehouse.Inventory, companyId uint64) (*Production, error) {
	finishesAt := time.Now().Add(time.Hour)
	production.Building.BusyUntil = &finishesAt

	r.lastId++
	production.Id = r.lastId

	_, ok := r.data[companyId]
	if !ok {
		r.data[companyId] = make(map[uint64]*Production)
	}

	r.data[companyId][r.lastId] = production

	return production, nil
}

func (r *fakeProductionRepository) CancelProduction(ctx context.Context, production *Production, inventory *warehouse.Inventory) error {
	r.data[inventory.CompanyId][production.Id] = production
	return nil
}

func (r *fakeProductionRepository) CollectResource(ctx context.Context, production *Production, inventory *warehouse.Inventory) error {
	r.data[inventory.CompanyId][production.Id] = production
	return nil
}
