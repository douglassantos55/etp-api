package building_test

import (
	"api/building"
	"api/database"
	"api/resource"
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.M) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		log.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	tx.Exec(`INSERT INTO categories (id, name) VALUES (1, "Construction"), (2, "Food")`)

	tx.Exec(`
        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Metal", 1), (2, "Concrete", 1), (3, "Glass", 1), (4, "Seeds", 2)
    `)

	tx.Exec(`INSERT INTO buildings (id, name, downtime) VALUES (1, "Plantation", 60), (2, "Factory", 120)`)

	tx.Exec(`
        INSERT INTO buildings_requirements (building_id, resource_id, qty)
        VALUES (1, 1, 500), (1, 2, 1000), (1, 3, 100), (2, 1, 1000), (2, 2, 5000)
    `)

	tx.Exec(`
        INSERT INTO buildings_resources (building_id, resource_id, qty_per_hour)
        VALUES (1, 4, 1000), (2, 1, 250), (2, 3, 100)
    `)

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	exitCode := t.Run()

	tx, err = conn.DB.Begin()
	if err != nil {
		log.Fatalf("could not start transaction: %s", err)
	}

	tx.Exec("DELETE FROM buildings_resources")
	tx.Exec("DELETE FROM buildings_requirements")
	tx.Exec("DELETE FROM resources")
	tx.Exec("DELETE FROM buildings")
	tx.Exec("DELETE FROM categories")

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	os.Exit(exitCode)
}

func TestBuildingRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	t.Cleanup(cancel)

	repository := building.NewRepository(conn, resource.NewRepository(conn))

	t.Run("should list all", func(t *testing.T) {
		buildings, err := repository.GetAll(ctx)
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

				if building.Downtime == nil {
					t.Error("expected downtime")
				} else {
					if *building.Downtime != 60 {
						t.Errorf("expected downtime of %d minutes, got %+v", 60, building.Downtime)
					}
				}
			}
			if building.Id == 2 {
				if len(building.Requirements) != 2 {
					t.Errorf("expected %d requirements, got %d", 2, len(building.Requirements))
				}
				if len(building.Resources) != 2 {
					t.Errorf("expected %d resources, got %d", 2, len(building.Resources))
				}

				if building.Downtime == nil {
					t.Error("expected downtime")
				} else {
					if *building.Downtime != 120 {
						t.Errorf("expected downtime of %d minutes, got %+v", 120, building.Downtime)
					}
				}

				for _, resource := range building.Resources {
					if resource.Name == "" {
						t.Error("should have a name")
					}
				}
			}
		}

	})

	t.Run("should return nil if not found", func(t *testing.T) {
		building, err := repository.GetById(ctx, 999)
		if err != nil {
			t.Fatalf("could not get building: %s", err)
		}

		if building != nil {
			t.Errorf("expected nil, got %+v", building)
		}
	})

	t.Run("should return with requirements", func(t *testing.T) {
		building, err := repository.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get building: %s", err)
		}

		if building == nil {
			t.Fatal("could not get building")
		}

		if len(building.Requirements) != 3 {
			t.Errorf("expected %d requirements, got %d", 3, len(building.Requirements))
		}

		if len(building.Resources) != 1 {
			t.Errorf("expected %d resources, got %d", 1, len(building.Resources))
		}
	})
}
