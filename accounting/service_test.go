package accounting_test

import (
	"api/accounting"
	"context"
	"testing"
	"time"
)

func TestAccountingService(t *testing.T) {
	service := accounting.NewService(accounting.NewFakeRepository())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("PayTaxes", func(t *testing.T) {
		t.Run("deferred", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-24 00:00:00")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-30 23:59:59")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			taxes, err := service.PayTaxes(ctx, start, end, 1)
			if err != nil {
				t.Fatalf("could not pay taxes: %s", err)
			}

			expectedTaxes := int64(-2_400_00)
			if taxes != expectedTaxes {
				t.Errorf("expected taxes %d, got %d", expectedTaxes, taxes)
			}
		})

		t.Run("incurred", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-24 00:00:00")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-30 23:59:59")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			taxes, err := service.PayTaxes(ctx, start, end, 2)
			if err != nil {
				t.Fatalf("could not pay taxes: %s", err)
			}

			expectedTaxes := int64(3_150_00)
			if taxes != expectedTaxes {
				t.Errorf("expected taxes %d, got %d", expectedTaxes, taxes)
			}
		})

		t.Run("incurred - deferred", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-24 00:00:00")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-30 23:59:59")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			taxes, err := service.PayTaxes(ctx, start, end, 3)
			if err != nil {
				t.Fatalf("could not pay taxes: %s", err)
			}

			expectedTaxes := int64(2_000_00)
			if taxes != expectedTaxes {
				t.Errorf("expected taxes %d, got %d", expectedTaxes, taxes)
			}
		})
	})

}
