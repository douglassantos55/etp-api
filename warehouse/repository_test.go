package warehouse_test

import (
	"api/database"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
)

func TestWarehouseRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatal(err)
	}

	builder := goqu.New(conn.Driver, conn.DB)

	err = builder.WithTx(func(td *goqu.TxDatabase) error {
		_, err := td.Insert("categories").Rows(
			goqu.Record{"id": 1, "name": "Food"},
			goqu.Record{"id": 2, "name": "Infrastructure"},
		).Executor().Exec()

		if err != nil {
			return err
		}

		_, err = td.Insert("resources").Rows(
			goqu.Record{"id": 1, "name": "Wood", "category_id": 2},
			goqu.Record{"id": 2, "name": "Window", "category_id": 2},
			goqu.Record{"id": 3, "name": "Tools", "category_id": 2},
		).Executor().Exec()

		if err != nil {
			return err
		}

		_, err = td.Insert("inventories").Rows(
			goqu.Record{"company_id": 1, "resource_id": 1, "quantity": 1300, "quality": 0, "sourcing_cost": 857},
			goqu.Record{"company_id": 1, "resource_id": 2, "quantity": 130, "quality": 0, "sourcing_cost": 10830},
			goqu.Record{"company_id": 1, "resource_id": 2, "quantity": 130, "quality": 2, "sourcing_cost": 15830},
			goqu.Record{"company_id": 1, "resource_id": 1, "quantity": 150, "quality": 1, "sourcing_cost": 905},
			goqu.Record{"company_id": 1, "resource_id": 3, "quantity": 150, "quality": 5, "sourcing_cost": 1905},
		).Executor().Exec()

		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	t.Cleanup(func() {
		cancel()

		err := builder.WithTx(func(td *goqu.TxDatabase) error {
			if _, err := td.Delete("inventories").Executor().Exec(); err != nil {
				return err
			}

			if _, err := td.Delete("resources").Executor().Exec(); err != nil {
				return err
			}

			_, err := td.Delete("categories").Executor().Exec()
			return err
		})

		if err != nil {
			t.Fatal(err)
		}
	})

	repository := warehouse.NewRepository(conn)

	t.Run("FetchInventory", func(t *testing.T) {
		t.Run("should return empty list", func(t *testing.T) {
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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

	t.Run("ReduceStock", func(t *testing.T) {
		inventory, err := repository.FetchInventory(ctx, 1)
		if err != nil {
			t.Fatalf("could not fetch inventory: %s", err)
		}

		t.Run("same quality", func(t *testing.T) {
			tx, err := builder.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("could not begin transaction: %s", err)
			}

			defer tx.Rollback()

			sourcingCost, err := repository.ReduceStock(&database.DB{TxDatabase: tx}, 1, inventory, []*resource.Item{
				{Qty: 1300, Quality: 0, Resource: &resource.Resource{Id: 1}},
			})

			if err != nil {
				t.Fatalf("could not reduce stocks: %s", err)
			}

			if sourcingCost != 857 {
				t.Errorf("expected sourcing cost %d, got %d", 857, sourcingCost/100)
			}

			if err := tx.Commit(); err != nil {
				t.Fatalf("could not commit transaction: %s", err)
			}

			inventory, err = repository.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			stockLeft := inventory.GetStock(1, 0)
			if stockLeft != 0 {
				t.Errorf("expected %d left, got %d", 0, stockLeft)
			}
		})

		t.Run("different qualities", func(t *testing.T) {
			tx, err := builder.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("could not begin transaction: %s", err)
			}

			defer tx.Rollback()

			sourcingCost, err := repository.ReduceStock(&database.DB{TxDatabase: tx}, 1, inventory, []*resource.Item{
				{Qty: 250, Quality: 0, Resource: &resource.Resource{Id: 2}},
			})

			if err != nil {
				t.Fatalf("could not reduce stocks: %s", err)
			}

			if sourcingCost != 13230 {
				t.Errorf("expected sourcing cost %d, got %d", 13230, sourcingCost/100)
			}

			if err := tx.Commit(); err != nil {
				t.Fatalf("could not commit transaction: %s", err)
			}

			inventory, err = repository.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			stockLeft := inventory.GetStock(2, 0)
			if stockLeft != 0 {
				t.Errorf("should have consumed everything, got %d left", stockLeft)
			}

			stockLeft = inventory.GetStock(2, 2)
			if stockLeft != 10 {
				t.Errorf("expected %d left, got %d", 10, stockLeft)
			}
		})

		t.Run("higher quality", func(t *testing.T) {
			tx, err := builder.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("could not begin transaction: %s", err)
			}

			defer tx.Rollback()

			sourcingCost, err := repository.ReduceStock(&database.DB{TxDatabase: tx}, 1, inventory, []*resource.Item{
				{Qty: 100, Quality: 2, Resource: &resource.Resource{Id: 3}},
			})

			if err != nil {
				t.Fatalf("could not reduce stocks: %s", err)
			}

			if sourcingCost != 1905 {
				t.Errorf("expected sourcing cost %d, got %d", 1905, sourcingCost/100)
			}

			if err := tx.Commit(); err != nil {
				t.Fatalf("could not commit transaction: %s", err)
			}

			inventory, err = repository.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			stockLeft := inventory.GetStock(3, 5)
			if stockLeft != 50 {
				t.Errorf("expected %d left, got %d", 50, stockLeft)
			}
		})
	})
}
