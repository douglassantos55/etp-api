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
		Upgrade(ctx context.Context, companyId, buildingId uint64) (*CompanyBuilding, error)
		Update(ctx context.Context, companyId uint64, companyBuilding *CompanyBuilding) error
	}

	Building struct {
		BuildingId uint64 `json:"building_id" validate:"required"`
		Position   uint8  `json:"position" validate:"required,min=0"`
	}

	CompanyBuilding struct {
		*building.Building

		Level       uint8      `db:"level" json:"level"`
		Position    *uint8     `db:"position" json:"position"`
		BusyUntil   *time.Time `db:"busy_until" json:"busy_until"`
		CompletesAt *time.Time `db:"completes_at" json:"completes_at"`
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

	requirements := make([]*resource.Item, 0)

	// Considers item's qty and quality
	for _, requirement := range resourceToProduce.Requirements {
		requirements = append(requirements, &resource.Item{
			Qty:        requirement.Qty * item.Qty,
			Resource:   requirement.Resource,
			ResourceId: requirement.ResourceId,
			Quality:    uint8(max(0, int(item.Quality)-1)),
		})
	}

	return requirements, nil
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

func (s *buildingService) Update(ctx context.Context, companyId uint64, companyBuilding *CompanyBuilding) error {
	return s.repository.Update(ctx, companyId, companyBuilding)
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

func (s *buildingService) Upgrade(ctx context.Context, companyId, buildingId uint64) (*CompanyBuilding, error) {
	buildingToUpgrade, err := s.GetBuilding(ctx, companyId, buildingId)
	if err != nil {
		return nil, err
	}

	if buildingToUpgrade == nil {
		return nil, server.NewBusinessRuleError("building not found")
	}

	if buildingToUpgrade.CompletesAt != nil {
		return nil, server.NewBusinessRuleError("building is not ready")
	}

	if buildingToUpgrade.BusyUntil != nil {
		return nil, server.NewBusinessRuleError("cannot upgrade busy building")
	}

	inventory, err := s.warehouseSvc.GetInventory(ctx, companyId)
	if err != nil {
		return nil, err
	}

	if !inventory.HasResources(buildingToUpgrade.Requirements) {
		return nil, server.NewBusinessRuleError("not enough resources")
	}

	inventory.ReduceStock(buildingToUpgrade.Requirements)

	completesAt := time.Now().Add(time.Minute * time.Duration(*buildingToUpgrade.Downtime))

	buildingToUpgrade.Level++
	buildingToUpgrade.CompletesAt = &completesAt

	err = s.repository.Upgrade(ctx, inventory, buildingToUpgrade)
	if err != nil {
		return nil, err
	}

	return buildingToUpgrade, nil
}
