package financing

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

func (s *scheduledService) TakeLoan(ctx context.Context, amount, companyId int64) (*Loan, error) {
	loan, err := s.service.TakeLoan(ctx, amount, companyId)
	if err != nil {
		return nil, err
	}

	s.scheduler.Repeat(fmt.Sprintf("LOAN_%d", loan.Id), Week, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err := s.PayLoanInterest(ctx, loan.Id, loan.CompanyId)
		return err
	})

	return loan, nil
}

func (s *scheduledService) PayLoanInterest(ctx context.Context, loanId, companyId int64) (bool, error) {
	ok, err := s.service.PayLoanInterest(ctx, loanId, companyId)
	if err != nil {
		return false, err
	}
	if !ok {
		s.scheduler.Remove(fmt.Sprintf("LOAN_%d", loanId))
	}
	return true, nil
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
