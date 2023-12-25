package financing_test

import (
	"api/company"
	"api/financing"
	"api/scheduler"
	"context"
	"testing"
	"time"
)

func TestFinancingService(t *testing.T) {
	companyRepo := company.NewFakeRepository()
	companySvc := company.NewService(companyRepo)
	service := financing.NewService(financing.NewFakeRepository(companyRepo), companySvc)

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

			timer.Repeat(int64(4), 100*time.Millisecond, func() error {
				run <- true
				return nil
			})

			svc := financing.NewScheduledService(service, timer)

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

	t.Run("EmitBond", func(t *testing.T) {
		t.Run("should not allow loan greater than total terrain value", func(t *testing.T) {
			_, err := service.EmitBond(ctx, 0.15, 1_500_000_00, 1)
			expectedError := "amount must not be higher than 800000.00"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})
	})

	t.Run("PayBondInterest", func(t *testing.T) {
		t.Run("should validate company", func(t *testing.T) {
			err := service.PayBondInterest(ctx, 2, 1)
			if err != financing.ErrBondNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", financing.ErrBondNotFound, err)
			}
		})

		t.Run("should skip if not enough cash", func(t *testing.T) {
			company, err := companyRepo.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			initialCash := company.AvailableCash

			err = service.PayBondInterest(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			if company.AvailableCash != initialCash {
				t.Errorf("expected cash %d, got %d", initialCash, company.AvailableCash)
			}
		})

		t.Run("should pay interest for all creditors", func(t *testing.T) {
			company, err := companyRepo.GetById(ctx, 3)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			initialCash := company.AvailableCash

			err = service.PayBondInterest(ctx, 2, 3)
			if err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			expectedCash := initialCash - 40_000_00 - 150_000_00
			if company.AvailableCash != expectedCash {
				t.Errorf("expected cash %d, got %d", expectedCash, company.AvailableCash)
			}
		})
	})
}
