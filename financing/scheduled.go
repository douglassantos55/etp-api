package financing

import (
	"api/scheduler"
	"context"
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

	s.scheduler.Repeat(loan.Id, Week, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err := s.PayInterest(ctx, loan.Id, loan.CompanyId)
		return err
	})

	return loan, nil
}

func (s *scheduledService) PayInterest(ctx context.Context, loanId, companyId int64) (bool, error) {
	ok, err := s.service.PayInterest(ctx, loanId, companyId)
	if err != nil {
		return false, err
	}
	if !ok {
		s.scheduler.Remove(loanId)
	}
	return true, nil
}
