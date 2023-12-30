package bonds

import (
	"api/scheduler"
	"context"
	"fmt"
	"time"
)

type scheduledService struct {
	service   Service
	scheduler *scheduler.Scheduler
}

func NewScheduledService(service Service, scheduler *scheduler.Scheduler) Service {
	return &scheduledService{service, scheduler}
}

func (s *scheduledService) GetBonds(ctx context.Context, page, limit uint) ([]*Bond, error) {
	return s.service.GetBonds(ctx, page, limit)
}

func (s *scheduledService) GetCompanyBonds(ctx context.Context, companyId int64) ([]*Bond, error) {
	return s.service.GetCompanyBonds(ctx, companyId)
}

func (s *scheduledService) EmitBond(ctx context.Context, rate float64, amount, companyId int64) (*Bond, error) {
	return s.service.EmitBond(ctx, rate, amount, companyId)
}

func (s *scheduledService) BuyBond(ctx context.Context, amount, bondId, companyId int64) (*Bond, *Creditor, error) {
	bond, creditor, err := s.service.BuyBond(ctx, amount, bondId, companyId)
	if err != nil {
		return nil, nil, err
	}

	s.scheduler.Repeat(fmt.Sprintf("BOND_%d_CREDITOR_%d", bond.Id, creditor.Id), Week, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		return s.PayBondInterest(ctx, creditor, bond)
	})

	return bond, creditor, nil
}

func (s *scheduledService) PayBondInterest(ctx context.Context, creditor *Creditor, bond *Bond) error {
	return s.service.PayBondInterest(ctx, creditor, bond)
}

func (s *scheduledService) BuyBackBond(ctx context.Context, amount, bondId, creditorId, companyId int64) (*Creditor, error) {
	return s.service.BuyBackBond(ctx, amount, bondId, creditorId, companyId)
}
