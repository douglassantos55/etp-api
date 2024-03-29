package financing_test

import (
	"api/financing"
	"api/notification"
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestFinancingService(t *testing.T) {
	service := financing.NewService(financing.NewFakeRepository(), notification.NoOpNotifier(), log.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("GetInflation", func(t *testing.T) {
		t.Run("no data", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-01-01 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-01-07 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			inflation, err := service.GetInflationPeriod(ctx, start, end)
			if err != nil {
				t.Fatalf("could not calculate inflation: %s", err)
			}

			if inflation != 0.0 {
				t.Errorf("expected inflation %f, got %f", 0.0, inflation)
			}
		})

		t.Run("with data", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-03 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-09 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			inflation, err := service.GetInflationPeriod(ctx, start, end)
			if err != nil {
				t.Fatalf("could not calculate inflation: %s", err)
			}

			if inflation != 0.125 {
				t.Errorf("expected inflation %f, got %f", 0.125, inflation)
			}
		})
	})

	t.Run("GetInterestRate", func(t *testing.T) {
		t.Run("calculates", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-03 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-09 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			rate, err := service.GetInterestPeriod(ctx, start, end, 0.125)
			if err != nil {
				t.Fatalf("could not get interest rate: %s", err)
			}

			// (0.0165 + 0.125) / (1 + 0.125) = 0.1257777778
			expectedRate := 0.125778
			if fmt.Sprintf("%f", rate) != fmt.Sprintf("%f", expectedRate) {
				t.Errorf("expected interest rate %f, got %f", expectedRate, rate)
			}
		})

		t.Run("no inflation data", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-01-01 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-01-07 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			rate, err := service.GetInterestPeriod(ctx, start, end, 0.0)
			if err != nil {
				t.Fatalf("could not get interest rate: %s", err)
			}

			// (0.0165 + 0.0) / (1 + 0.0) = 0.165
			if rate != 0.0165 {
				t.Errorf("expected rate %f, got %f", 0.0165, rate)
			}
		})
	})
}
