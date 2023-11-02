package building_test

import (
	"api/building"
	"api/database"
	"api/resource"
	"testing"
)

func TestBuildingRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	tx.Exec(`
        INSERT INTO categories (id, name) VALUES (1, "Construction"), (2, "Food");

        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Metal", 1), (2, "Concrete", 1), (3, "Glass", 1), (4, "Seeds", 2);

        INSERT INTO buildings (id, name) VALUES (1, "Plantation"), (2, "Factory");

        INSERT INTO buildings_requirements (building_id, resource_id, qty, quality)
        VALUES (1, 1, 500, 0), (1, 2, 1000, 0), (1, 3, 100, 1), (2, 1, 1000, 1), (2, 2, 5000, 1);

        INSERT INTO buildings_resources (building_id, resource_id, qty_per_hour)
        VALUES (1, 4, 1000), (2, 1, 250), (2, 3, 100);
    `)

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		_, err := conn.DB.Exec(`
            DELETE FROM buildings_resources;
            DELETE FROM buildings_requirements;
            DELETE FROM resources;
            DELETE FROM buildings;
            DELETE FROM categories;
        `)

		if err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	repository := building.NewRepository(conn, resource.NewRepository(conn))

	t.Run("should list all", func(t *testing.T) {
		t.Parallel()

		buildings, err := repository.GetAll()
		if err != nil {
			t.Fatalf("could not fetch buildings: %s", err)
		}

		if buildings == nil {
			t.Fatal("expected array, got nil")
		}

		if len(buildings) != 2 {
			t.Errorf("expected %d buildings, got %d", 2, len(buildings))
		}

		for _, building := range buildings {
			if building.Id == 1 {
				if len(building.Requirements) != 3 {
					t.Errorf("expected %d requirements, got %d", 3, len(building.Requirements))
				}
				if len(building.Resources) != 1 {
					t.Errorf("expected %d resources, got %d", 1, len(building.Resources))
				}
			}
			if building.Id == 2 {
				if len(building.Requirements) != 2 {
					t.Errorf("expected %d requirements, got %d", 2, len(building.Requirements))
				}
				if len(building.Resources) != 2 {
					t.Errorf("expected %d resources, got %d", 2, len(building.Resources))
				}
			}
		}

	})

	t.Run("should return nil if not found", func(t *testing.T) {
		t.Parallel()

		building, err := repository.GetById(999)
		if err != nil {
			t.Fatalf("could not get building: %s", err)
		}

		if building != nil {
			t.Errorf("expected nil, got %+v", building)
		}
	})

	t.Run("should return with requirements", func(t *testing.T) {
		t.Parallel()

		building, err := repository.GetById(1)
		if err != nil {
			t.Fatalf("could not get building: %s", err)
		}

		if building == nil {
			t.Error("could not get building")
		}

		if len(building.Requirements) != 3 {
			t.Errorf("expected %d requirements, got %d", 3, len(building.Requirements))
		}

		if len(building.Resources) != 1 {
			t.Errorf("expected %d resources, got %d", 1, len(building.Resources))
		}
	})
}
