package warehouse_test

import (
	"api/database"
	"api/warehouse"
	"testing"

	"github.com/doug-martin/goqu/v9"
)

func setup(t testing.TB, conn *database.Connection) {
	t.Helper()

	builder := goqu.New(conn.Driver, conn.DB)

	err := builder.WithTx(func(td *goqu.TxDatabase) error {
		_, err := td.Insert("resources").Rows(
			goqu.Record{"id": 1, "name": "Wood"},
			goqu.Record{"id": 2, "name": "Window"},
		).Executor().Exec()

		if err != nil {
			return err
		}

		_, err = td.Insert("inventories").Rows(
			goqu.Record{"company_id": 1, "resource_id": 1, "quantity": 1300, "quality": 0, "sourcing_cost": 8.57},
			goqu.Record{"company_id": 1, "resource_id": 2, "quantity": 130, "quality": 0, "sourcing_cost": 108.3},
			goqu.Record{"company_id": 1, "resource_id": 1, "quantity": 150, "quality": 1, "sourcing_cost": 9.05},
		).Executor().Exec()

		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := builder.WithTx(func(td *goqu.TxDatabase) error {
			if _, err := td.Delete("inventories").Executor().Exec(); err != nil {
				return err
			}

			_, err := td.Delete("resources").Executor().Exec()
			return err
		})

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestGoquRepository(t *testing.T) {
	db, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatal(err)
	}

	setup(t, db)
	repository := warehouse.NewRepository(db)

	t.Run("should return empty list", func(t *testing.T) {
		resources, err := repository.FetchInventory(2)
		if err != nil {
			t.Fatal(err)
		}
		if resources == nil {
			t.Fatal("expected empty list, got nil")
		}
		if len(resources) != 0 {
			t.Errorf("expected no result, got %d", len(resources))
		}
	})

	t.Run("should list inventory grouped by resource", func(t *testing.T) {
		items, err := repository.FetchInventory(1)
		if err != nil {
			t.Fatal(err)
		}

		if items == nil {
			t.Fatal("expected result, got nil")
		}

		if len(items) == 0 {
			t.Fatal("expected items, got 0")
		}

		for _, item := range items {
			if item.Resource.Id == 1 {
				if item.Qty != 1450 {
					t.Errorf("expected qty %d, got %d", 1450, item.Qty)
				}
				if resource.Cost != 8.78 {
					t.Errorf("expected cost %.2f, got %.2f", 8.78, resource.Cost)
				}
			}

			if item.Resource.Id == 2 {
				if item.Qty != 130 {
					t.Errorf("expected qty %d, got %d", 130, item.Qty)
				}
				if resource.Cost != 108.3 {
					t.Errorf("expected cost %.2f, got %.2f", 108.3, resource.Cost)
				}
			}
		}
	})
}
