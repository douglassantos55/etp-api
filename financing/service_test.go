package financing_test

import (
	"api/company"
	"api/financing"
	"context"
	"testing"
	"time"
)

func TestFinancingService(t *testing.T) {
	companySvc := company.NewService(company.NewFakeRepository())
	service := financing.NewService(financing.NewFakeRepository(), companySvc)

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
}
