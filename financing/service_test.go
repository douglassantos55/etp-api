package financing_test

import (
	"api/financing"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFinancingService(t *testing.T) {
	service := financing.NewService(financing.NewFakeRepository())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("GetInflation", func(t *testing.T) {
		start, err := time.Parse(time.DateTime, "2023-12-01 00:00:00")
		if err != nil {
			t.Fatalf("could not parse date: %s", err)
		}

		end, err := time.Parse(time.DateTime, "2023-12-31 23:59:59")
		if err != nil {
			t.Fatalf("could not parse date: %s", err)
		}

		inflation, categories, err := service.GetInflation(ctx, start, end)
		if err != nil {
			t.Fatalf("could not calculate inflation: %s", err)
		}

		if inflation != 0.125 {
			t.Errorf("expected inflation %f, got %f", 0.125, inflation)
		}

		if categories[1] != 0.25 {
			t.Errorf("expected inflation %.2f, got %.2f", 0.25, categories[1])
		}

		if categories[2] != 0 {
			t.Errorf("expected inflation %.2f, got %.2f", 0.0, categories[2])
		}
	})

	t.Run("GetInterestRate", func(t *testing.T) {
		start, err := time.Parse(time.DateTime, "2023-12-01 00:00:00")
		if err != nil {
			t.Fatalf("could not parse date: %s", err)
		}

		end, err := time.Parse(time.DateTime, "2023-12-31 23:59:59")
		if err != nil {
			t.Fatalf("could not parse date: %s", err)
		}

		rate, err := service.GetInterestRate(ctx, start, end)
		if err != nil {
			t.Fatalf("could not get interest rate: %s", err)
		}

		// (0.0165 + 0.125) / (1 + 0.125) = 0.1257777778
		expectedRate := 0.125778
		if fmt.Sprintf("%f", rate) != fmt.Sprintf("%f", expectedRate) {
			t.Errorf("expected interest rate %f, got %f", expectedRate, rate)
		}
	})
}
