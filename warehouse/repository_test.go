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
			goqu.Record{"name": "Wood"},
			goqu.Record{"name": "Window"},
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
		resources, err := repository.FetchInventory(1)
		if err != nil {
			t.Fatal(err)
		}

		for _, resource := range resources {
			if resource.Id == 1 {
				if resource.Qty != 1450 {
					t.Errorf("expected qty %d, got %d", 1450, resource.Qty)
				}
				if resource.Cost != 8.78 {
					t.Errorf("expected cost %.2f, got %.2f", 8.78, resource.Cost)
				}
			}

			if resource.Id == 2 {
				if resource.Qty != 130 {
					t.Errorf("expected qty %d, got %d", 130, resource.Qty)
				}
				if resource.Cost != 108.3 {
					t.Errorf("expected cost %.2f, got %.2f", 108.3, resource.Cost)
				}
			}
		}
	})
}
