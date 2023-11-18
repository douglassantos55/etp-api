package scheduler

import (
	"api/company/building"
	"api/company/building/production"
	"api/resource"
	"api/warehouse"
	"context"
	"log"
	"sync"
	"time"
)

type (
	Scheduler struct {
		retries *sync.Map
		timers  *sync.Map
	}

	ScheduledBuildingService struct {
		scheduler *Scheduler
		service   building.BuildingService
	}

	ScheduledProductionService struct {
		scheduler *Scheduler
		service   production.ProductionService
	}
)

func NewScheduler() *Scheduler {
	return &Scheduler{
		retries: &sync.Map{},
		timers:  &sync.Map{},
	}
}

func (s *Scheduler) Add(id uint64, duration time.Duration, callback func() error) {
	s.timers.Store(id, time.AfterFunc(duration, func() {
		s.timers.Delete(id)

		if err := callback(); err != nil {
			log.Println("error, retrying")

			s.retries.Store(id, time.AfterFunc(time.Second, func() {
				s.retries.Delete(id)

				if err := callback(); err != nil {
					log.Printf("could not run callback: %d", id)
				}
			}))
		}
	}))
}

func (s *Scheduler) Remove(id uint64) {
	if retry, found := s.retries.LoadAndDelete(id); found {
		timer := retry.(*time.Timer)
		if !timer.Stop() {
			<-timer.C
		}
	}

	if timer, found := s.timers.LoadAndDelete(id); found {
		if !timer.(*time.Timer).Stop() {
			<-timer.(*time.Timer).C
		}
	}
}

func NewScheduledBuildingService(buildingSvc building.BuildingService) building.BuildingService {
	return &ScheduledBuildingService{
		scheduler: NewScheduler(),
		service:   buildingSvc,
	}
}

func (s *ScheduledBuildingService) GetBuilding(ctx context.Context, companyId, buildingId uint64) (*building.CompanyBuilding, error) {
	return s.service.GetBuilding(ctx, companyId, buildingId)
}

func (s *ScheduledBuildingService) GetBuildings(ctx context.Context, companyId uint64) ([]*building.CompanyBuilding, error) {
	return s.service.GetBuildings(ctx, companyId)
}

func (s *ScheduledBuildingService) Update(ctx context.Context, companyId uint64, companyBuilding *building.CompanyBuilding) error {
	return s.service.Update(ctx, companyId, companyBuilding)
}

func (s *ScheduledBuildingService) AddBuilding(ctx context.Context, companyId, buildingId uint64, position uint8) (*building.CompanyBuilding, error) {
	companyBuilding, err := s.service.AddBuilding(ctx, companyId, buildingId, position)
	if err != nil {
		return nil, err
	}

	duration := companyBuilding.CompletesAt.Sub(time.Now())
	s.scheduler.Add(companyBuilding.Id, duration, func() error {
		return s.completeConstruction(companyId, companyBuilding)
	})

	return companyBuilding, nil
}

func (s *ScheduledBuildingService) Demolish(ctx context.Context, companyId, buildingId uint64) error {
	err := s.service.Demolish(ctx, companyId, buildingId)
	if err != nil {
		return err
	}

	s.scheduler.Remove(buildingId)
	return nil
}

func (s *ScheduledBuildingService) Upgrade(ctx context.Context, companyId, buildingId uint64) (*building.CompanyBuilding, error) {
	companyBuilding, err := s.service.Upgrade(ctx, companyId, buildingId)
	if err != nil {
		return nil, err
	}

	duration := companyBuilding.CompletesAt.Sub(time.Now())
	s.scheduler.Add(buildingId, duration, func() error {
		return s.completeConstruction(companyId, companyBuilding)
	})

	return companyBuilding, nil
}

func (s *ScheduledBuildingService) completeConstruction(companyId uint64, companyBuilding *building.CompanyBuilding) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	companyBuilding.CompletesAt = nil
	return s.service.Update(ctx, companyId, companyBuilding)
}

func NewScheduledProductionService(service production.ProductionService) production.ProductionService {
	return &ScheduledProductionService{
		scheduler: NewScheduler(),
		service:   service,
	}
}

func (s *ScheduledProductionService) Produce(ctx context.Context, companyId, companyBuildingId uint64, item *resource.Item) (*production.Production, error) {
	startedProduction, err := s.service.Produce(ctx, companyId, companyBuildingId, item)
	if err != nil {
		return nil, err
	}

	duration := startedProduction.FinishesAt.Sub(time.Now())
	s.scheduler.Add(startedProduction.Id, duration, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := s.service.CollectResource(ctx, companyId, companyBuildingId, startedProduction.Id)

		return err
	})

	return startedProduction, nil
}

func (s *ScheduledProductionService) CancelProduction(ctx context.Context, companyId, companyBuildingId, productionId uint64) error {
	err := s.service.CancelProduction(ctx, companyId, companyBuildingId, productionId)
	if err != nil {
		return err
	}

	s.scheduler.Remove(productionId)
	return nil
}

func (s *ScheduledProductionService) CollectResource(ctx context.Context, companyId, companyBuildingId, productionId uint64) (*warehouse.StockItem, error) {
	return s.service.CollectResource(ctx, companyId, companyBuildingId, productionId)
}
