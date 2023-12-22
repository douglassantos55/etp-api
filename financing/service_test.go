package financing_test

import (
	"api/company"
	"api/financing"
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
			loan := &financing.Loan{
				Id:              1,
				Principal:       1_000_000_00,
				CompanyId:       1,
				InterestRate:    0.1,
				DelayedPayments: 1,
				PrincipalPaid:   300_000_00,
			}

			if err := service.PayInterest(ctx, loan); err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			if loan.DelayedPayments != 2 {
				t.Errorf("expected delayed payments %d, got %d", 2, loan.DelayedPayments)
			}

			if loan.InterestPaid != 0 {
				t.Errorf("expected interest paid %d, got %d", 0, loan.InterestPaid)
			}
		})

		t.Run("should pay interest", func(t *testing.T) {
			loan := &financing.Loan{
				Id:              2,
				Principal:       1_000_000_00,
				CompanyId:       3,
				InterestRate:    0.1,
				DelayedPayments: 3,
				PrincipalPaid:   500_000_00,
			}

			if err := service.PayInterest(ctx, loan); err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			if loan.DelayedPayments != 0 {
				t.Errorf("should reset delayed payments, got %d", loan.DelayedPayments)
			}

			if loan.InterestPaid != 50_000_00 {
				t.Errorf("expected interest paid %d, got %d", 50_000_00, loan.InterestPaid)
			}
		})

		t.Run("should force principal payment", func(t *testing.T) {
			loan := &financing.Loan{
				Id:              3,
				Principal:       4_000_000_00,
				CompanyId:       1,
				InterestRate:    0.1,
				DelayedPayments: 3,
			}

			if err := service.PayInterest(ctx, loan); err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			company, err := companySvc.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			if company.AvailableTerrains != 0 {
				t.Errorf("expected %d terrains, got %d", 0, company.AvailableTerrains)
			}
		})
	})
}
