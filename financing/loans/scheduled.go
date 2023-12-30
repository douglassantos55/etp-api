package loans

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

func (s *scheduledService) GetLoans(ctx context.Context, companyId int64) ([]*Loan, error) {
	return s.service.GetLoans(ctx, companyId)
}

func (s *scheduledService) BuyBackLoan(ctx context.Context, amount, loanId, companyId int64) (*Loan, error) {
	return s.service.BuyBackLoan(ctx, amount, loanId, companyId)
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
