package building

import (
	"api/scheduler"
	"context"
	"time"
)

type ScheduledBuildingService struct {
	scheduler *scheduler.Scheduler
	service   BuildingService
}

func NewScheduledBuildingService(buildingSvc BuildingService) BuildingService {
	return &ScheduledBuildingService{
		scheduler: scheduler.NewScheduler(),
		service:   buildingSvc,
	}
}

func (s *ScheduledBuildingService) GetBuilding(ctx context.Context, companyId, buildingId uint64) (*CompanyBuilding, error) {
	return s.service.GetBuilding(ctx, companyId, buildingId)
}

func (s *ScheduledBuildingService) GetBuildings(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error) {
	return s.service.GetBuildings(ctx, companyId)
}

func (s *ScheduledBuildingService) Update(ctx context.Context, companyId uint64, companyBuilding *CompanyBuilding) error {
	return s.service.Update(ctx, companyId, companyBuilding)
}

func (s *ScheduledBuildingService) AddBuilding(ctx context.Context, companyId, buildingId uint64, position uint8) (*CompanyBuilding, error) {
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

func (s *ScheduledBuildingService) Upgrade(ctx context.Context, companyId, buildingId uint64) (*CompanyBuilding, error) {
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

func (s *ScheduledBuildingService) completeConstruction(companyId uint64, companyBuilding *CompanyBuilding) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	companyBuilding.CompletesAt = nil
	return s.service.Update(ctx, companyId, companyBuilding)
}
