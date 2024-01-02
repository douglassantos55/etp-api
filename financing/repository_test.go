package financing_test

import (
	"api/database"
	"api/financing"
	"context"
	"testing"
	"time"
)

func TestFinancingRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not get connection: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password)
        VALUES (1, "Test", "", ""), (2, "Test", "", "")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO classifications (id, name)
        VALUES (6, "Market sale")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO transactions (id, value, classification_id, company_id, created_at) VALUES
        (1, 100000, 6, 1, '2023-11-15T15:53:05Z'),
        (2, 100000, 6, 1, '2023-12-01T02:55:03Z'),
        (3, 120000, 6, 1, '2023-12-20T11:34:52Z'),
        (4, 30000, 6, 1, '2023-12-21T13:34:52Z'),
        (5, 355000, 6, 1, '2023-12-25T13:34:52Z')
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO categories (id, name) VALUES
        (1, "Food"), (2, "Construction")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO resources (id, name, category_id) VALUES
        (1, "Rice", 1), (2, "Iron", 2)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO orders (id, quantity, quality, price, company_id, resource_id)
        VALUES (1, 200, 0, 1000, 1, 1), (2, 100, 1, 1500, 1, 1), (3, 100, 0, 3550, 1, 2)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO orders_transactions (order_id, transaction_id, quantity)
        VALUES (1, 1, 100), (1, 2, 100), (2, 3, 80), (2, 4, 20), (3, 5, 100)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO loans (company_id, principal, interest_rate, payable_from, created_at) VALUES
        (1, 100000, 0.015, '2030-01-01 10:00:00', '2023-10-05 22:50:45'),
        (2, 100000, 0.012, '2030-01-01 10:00:00', '2023-12-05 22:50:45'),
        (1, 100000, 0.018, '2030-01-01 10:00:00', '2023-12-15 22:50:45'),
        (2, 100000, 0.015, '2030-01-01 10:00:00', '2023-12-25 22:50:45'),
        (1, 100000, 0.021, '2030-01-01 10:00:00', '2023-12-31 23:59:58'),
        (2, 100000, 0.019, '2030-01-01 10:00:00', '2024-01-01 00:00:00')
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM loans`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM orders_transactions`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM orders`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM resources`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM categories`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM transactions`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM classifications`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Errorf("could not cleanup database: %s", err)
		}
	})

	repository := financing.NewRepository(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("GetAveragePrices", func(t *testing.T) {
		t.Run("no data", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-01-01 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-01-31 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			avgPrices, err := repository.GetAveragePrices(ctx, start, end)
			if err != nil {
				t.Fatalf("could not get average prices: %s", err)
			}

			if len(avgPrices) != 0 {
				t.Errorf("expected no data, got %d", len(avgPrices))
			}
		})

		t.Run("skip category", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-11-01 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-11-30 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			avgPrices, err := repository.GetAveragePrices(ctx, start, end)
			if err != nil {
				t.Fatalf("could not get average prices: %s", err)
			}

			if len(avgPrices) != 1 {
				t.Errorf("expected one category, got %d", len(avgPrices))
			}

			if avgPrices[1] != 1000 {
				t.Errorf("expected %d, got %d", 1000, avgPrices[1])
			}
		})

		t.Run("group by category", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-01 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-31 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			avgPrices, err := repository.GetAveragePrices(ctx, start, end)
			if err != nil {
				t.Fatalf("could not get average prices: %s", err)
			}

			if len(avgPrices) != 2 {
				t.Errorf("expected %d categories, got %d", 2, len(avgPrices))
			}

			if avgPrices[1] != 1250 {
				t.Errorf("expected %d, got %d", 1250, avgPrices[1])
			}

			if avgPrices[2] != 3550 {
				t.Errorf("expected %d, got %d", 3550, avgPrices[2])
			}
		})
	})

	t.Run("GetAverageInterestRate", func(t *testing.T) {
		t.Run("defaults to 10%", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-11-01 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-11-30 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			rate, err := repository.GetAverageInterestRate(ctx, start, end)
			if err != nil {
				t.Fatalf("could not get interest rate: %s", err)
			}

			if rate != 0.01 {
				t.Errorf("expected rate %f, got %f", 0.01, rate)
			}
		})

		t.Run("calculates", func(t *testing.T) {
			start, err := time.Parse(time.DateTime, "2023-12-01 00:00:00")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			end, err := time.Parse(time.DateTime, "2023-12-31 23:59:59")
			if err != nil {
				t.Fatalf("could not parse date: %s", err)
			}

			rate, err := repository.GetAverageInterestRate(ctx, start, end)
			if err != nil {
				t.Fatalf("could not get interest rate: %s", err)
			}

			if rate != 0.0165 {
				t.Errorf("expected rate %f, got %f", 0.0165, rate)
			}
		})
	})
}
