package company_test

import (
	"api/company"
	"context"
	"testing"
	"time"
)

func TestCompanyService(t *testing.T) {
	service := company.NewService(company.NewFakeRepository())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("PurchaseTerrain", func(t *testing.T) {
		t.Run("should validate cash", func(t *testing.T) {
			err := service.PurchaseTerrain(ctx, 1, 10)
			expectedError := "not enough cash"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("should validate company", func(t *testing.T) {
			err := service.PurchaseTerrain(ctx, 10, 0)
			expectedError := "company not found"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})
	})
}
