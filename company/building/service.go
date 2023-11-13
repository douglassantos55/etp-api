package building

import (
	"api/building"
	"api/resource"
	"api/server"
	"api/warehouse"
	"context"
	"time"
)

type (
	BuildingService interface {
		GetBuilding(ctx context.Context, companyId, buildingId uint64) (*CompanyBuilding, error)
		GetBuildings(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error)
		AddBuilding(ctx context.Context, companyId, buildingId uint64, position uint8) (*CompanyBuilding, error)
		Demolish(ctx context.Context, companyId, buildingId uint64) error
	}

	Building struct {
		BuildingId uint64 `json:"building_id" validate:"required"`
		Position   uint8  `json:"position" validate:"required,min=0"`
	}

	CompanyBuilding struct {
		Id              uint64 `db:"id" json:"id"`
		Name            string `db:"name" json:"name"`
		WagesHour       uint64 `db:"wages_per_hour" json:"wages_per_hour"`
		AdminHour       uint64 `db:"admin_per_hour" json:"admin_per_hour"`
		MaintenanceHour uint64 `db:"maintenance_per_hour" json:"maintenance_per_hour"`
		Level           uint8  `db:"level" json:"level"`
		Position        *uint8 `db:"position" json:"position"`
		Resources       []*building.BuildingResource
		BusyUntil       *time.Time `db:"busy_until" json:"busy_until"`
		CompletesAt     *time.Time `db:"completes_at" json:"completes_at"`
	}

	buildingService struct {
		repository   BuildingRepository
		warehouseSvc warehouse.Service
		buildingSvc  building.Service
	}
)

func (b *CompanyBuilding) GetResource(resourceId uint64) (*building.BuildingResource, error) {
	for _, resource := range b.Resources {
		if resource.Id == resourceId {
			return resource, nil
		}
	}
	return nil, server.NewBusinessRuleError("resource not found")
}

func (b *CompanyBuilding) GetProductionRequirements(item *resource.Item) ([]*resource.Item, error) {
	var resourceId uint64
	if item.Resource != nil {
		resourceId = item.Resource.Id
	} else {
		resourceId = item.ResourceId
	}

	resourceToProduce, err := b.GetResource(resourceId)
	if err != nil {
		return nil, err
	}

	// Considers the qty
	for _, requirement := range resourceToProduce.Requirements {
		requirement.Qty *= item.Qty
	}

	return resourceToProduce.Requirements, nil
}

func (b *CompanyBuilding) GetProductionTime(item *resource.Item) (float64, error) {
	var resourceId uint64
	if item.Resource != nil {
		resourceId = item.Resource.Id
	} else {
		resourceId = item.ResourceId
	}

	resourceToProduce, err := b.GetResource(resourceId)
	if err != nil {
		return 0, err
	}

	return float64(item.Qty) / (float64(resourceToProduce.QtyPerHours) / 60.0), nil
}

func (b *CompanyBuilding) GetProductionCost(item *resource.Item) (uint64, error) {
	timeToProduce, err := b.GetProductionTime(item)
	if err != nil {
		return 0, err
	}

	adminCost := uint64(float64(b.AdminHour) / 60.0 * timeToProduce)
	wagesCost := uint64(float64(b.WagesHour) / 60.0 * timeToProduce)

	return uint64(adminCost + wagesCost), nil
}

func NewBuildingService(repository BuildingRepository, warehouseSvc warehouse.Service, buildingSvc building.Service) BuildingService {
	return &buildingService{repository, warehouseSvc, buildingSvc}
}

func (s *buildingService) GetBuilding(ctx context.Context, companyId, buildingId uint64) (*CompanyBuilding, error) {
	return s.repository.GetById(ctx, buildingId, companyId)
}

func (s *buildingService) GetBuildings(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error) {
	return s.repository.GetAll(ctx, companyId)
}

func (s *buildingService) AddBuilding(ctx context.Context, companyId, buildingId uint64, position uint8) (*CompanyBuilding, error) {
	buildingToConstruct, err := s.buildingSvc.GetById(ctx, buildingId)
	if err != nil {
		return nil, err
	}

	if buildingToConstruct == nil {
		return nil, server.NewBusinessRuleError("building not found")
	}

	inventory, err := s.warehouseSvc.GetInventory(ctx, companyId)
	if err != nil {
		return nil, err
	}

	if !inventory.HasResources(buildingToConstruct.Requirements) {
		return nil, server.NewBusinessRuleError("not enough resources")
	}

	inventory.ReduceStock(buildingToConstruct.Requirements)

	return s.repository.AddBuilding(ctx, companyId, inventory, buildingToConstruct, position)
}

func (s *buildingService) Demolish(ctx context.Context, companyId, buildingId uint64) error {
	buildingToDemolish, err := s.GetBuilding(ctx, companyId, buildingId)
	if err != nil {
		return err
	}

	if buildingToDemolish == nil {
		return server.NewBusinessRuleError("building not found")
	}

	if buildingToDemolish.BusyUntil != nil {
		return server.NewBusinessRuleError("cannot demolish busy building")
	}

	return s.repository.Demolish(ctx, companyId, buildingId)
}
