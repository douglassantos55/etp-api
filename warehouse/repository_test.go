package warehouse_test

import (
	"api/database"
	"api/warehouse"
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.M) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		log.Fatal(err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		log.Fatalf("could not start transaction: %s", err)
	}

	tx.Exec(`INSERT INTO categories (id, name) VALUES (1, "Food"), (2, "Infrastructure")`)

	tx.Exec(`
        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Wood", 2), (2, "Window", 2), (3, "Tools", 2)
    `)

	tx.Exec(`
        INSERT INTO inventories (company_id, resource_id, quantity, quality, sourcing_cost)
        VALUES (1, 1, 1300, 0, 857), (1, 2, 130, 0, 10830), (1, 2, 130, 2, 15830), (1, 1, 150, 1, 905), (1, 3, 150, 5, 1905)
    `)

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	exitCode := t.Run()

	tx, err = conn.DB.Begin()
	if err != nil {
		log.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	tx.Exec("DELETE FROM inventories")
	tx.Exec("DELETE FROM resources")
	tx.Exec("DELETE FROM categories")

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	os.Exit(exitCode)
}

func TestWarehouseRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	t.Cleanup(func() {
		cancel()

	})

	repository := warehouse.NewRepository(conn)

	t.Run("FetchInventory", func(t *testing.T) {
		t.Run("should return empty list", func(t *testing.T) {
			resources, err := repository.FetchInventory(ctx, 2)
			if err != nil {
				t.Fatal(err)
			}
			if resources == nil {
				t.Fatal("expected empty list, got nil")
			}
			if len(resources.Items) != 0 {
				t.Errorf("expected no result, got %d", len(resources.Items))
			}
		})

		t.Run("should list inventory grouped by resource/quality", func(t *testing.T) {
			items, err := repository.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}

			if items == nil {
				t.Fatal("expected result, got nil")
			}

			if len(items.Items) == 0 {
				t.Fatal("expected items, got 0")
			}

			for _, item := range items.Items {
				if item.Resource.Id == 1 {
					if item.Quality == 0 {
						if item.Qty != 1300 {
							t.Errorf("expected qty %d, got %d", 1300, item.Qty)
						}
						if item.Cost != 857 {
							t.Errorf("expected cost %d, got %d", 857, item.Cost)
						}
					} else if item.Quality == 1 {
						if item.Qty != 150 {
							t.Errorf("expected qty %d, got %d", 150, item.Qty)
						}
						if item.Cost != 905 {
							t.Errorf("expected cost %d, got %d", 905, item.Cost)
						}
					}
				}

				if item.Resource.Id == 2 {
					if item.Quality == 0 {
						if item.Qty != 130 {
							t.Errorf("expected qty %d, got %d", 130, item.Qty)
						}
						if item.Cost != 10830 {
							t.Errorf("expected cost %d, got %d", 10830, item.Cost)
						}
					}
					if item.Quality == 2 {
						if item.Qty != 130 {
							t.Errorf("expected qty %d, got %d", 130, item.Qty)
						}
						if item.Cost != 15830 {
							t.Errorf("expected cost %d, got %d", 15830, item.Cost)
						}
					}
				}
			}
		})

		t.Run("should include resource category", func(t *testing.T) {
			items, err := repository.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}

			if items == nil {
				t.Fatal("expected items, got nil")
			}

			for _, item := range items.Items {
				if item.Resource.Category == nil {
					t.Error("expected category, got nil")
				}
				if item.Resource.Category.Name == "" {
					t.Error("expected name, got empty")
				}
			}
		})
	})
}
