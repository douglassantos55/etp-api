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
	requirements map[uint64][]resource.Requirement
	lastId       uint64
}

func NewFakeBuildingRepository() BuildingRepository {
	downtime := uint16(80)
	busyUntil := time.Now().Add(time.Minute)

	requirements := map[uint64][]resource.Requirement{
		1: {
			{Qty: 15, Resource: &resource.Resource{Id: 2}},
		},
		5: {
			{Qty: 15, Resource: &resource.Resource{Id: 2}},
		},
		6: {
			{Qty: 1500, Resource: &resource.Resource{Id: 3}},
		},
	}

	data := map[uint64]map[uint64]*CompanyBuilding{
		1: {
			1: {
				Level: 2,
				Building: &building.Building{
					Id:        1,
					Name:      "Plantation",
					WagesHour: 100,
					AdminHour: 500,
					Downtime:  &downtime,
					Resources: []*building.BuildingResource{
						{
							QtyPerHours: 100,
							Resource:    &resource.Resource{Id: 1, Name: "Seeds"},
						},
					},
					Requirements: []*resource.Item{
						{Qty: 100, Quality: 0, Resource: &resource.Resource{Id: 3}},
					},
				},
			},
			3: {
				Level: 1,
				Building: &building.Building{
					Id:        3,
					Name:      "Laboratory",
					WagesHour: 1000000,
					AdminHour: 5000000,
					Resources: []*building.BuildingResource{
						{
							QtyPerHours: 100,
							Resource:    &resource.Resource{Id: 5, Name: "Vaccine"},
						},
					},
					Requirements: []*resource.Item{
						{Qty: 1, Quality: 10, Resource: &resource.Resource{Id: 1}},
						{Qty: 5, Quality: 9, Resource: &resource.Resource{Id: 2}},
					},
				},
			},
			4: {
				Level:     1,
				BusyUntil: &busyUntil,
				Building: &building.Building{
					Id:        4,
					Name:      "Factory",
					WagesHour: 10,
					AdminHour: 50,
					Resources: []*building.BuildingResource{
						{
							QtyPerHours: 1000,
							Resource:    &resource.Resource{Id: 6, Name: "Iron bar"},
						},
					},
				},
			},
			5: {
				Level:       1,
				CompletesAt: &busyUntil,
				Building: &building.Building{
					Id:        5,
					Name:      "Factory",
					WagesHour: 10,
					AdminHour: 50,
					Resources: []*building.BuildingResource{
						{
							QtyPerHours: 1000,
							Resource:    &resource.Resource{Id: 6, Name: "Iron bar"},
						},
					},
				},
			},
		},
		2: {
			2: {
				Level:       1,
				CompletesAt: &busyUntil,
				Building: &building.Building{
					Id:        2,
					Name:      "Plantation",
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
		},
	}
	return &fakeBuildingRepository{data, requirements, 5}
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
		requirements := make([]*resource.Requirement, 0)
		for _, req := range r.requirements[buildingResource.Resource.Id] {
			requirements = append(requirements, &resource.Requirement{
				Qty:        req.Qty,
				ResourceId: req.ResourceId,
				Resource:   req.Resource,
			})
		}
		buildingResource.Requirements = requirements
	}
}

func (r *fakeBuildingRepository) AddBuilding(ctx context.Context, companyId uint64, inventory *warehouse.Inventory, buildingToConstruct *building.Building, position uint8) (*CompanyBuilding, error) {
	r.lastId++

	companyBuilding := &CompanyBuilding{
		Position: &position,
		Level:    1,
		Building: &building.Building{
			Id:              r.lastId,
			Name:            buildingToConstruct.Name,
			WagesHour:       buildingToConstruct.WagesHour,
			AdminHour:       buildingToConstruct.AdminHour,
			MaintenanceHour: buildingToConstruct.MaintenanceHour,
		},
	}

	r.data[companyId][r.lastId] = companyBuilding
	return companyBuilding, nil
}

func (r *fakeBuildingRepository) Demolish(ctx context.Context, companyId, buildingId uint64) error {
	companyBuildings := r.data[companyId]
	delete(companyBuildings, buildingId)
	return nil
}

func (r *fakeBuildingRepository) Upgrade(ctx context.Context, inventory *warehouse.Inventory, buildingToUpgrade *CompanyBuilding) error {
	return nil
}
