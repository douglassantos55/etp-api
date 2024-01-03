package production

import (
	"api/resource"
	"api/scheduler"
	"api/warehouse"
	"context"
	"time"
)

type ScheduledProductionService struct {
	timer   *scheduler.Scheduler
	service ProductionService
}

func NewScheduledProductionService(service ProductionService, timer *scheduler.Scheduler) ProductionService {
	return &ScheduledProductionService{
		timer:   timer,
		service: service,
	}
}

func (s *ScheduledProductionService) Produce(ctx context.Context, companyId, companyBuildingId uint64, item *resource.Item) (*Production, error) {
	startedProduction, err := s.service.Produce(ctx, companyId, companyBuildingId, item)
	if err != nil {
		return nil, err
	}

	duration := startedProduction.FinishesAt.Sub(time.Now())
	s.timer.Add(startedProduction.Id, duration, func() error {
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

	s.timer.Remove(productionId)
	return nil
}

func (s *ScheduledProductionService) CollectResource(ctx context.Context, companyId, companyBuildingId, productionId uint64) (*warehouse.StockItem, error) {
	return s.service.CollectResource(ctx, companyId, companyBuildingId, productionId)
}
