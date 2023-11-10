package production

import (
	"api/company"
	"api/company/building"
	"api/resource"
	"api/server"
	"api/warehouse"
	"context"
	"time"
)

type (
	ProductionService interface {
		Produce(ctx context.Context, companyId, companyBuildingId uint64, item *resource.Item) (*Production, error)
		CancelProduction(ctx context.Context, companyId, buildingId, productionId uint64) error
		CollectResource(ctx context.Context, companyId, buildingId, productionId uint64) (*warehouse.StockItem, error)
	}

	Production struct {
		*resource.Item
		Id             uint64                    `db:"id" json:"id"`
		Building       *building.CompanyBuilding `db:"-" json:"-"`
		StartedAt      time.Time                 `db:"created_at" json:"started_at"`
		FinishesAt     time.Time                 `db:"finishes_at" json:"finishes_at"`
		CanceledAt     *time.Time                `db:"canceled_at" json:"canceled_at"`
		LastCollection *time.Time                `db:"collected_at" json:"last_collection"`
		SourcingCost   uint64                    `db:"sourcing_cost" json:"sourcing_cost"`
		ProductionCost uint64                    `db:"-" json:"-"`
		ResourcesCost  uint64                    `db:"-" json:"-"`
	}

	productionService struct {
		repository   ProductionRepository
		companySvc   company.Service
		buildingSvc  building.BuildingService
		warehouseSvc warehouse.Service
	}
)

func (p *Production) CalculateSourcingCost() uint64 {
	return p.ProductionCost/p.Qty + p.SourcingCost
}

func (p *Production) ProducedUntil(t time.Time) (*warehouse.StockItem, error) {
	producedResource, err := p.Building.GetResource(p.Resource.Id)
	if err != nil {
		return nil, err
	}

	var lastCollection time.Time
	if p.LastCollection == nil {
		lastCollection = p.StartedAt
	} else {
		lastCollection = *p.LastCollection
	}

	qtyPerMinute := (float64(producedResource.QtyPerHours) / 60.0)
	qtyProduced := t.Sub(lastCollection).Minutes() * qtyPerMinute

	return &warehouse.StockItem{
		Cost: p.SourcingCost,
		Item: &resource.Item{
			Qty:        uint64(qtyProduced),
			Quality:    p.Quality,
			ResourceId: producedResource.Id,
			Resource:   producedResource.Resource,
		},
	}, nil
}

func NewProductionService(repository ProductionRepository, companySvc company.Service, buildingSvc building.BuildingService, warehouseSvc warehouse.Service) ProductionService {
	return &productionService{repository, companySvc, buildingSvc, warehouseSvc}
}

func (s *productionService) Produce(ctx context.Context, companyId, buildingId uint64, item *resource.Item) (*Production, error) {
	buildingToProduce, err := s.buildingSvc.GetBuilding(ctx, companyId, buildingId)
	if err != nil {
		return nil, err
	}

	if buildingToProduce == nil {
		return nil, server.NewBusinessRuleError("building not found")
	}

	if buildingToProduce.BusyUntil != nil {
		return nil, server.NewBusinessRuleError("building is busy")
	}

	inventory, err := s.warehouseSvc.GetInventory(ctx, companyId)
	if err != nil {
		return nil, err
	}

	requirements, err := buildingToProduce.GetProductionRequirements(item)
	if err != nil {
		return nil, err
	}

	if !inventory.HasResources(requirements) {
		return nil, server.NewBusinessRuleError("not enough resources")
	}

	productionCost, err := buildingToProduce.GetProductionCost(item)
	if err != nil {
		return nil, err
	}

	company, err := s.companySvc.GetById(ctx, companyId)
	if err != nil {
		return nil, err
	}

	if company.AvailableCash < int(productionCost) {
		return nil, server.NewBusinessRuleError("not enough cash")
	}

	timeToProduce, err := buildingToProduce.GetProductionTime(item)
	if err != nil {
		return nil, err
	}

	production := &Production{
		Item:           item,
		FinishesAt:     time.Now().Add(time.Second * time.Duration(timeToProduce)),
		Building:       buildingToProduce,
		StartedAt:      time.Now(),
		ProductionCost: productionCost,
		ResourcesCost:  inventory.ReduceStock(requirements),
	}

	return s.repository.SaveProduction(ctx, production, inventory, companyId)
}

func (s *productionService) CancelProduction(ctx context.Context, companyId, buildingId, productionId uint64) error {
	companyBuilding, err := s.buildingSvc.GetBuilding(ctx, companyId, buildingId)
	if err != nil {
		return err
	}

	if companyBuilding == nil {
		return server.NewBusinessRuleError("building not found")
	}

	if companyBuilding.BusyUntil == nil {
		return server.NewBusinessRuleError("no production in process")
	}

	production, err := s.repository.GetProduction(ctx, productionId, buildingId, companyId)
	if err != nil {
		return err
	}

	if production == nil {
		return server.NewBusinessRuleError("production not found")
	}

	now := time.Now()
	production.CanceledAt = &now

	resourceProduced, err := production.ProducedUntil(now)
	if err != nil {
		return err
	}

	inventory, err := s.warehouseSvc.GetInventory(ctx, companyId)
	if err != nil {
		return err
	}

	inventory.IncrementStock([]*warehouse.StockItem{resourceProduced})

	return s.repository.CancelProduction(ctx, production, inventory)
}

func (s *productionService) CollectResource(ctx context.Context, companyId, buildingId, productionId uint64) (*warehouse.StockItem, error) {
	companyBuilding, err := s.buildingSvc.GetBuilding(ctx, companyId, buildingId)
	if err != nil {
		return nil, err
	}

	if companyBuilding == nil {
		return nil, server.NewBusinessRuleError("building not found")
	}

	if companyBuilding.BusyUntil == nil {
		return nil, server.NewBusinessRuleError("no production in process")
	}

	production, err := s.repository.GetProduction(ctx, productionId, buildingId, companyId)
	if err != nil {
		return nil, err
	}

	if production == nil {
		return nil, server.NewBusinessRuleError("production not found")
	}

	now := time.Now()

	resourceProduced, err := production.ProducedUntil(now)
	if err != nil {
		return nil, err
	}

	inventory, err := s.warehouseSvc.GetInventory(ctx, companyId)
	if err != nil {
		return nil, err
	}

	inventory.IncrementStock([]*warehouse.StockItem{resourceProduced})

	production.LastCollection = &now

	if err := s.repository.CollectResource(ctx, production, inventory); err != nil {
		return nil, err
	}

	return resourceProduced, nil
}
