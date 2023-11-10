package building

import (
	"api/building"
	"api/resource"
	"api/warehouse"
	"context"
	"time"
)

type fakeBuildingRepository struct {
	data         map[uint64]map[uint64]*CompanyBuilding
	requirements map[uint64][]resource.Item
}

func NewFakeBuildingRepository() BuildingRepository {
	busyUntil := time.Now().Add(time.Minute)

	requirements := map[uint64][]resource.Item{
		1: {
			{Qty: 15, Quality: 0, Resource: &resource.Resource{Id: 2}},
		},
		5: {
			{Qty: 15, Quality: 0, Resource: &resource.Resource{Id: 2}},
		},
		6: {
			{Qty: 1500, Quality: 0, Resource: &resource.Resource{Id: 3}},
		},
	}

	data := map[uint64]map[uint64]*CompanyBuilding{
		1: {
			1: {
				Id:        1,
				Name:      "Plantation",
				Level:     1,
				WagesHour: 100,
				AdminHour: 500,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 100,
						Resource:    &resource.Resource{Id: 1, Name: "Seeds"},
					},
				},
			},
			3: {
				Id:        3,
				Name:      "Laboratory",
				Level:     1,
				WagesHour: 1000000,
				AdminHour: 5000000,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 100,
						Resource:    &resource.Resource{Id: 5, Name: "Vaccine"},
					},
				},
			},
			4: {
				Id:        4,
				Name:      "Factory",
				Level:     1,
				WagesHour: 10,
				AdminHour: 50,
				BusyUntil: &busyUntil,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 1000,
						Resource:    &resource.Resource{Id: 6, Name: "Iron bar"},
					},
				},
			},
		},
		2: {
			2: {
				Id:        2,
				Name:      "Plantation",
				Level:     1,
				WagesHour: 100,
				AdminHour: 500,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 1000,
						Resource:    &resource.Resource{Id: 1, Name: "Seeds"},
					},
				},
			},
		},
	}
	return &fakeBuildingRepository{data, requirements}
}

func (r *fakeBuildingRepository) GetAll(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error) {
	buildings := make([]*CompanyBuilding, 0)
	for _, building := range r.data[companyId] {
		r.getRequirements(building)
		buildings = append(buildings, building)
	}
	return buildings, nil
}

func (r *fakeBuildingRepository) GetById(ctx context.Context, buildingId, companyId uint64) (*CompanyBuilding, error) {
	buildings, ok := r.data[companyId]
	if !ok {
		return nil, nil
	}

	companyBuilding, ok := buildings[buildingId]
	if !ok {
		return nil, nil
	}

	r.getRequirements(companyBuilding)
	return companyBuilding, nil
}

func (r *fakeBuildingRepository) getRequirements(companyBuilding *CompanyBuilding) {
	for _, buildingResource := range companyBuilding.Resources {
		requirements := make([]*resource.Item, 0)
		for _, req := range r.requirements[buildingResource.Resource.Id] {
			requirements = append(requirements, &resource.Item{
				Qty:        req.Qty,
				Quality:    req.Quality,
				ResourceId: req.ResourceId,
				Resource:   req.Resource,
			})
		}
		buildingResource.Requirements = requirements
	}
}

func (r *fakeBuildingRepository) AddBuilding(ctx context.Context, companyId uint64, inventory *warehouse.Inventory, building *building.Building, position uint8) (*CompanyBuilding, error) {
	id := uint64(len(r.data) + 1)

	companyBuilding := &CompanyBuilding{
		Id:              id,
		Name:            building.Name,
		Position:        &position,
		Level:           1,
		WagesHour:       building.WagesHour,
		AdminHour:       building.AdminHour,
		MaintenanceHour: building.MaintenanceHour,
	}

	r.data[companyId][id] = companyBuilding
	return companyBuilding, nil
}
