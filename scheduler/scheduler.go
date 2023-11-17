package scheduler

import (
	"api/company/building"
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
		scheduler   *Scheduler
		buildingSvc building.BuildingService
	}
)

func NewScheduler() *Scheduler {
	return &Scheduler{
		retries: &sync.Map{},
		timers:  &sync.Map{},
	}
}

func NewScheduledBuildingService(buildingSvc building.BuildingService) *ScheduledBuildingService {
	return &ScheduledBuildingService{
		scheduler:   NewScheduler(),
		buildingSvc: buildingSvc,
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
	s.timers.Delete(id)
	s.retries.Delete(id)
}

func (s *ScheduledBuildingService) GetBuilding(ctx context.Context, companyId, buildingId uint64) (*building.CompanyBuilding, error) {
	return s.buildingSvc.GetBuilding(ctx, companyId, buildingId)
}

func (s *ScheduledBuildingService) GetBuildings(ctx context.Context, companyId uint64) ([]*building.CompanyBuilding, error) {
	return s.buildingSvc.GetBuildings(ctx, companyId)
}

func (s *ScheduledBuildingService) Update(ctx context.Context, companyId uint64, companyBuilding *building.CompanyBuilding) error {
	return s.buildingSvc.Update(ctx, companyId, companyBuilding)
}

func (s *ScheduledBuildingService) AddBuilding(ctx context.Context, companyId, buildingId uint64, position uint8) (*building.CompanyBuilding, error) {
	companyBuilding, err := s.buildingSvc.AddBuilding(ctx, companyId, buildingId, position)
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
	err := s.buildingSvc.Demolish(ctx, companyId, buildingId)
	if err != nil {
		return err
	}

	s.scheduler.Remove(buildingId)
	return nil
}

func (s *ScheduledBuildingService) Upgrade(ctx context.Context, companyId, buildingId uint64) (*building.CompanyBuilding, error) {
	companyBuilding, err := s.buildingSvc.Upgrade(ctx, companyId, buildingId)
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
	return s.buildingSvc.Update(ctx, companyId, companyBuilding)
}
