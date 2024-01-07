package accounting_test

import (
	"api/accounting"
	"api/scheduler"
	"context"
	"testing"
	"time"
)

func TestAccountingService(t *testing.T) {
	repository := accounting.NewFakeRepository()
	service := accounting.NewService(repository, scheduler.NewScheduler())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("PayTaxes", func(t *testing.T) {
		start, err := time.Parse(time.DateTime, "2023-12-24 00:00:00")
		if err != nil {
			t.Fatalf("could not parse time: %s", err)
		}

		end, err := time.Parse(time.DateTime, "2023-12-30 23:59:59")
		if err != nil {
			t.Fatalf("could not parse time: %s", err)
		}

		err = service.PayTaxes(ctx, start, end)
		if err != nil {
			t.Fatalf("could not pay taxes: %s", err)
		}

		t.Run("deferred", func(t *testing.T) {
			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 1)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) != 6 {
				t.Errorf("expected %d transactions, got %d", 6, len(transactions))
			}

			found := false
			for _, transaction := range transactions {
				if transaction.Classification == accounting.TAXES_DEFERRED {
					found = true
					if transaction.Value != 240000 {
						t.Errorf("expected deferred taxes %d, got %d", 240000, transaction.Value)
					}
				}
				if transaction.Classification == accounting.TAXES_PAID {
					t.Errorf("expected no taxes, got %d", transaction.Value)
				}
			}

			if !found {
				t.Error("should have deferred taxes")
			}
		})

		t.Run("incurred", func(t *testing.T) {
			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 2)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) != 6 {
				t.Errorf("expected %d transactions, got %d", 6, len(transactions))
			}

			found := false
			for _, transaction := range transactions {
				if transaction.Classification == accounting.TAXES_PAID {
					found = true
					if transaction.Value != -315000 {
						t.Errorf("expected deferred taxes %d, got %d", -315000, transaction.Value)
					}
				}
				if transaction.Classification == accounting.TAXES_DEFERRED {
					t.Errorf("expected no taxes deferred, got %d", transaction.Value)
				}
			}

			if !found {
				t.Error("should have taxes")
			}
		})

		t.Run("incurred - deferred", func(t *testing.T) {
			// Wait for 3 seconds because it should error and setup the timer
			time.Sleep(3500 * time.Millisecond)

			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 3)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) != 7 {
				t.Errorf("expected %d transactions, got %d", 7, len(transactions))
			}

			for _, transaction := range transactions {
				if transaction.Classification == accounting.TAXES_PAID && transaction.Value != -200000 {
					t.Errorf("expected taxes %d, got %d", -200000, transaction.Value)
				}
				if transaction.Classification == accounting.TAXES_DEFERRED && transaction.Value != 115000 {
					t.Errorf("expected deferred taxes %d, got %d", 115000, transaction.Value)
				}
			}
		})
	})

}
