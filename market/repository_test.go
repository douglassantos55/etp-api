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
        (1, "Coca-Cola", "coke@email.com", "aoeu"),
        (2, "McDonalds", "mcdonalds@email.com", "aoeu"),
        (3, "McDonalds", "mcdonalds@email.com", "aoeu"),
        (4, "McDonalds", "mcdonalds@email.com", "aoeu")
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
        INSERT INTO orders (id, company_id, resource_id, quality, quantity, price, sourcing_cost, transport_fee, canceled_at) VALUES
        (1, 1, 1, 0, 100, 1824, 1553, 1137, NULL),

        (2, 1, 2, 1, 1000, 4335, 3768, 5000, NULL),

        (3, 1, 3, 1, 500, 4335, 3768, 5000, NULL),
        (4, 3, 3, 0, 500, 4435, 3868, 5100, NULL),
        (5, 4, 3, 2, 2000, 4535, 3968, 5200, NULL),
        (6, 4, 3, 3, 0, 4545, 3968, 5200, NULL),
        (7, 4, 3, 4, 2000, 4555, 3968, 5200, '2023-11-11')
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO transactions (company_id, value)
        VALUES (1, 100000), (2, 100000000);
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
		if _, err := conn.DB.Exec(`DELETE FROM orders`); err != nil {
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

	t.Run("GetById", func(t *testing.T) {
	})

	t.Run("GetByResource", func(t *testing.T) {
		t.Run("should list greater qualities", func(t *testing.T) {
			orders, err := repository.GetByResource(ctx, 3, 0)
			if err != nil {
				t.Fatalf("could not list orders: %s", err)
			}

			if len(orders) != 3 {
				t.Errorf("expected %d orders, got %d", 3, len(orders))
			}
		})

		t.Run("should not list lower qualities", func(t *testing.T) {
			orders, err := repository.GetByResource(ctx, 3, 1)
			if err != nil {
				t.Fatalf("could not list orders: %s", err)
			}

			if len(orders) != 2 {
				t.Errorf("expected %d orders, got %d", 2, len(orders))
			}
		})

		t.Run("should return empty when nothing found", func(t *testing.T) {
			orders, err := repository.GetByResource(ctx, 7, 1)
			if err != nil {
				t.Fatalf("could not list orders: %s", err)
			}

			if len(orders) != 0 {
				t.Errorf("expected %d orders, got %d", 0, len(orders))
			}
		})

		t.Run("should order by ascending price", func(t *testing.T) {
			orders, err := repository.GetByResource(ctx, 3, 0)
			if err != nil {
				t.Fatalf("could not list orders: %s", err)
			}

			for i, order := range orders {
				if i > 0 && order.Price < orders[i-1].Price {
					t.Errorf("should have higher price: %d, %d", order.Price, orders[i-1].Price)
				}
			}
		})
	})

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

	t.Run("CancelOrder", func(t *testing.T) {
		inventory, err := warehouseRepo.FetchInventory(ctx, 1)
		if err != nil {
			t.Fatalf("could not get inventory: %s", err)
		}

		inventory.IncrementStock([]*warehouse.StockItem{
			{
				Item: &resource.Item{
					Qty:      100,
					Quality:  0,
					Resource: &resource.Resource{Id: 1},
				},
				Cost: 1553,
			},
		})

		order := &market.Order{
			Id:           1,
			Quality:      0,
			ResourceId:   1,
			Quantity:     100,
			Price:        1824,
			TransportFee: 1137,
			SourcingCost: 1553,
			Company:      &company.Company{Id: 1},
		}

		if err := repository.CancelOrder(ctx, order, inventory); err != nil {
			t.Fatalf("could not cancel order: %s", err)
		}

		// Test if transport fee is refunded
		company, err := companyRepo.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		expectedCash := 100000 - 776 + 1137
		if company.AvailableCash != expectedCash {
			t.Errorf("expected cash %d, got %d", expectedCash, company.AvailableCash)
		}

		// Test if stock is restored
		inventory, err = warehouseRepo.FetchInventory(ctx, 1)
		if err != nil {
			t.Fatalf("could not get inventory: %s", err)
		}

		stock := inventory.GetStock(1, 0)
		if stock != 200 {
			t.Errorf("expected stock %d, got %d", 200, stock)
		}
	})

	t.Run("Purchase", func(t *testing.T) {
		t.Run("not enough resources", func(t *testing.T) {
			// Single order
			purchase := &market.Purchase{
				ResourceId: 2,
				Quantity:   100,
				Quality:    3,
			}

			_, err := repository.Purchase(ctx, purchase, 2)
			expectedError := "not enough market orders"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}

			// Multiple orders
			purchase = &market.Purchase{
				ResourceId: 3,
				Quantity:   10000,
				Quality:    0,
			}

			_, err = repository.Purchase(ctx, purchase, 2)
			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("not enough cash", func(t *testing.T) {
			purchase := &market.Purchase{
				ResourceId: 3,
				Quantity:   1000,
				Quality:    0,
			}

			_, err := repository.Purchase(ctx, purchase, 3)
			expectedError := "not enough cash"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("single order", func(t *testing.T) {
			purchase := &market.Purchase{
				ResourceId: 2,
				Quantity:   500,
				Quality:    1,
			}

			items, err := repository.Purchase(ctx, purchase, 2)
			if err != nil {
				t.Fatalf("could not purchase order: %s", err)
			}

			if len(items) != 1 {
				t.Errorf("expected 1 item, got %d", len(items))
			}

			if items[0].Qty != 500 {
				t.Errorf("expected qty %d, got %d", 500, items[0].Qty)
			}

			inventory, err := warehouseRepo.FetchInventory(ctx, 2)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			stock := inventory.GetStock(2, 1)
			if stock != 500 {
				t.Errorf("expected stock %d, got %d", 500, stock)
			}

			order, err := repository.GetById(ctx, 2)
			if err != nil {
				t.Fatalf("could not get order: %s", err)
			}

			if order.Quantity != 500 {
				t.Errorf("expected qty %d, got %d", 500, order.Quantity)
			}

			if order.LastPurchase == nil {
				t.Error("should have set purchased_at")
			}

			company, err := companyRepo.GetById(ctx, 2)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			expectedCash := 100000000 - (500 * 4335)
			if company.AvailableCash != expectedCash {
				t.Errorf("expected cash %d, got %d", expectedCash, company.AvailableCash)
			}

			seller, err := companyRepo.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			expectedCash = 100000 - 776 + 1137 + (500 * 4335)
			if seller.AvailableCash != expectedCash {
				t.Errorf("expected cash %d, got %d", expectedCash, seller.AvailableCash)
			}
		})

		t.Run("multiple orders", func(t *testing.T) {
			purchase := &market.Purchase{
				ResourceId: 3,
				Quantity:   1500,
				Quality:    0,
			}

			items, err := repository.Purchase(ctx, purchase, 2)
			if err != nil {
				t.Fatalf("could not purchase order: %s", err)
			}

			if len(items) != 3 {
				t.Errorf("expected 3 items, got %d", len(items))
			}

			inventory, err := warehouseRepo.FetchInventory(ctx, 2)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			for _, item := range items {
				if item.Qty != 500 {
					t.Errorf("expected qty %d, got %d", 500, item.Qty)
				}

				stock := inventory.GetStock(item.Resource.Id, item.Quality)
				if stock != 500 {
					t.Errorf("expected stock %d, got %d", 500, stock)
				}
			}

			for _, i := range []int{3, 4} {
				order, err := repository.GetById(ctx, uint64(i))
				if err != nil {
					t.Fatalf("could not get order: %s", err)
				}
				if order != nil {
					t.Errorf("should not find zeroed order, got quantity: %d", order.Quantity)
				}
			}

			order, err := repository.GetById(ctx, 5)
			if err != nil {
				t.Fatalf("could not get order: %s", err)
			}

			if order == nil {
				t.Fatal("should find order")
			}

			if order.Quantity != 1500 {
				t.Errorf("expected qty %d, got %d", 1500, order.Quantity)
			}

			if order.LastPurchase == nil {
				t.Error("should have set purchased_at")
			}

			buyer, err := companyRepo.GetById(ctx, 2)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			expectedCash := 100000000 - ((500 * 4335) + (500 * 4335) + (500 * 4435) + (500 * 4535))
			if buyer.AvailableCash != expectedCash {
				t.Errorf("expected cash %d, got %d", expectedCash, buyer.AvailableCash)
			}

			for _, i := range []int{1, 3, 4} {
				seller, err := companyRepo.GetById(ctx, uint64(i))
				if err != nil {
					t.Fatalf("could not get company: %s", err)
				}

				if i == 1 {
					expectedCash = 100000 - 776 + 1137 + ((500 * 4335) * 2)
					if seller.AvailableCash != expectedCash {
						t.Errorf("expected cash %d, got %d", expectedCash, seller.AvailableCash)
					}
				}

				if i == 3 {
					expectedCash = (500 * 4435)
					if seller.AvailableCash != expectedCash {
						t.Errorf("expected cash %d, got %d", expectedCash, seller.AvailableCash)
					}
				}

				if i == 4 {
					expectedCash = (500 * 4535)
					if seller.AvailableCash != expectedCash {
						t.Errorf("expected cash %d, got %d", expectedCash, seller.AvailableCash)
					}
				}
			}
		})
	})
}
