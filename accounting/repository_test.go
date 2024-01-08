package accounting_test

import (
	"api/accounting"
	"api/database"
	"context"
	"testing"
	"time"
)

func TestAccountRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password) VALUES
        (1, "Foo", "bar", "bazz"), (2, "Bar", "foo", "bazz"), (3, "Bazz", "bar", "foo")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO classifications (id, name) VALUES
        (1, "Wages"), (3, "Transport fee"), (4, "Refunds"), (5, "Market outflow"),
        (6, "Market inflow"), (7, "Market fee"), (19, "Taxes paid"), (20, "Deferred taxes")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO transactions (company_id, classification_id, value, created_at) VALUES
        (3, 6, 5000000, "2023-12-29 18:55:33"),
        (3, 6, 5000000, "2023-12-28 05:25:33"),
        (3, 7, -700000, "2023-12-27 23:15:53"),
        (3, 7, -1500000, "2023-12-26 11:15:53"),
        (3, 7, -1500000, "2023-12-25 15:15:53"),
        (3, 5, -850000, "2023-12-24 17:35:59"),

        (2, 5, -850000, "2023-12-23 15:15:53"),
        (2, 3, -700000, "2023-12-20 15:15:53"),
        (2, 1, -1500000, "2023-12-18 15:15:53"),

        (1, 1, -1850000, "2023-12-14 05:15:53"),
        (1, 3, -500000, "2023-12-13 13:15:53"),
        (1, 5, -8500000, "2023-12-12 01:15:53"),
        (1, 7, -1500000, "2023-12-12 08:15:53"),
        (1, 21, 700000, "2023-11-11 00:15:53"),
        (1, 21, 700000, "2023-11-11 11:15:53"),
        (1, 21, 1500000, "2023-11-10 00:15:53")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM transactions`); err != nil {
			t.Errorf("could not clean up table: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM classifications`); err != nil {
			t.Errorf("could not clean up table: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Errorf("could not clean up table: %s", err)
		}
	})

	repository := accounting.NewRepository(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("GetPeriodResults", func(t *testing.T) {
		start, err := time.Parse(time.DateTime, "2023-12-24 00:00:00")
		if err != nil {
			t.Fatalf("could not parse time: %s", err)
		}

		end, err := time.Parse(time.DateTime, "2023-12-30 23:59:59")
		if err != nil {
			t.Fatalf("could not parse time: %s", err)
		}

		results, err := repository.GetPeriodResults(ctx, start, end)
		if err != nil {
			t.Fatalf("could not get transactions: %s", err)
		}

		if len(results) != 3 {
			t.Fatalf("expected %d results, got %d", 3, len(results))
		}

		for _, result := range results {
			if result.CompanyId == 1 {
				if result.TaxableIncome != 0 {
					t.Errorf("expected no taxable income, got %d", result.TaxableIncome)
				}
				if result.DeferredTaxes != 2900000 {
					t.Errorf("expected %d deferred taxes, got %d", 2900000, result.DeferredTaxes)
				}
			}
			if result.CompanyId == 2 {
				if result.TaxableIncome != 0 || result.DeferredTaxes != 0 {
					t.Errorf("expected no taxable income and deferred taxes, got %d, %d", result.TaxableIncome, result.DeferredTaxes)
				}
			}
			if result.CompanyId == 3 {
				if result.TaxableIncome != 5450000 {
					t.Errorf("expected taxable income %d, got %d", 5450000, result.TaxableIncome)
				}
				if result.DeferredTaxes != 0 {
					t.Errorf("expected deferred taxes %d, got %d", 0, result.DeferredTaxes)
				}
			}
		}
	})

	t.Run("GetIncomeTransactions", func(t *testing.T) {
		t.Run("should filter by company", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-24 00:00:00")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-30 23:59:59")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 53642)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) != 0 {
				t.Errorf("expected empty transactions, got %d", len(transactions))
			}
		})

		t.Run("should filter by period", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-17 00:00:00")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-23 23:59:59")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 2)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) != 3 {
				t.Errorf("expected %d transactions, got %d", 3, len(transactions))
			}
		})

		t.Run("should group results", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-24 00:00:00")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-30 23:59:59")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 3)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) != 3 {
				t.Errorf("expected %d transactions, got %d", 3, len(transactions))
			}

			for _, transaction := range transactions {
				if transaction.Classification == accounting.MARKET_PURCHASE && transaction.Value != -850000 {
					t.Errorf("expected value %d, got %d", -850000, transaction.Value)
				}
				if transaction.Classification == accounting.MARKET_SALE && transaction.Value != 10000000 {
					t.Errorf("expected value %d, got %d", 10000000, transaction.Value)
				}
				if transaction.Classification == accounting.MARKET_FEE && transaction.Value != -3700000 {
					t.Errorf("expected value %d, got %d", -3700000, transaction.Value)
				}
			}
		})

		t.Run("should include all deferred taxes", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-10 00:00:00")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-16 23:59:59")
			if err != nil {
				t.Fatalf("could not parse time: %s", err)
			}

			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 1)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) != 5 {
				t.Errorf("expected %d transactions, got %d", 5, len(transactions))
			}
		})
	})

	t.Run("SaveTaxes", func(t *testing.T) {
		t.Run("should save deferred", func(t *testing.T) {
			err := repository.SaveTaxes(ctx, -3250000, 3)
			if err != nil {
				t.Fatalf("could not save taxes: %s", err)
			}

			transactions, err := repository.GetIncomeTransactions(ctx, time.Now().UTC(), time.Now().Add(time.Second).UTC(), 3)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			income := accounting.NewIncomeStatement(transactions)
			if income.GetTaxes() != 0 {
				t.Errorf("expected taxes %d, got %d", 0, income.GetTaxes())
			}
			if income.GetDeferredTaxes() != 3250000 {
				t.Errorf("expected deferred taxes %d, got %d", 3250000, income.GetDeferredTaxes())
			}
		})

		t.Run("should remove deferred when saving incurred", func(t *testing.T) {
			err := repository.SaveTaxes(ctx, 3250000, 3)
			if err != nil {
				t.Fatalf("could not save taxes: %s", err)
			}

			start := time.Now().Add(-1 * time.Second).Round(time.Second).UTC()
			end := time.Now().Add(1 * time.Second).Round(time.Second).UTC()

			transactions, err := repository.GetIncomeTransactions(ctx, start, end, 3)
			if err != nil {
				t.Fatalf("could not get transactions: %s", err)
			}

			if len(transactions) == 0 {
				t.Fatal("should retrieve transactions")
			}

			income := accounting.NewIncomeStatement(transactions)
			if income.GetTaxes() != -3250000 {
				t.Errorf("expected taxes %d, got %d", 3250000, income.GetTaxes())
			}
			if income.GetDeferredTaxes() != 0 {
				t.Errorf("expected deferred taxes %d, got %d", 0, income.GetDeferredTaxes())
			}
		})
	})
}
