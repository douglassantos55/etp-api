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

		return s.PayInterest(ctx, loan)
	})

	return loan, nil
}
