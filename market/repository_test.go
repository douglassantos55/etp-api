package market_test

import (
	"api/company"
	"api/database"
	"api/market"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
)

func TestMarketRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password) VALUES
        (1, "Coca-Cola", "coke@email.com", "aoeu")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`INSERT INTO categories (id, name) VALUES (1, "Construction")`); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Metal", 1), (2, "Concrete", 1), (3, "Glass", 1)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO inventories (company_id, resource_id, quantity, quality, sourcing_cost)
        VALUES (1, 1, 100, 0, 137), (1, 3, 1000, 1, 470), (1, 2, 700, 0, 1553)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO transactions (company_id, value)
        VALUES (1, 100000);
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM transactions`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM inventories`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM resources`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM categories`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	companyRepo := company.NewRepository(conn)
	warehouseRepo := warehouse.NewRepository(conn)
	repository := market.NewRepository(conn, companyRepo, warehouseRepo)

	ctx := context.Background()

	t.Run("PlaceOrder", func(t *testing.T) {
		inventory, err := warehouseRepo.FetchInventory(ctx, 1)
		if err != nil {
			t.Fatalf("could not get inventory: %s", err)
		}

		inventory.ReduceStock([]*resource.Item{
			{Qty: 100, Quality: 0, Resource: &resource.Resource{Id: 2}},
		})

		order := &market.Order{
			CompanyId:    1,
			Quality:      0,
			ResourceId:   2,
			Quantity:     100,
			Price:        2275,
			TransportFee: 776,
			SourcingCost: 1553,
		}

		dbOrder, err := repository.PlaceOrder(ctx, order, inventory)
		if err != nil {
			t.Fatalf("could not place order: %s", err)
		}

		if dbOrder.Id == 0 {
			t.Errorf("should have an ID, got %d", dbOrder.Id)
		}

		// Test if inventory is reduced
		inventory, err = warehouseRepo.FetchInventory(ctx, 1)
		if err != nil {
			t.Fatalf("could not get inventory: %s", err)
		}

		stock := inventory.GetStock(2, 0)
		if stock != 600 {
			t.Errorf("expected stock %d, got %d", 600, stock)
		}

		// Test if transport fee is paid
		company, err := companyRepo.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		expectedCash := 100000 - 776
		if company.AvailableCash != expectedCash {
			t.Errorf("expected cash %d, got %d", expectedCash, company.AvailableCash)
		}
	})
}
