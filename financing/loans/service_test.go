package loans_test

import (
	"api/company"
	"api/financing"
	"api/financing/loans"
	"api/scheduler"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestLoansService(t *testing.T) {
	companyRepo := company.NewFakeRepository()
	companySvc := company.NewService(companyRepo)
	financingSvc := financing.NewService(financing.NewFakeRepository())
	service := loans.NewService(loans.NewFakeRepository(companyRepo), companySvc, financingSvc)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("TakeLoan", func(t *testing.T) {
		t.Run("should not allow loan greater than total terrain value", func(t *testing.T) {
			_, err := service.TakeLoan(ctx, 1_500_000_00, 1)
			expectedError := "amount must not be higher than 1300000.00"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("should save loan", func(t *testing.T) {
			loan, err := service.TakeLoan(ctx, 1_000_000_00, 1)
			if err != nil {
				t.Errorf("could not take loan: %s", err)
			}

			if loan.Principal != 1_000_000_00 {
				t.Errorf("expected principal %d, got %d", 1_000_000_00, loan.Principal)
			}
			if loan.CompanyId != 1 {
				t.Errorf("expected company_id %d, got %d", 1, loan.CompanyId)
			}
			if loan.InterestRate != 0.15 {
				t.Errorf("expected interest rate %f, got %f", 0.15, loan.InterestRate)
			}
			if loan.PayableFrom.IsZero() {
				t.Error("should have set payable from")
			}
		})
	})

	t.Run("PayInterest", func(t *testing.T) {
		t.Run("should count delayed payments", func(t *testing.T) {
			ok, err := service.PayLoanInterest(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}
			if !ok {
				t.Error("should not force payment yet")
			}

			ok, err = service.PayLoanInterest(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}
			if ok {
				t.Error("should force payment")
			}
		})

		t.Run("should clear timer", func(t *testing.T) {
			run := make(chan bool)
			timer := scheduler.NewScheduler()

			timer.Repeat(fmt.Sprintf("LOAN_%d", int64(4)), 100*time.Millisecond, func() error {
				run <- true
				return nil
			})

			svc := loans.NewScheduledService(service, timer)

			ok, err := svc.PayLoanInterest(ctx, 4, 1)
			if err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			if !ok {
				t.Error("should be ok")
			}

			select {
			case <-time.After(150 * time.Millisecond):
			case <-run:
				t.Error("should not execute callback")
			}
		})
	})

	t.Run("BuyBackLoan", func(t *testing.T) {
		t.Run("should not buy back more than principal", func(t *testing.T) {
			_, err := service.BuyBackLoan(ctx, 5_000_000_00, 1, 1)
			if err != loans.ErrAmountHigherThanPrincipal {
				t.Errorf("expected error \"%s\", got \"%s\"", loans.ErrAmountHigherThanAvailable, err)
			}
		})

		t.Run("should validate cash", func(t *testing.T) {
			_, err := service.BuyBackLoan(ctx, 1_000_000_00, 3, 1)
			if err != loans.ErrNotEnoughCash {
				t.Errorf("expected error \"%s\", got \"%s\"", loans.ErrNotEnoughCash, err)
			}
		})

		t.Run("should buy back loan", func(t *testing.T) {
			loan, err := service.BuyBackLoan(ctx, 500_000_00, 2, 3)
			if err != nil {
				t.Fatalf("could not buy back loan: %s", err)
			}

			if loan.PrincipalPaid != 1_000_000_00 {
				t.Errorf("expected principal paid %d, got %d", 1_000_000_00, loan.PrincipalPaid)
			}
		})
	})
}
