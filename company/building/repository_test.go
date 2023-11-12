package building_test

import (
	"api/building"
	companyBuilding "api/company/building"
	"api/database"
	"api/resource"
	"api/warehouse"
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.M) {
	conn, err := database.GetConnection(database.SQLITE, "../../test.db")
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		log.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, name FROM companies")
	if err != nil {
		log.Fatalf("could not fetch companies: %s", err)
	}

	defer rows.Close()

	var id uint64
	var name string
	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatalf("could not scan row: %s", err)
		}
		log.Printf("id: %d, name: %s", id, name)
	}
	if rows.Err() != nil {
		log.Fatalf("oops: %s", rows.Err())
	}

	log.Println("STARTING SEED OF COMPANY BUILDING TEST")

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password, created_at, blocked_at, deleted_at) VALUES
        (1, "Coca-Cola", "coke@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, NULL),
        (2, "Blocked", "blocked@email.com", "aoeu", "2023-10-22T01:11:53Z", "2023-10-22T01:11:53Z", NULL),
        (3, "Deleted", "deleted@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, "2023-10-22T01:11:53Z")
    `); err != nil {
		log.Fatalf("could not seed database 1: %s", err)
	}

	if _, err := tx.Exec(`INSERT INTO categories (id, name) VALUES (1, "Construction"), (2, "Food")`); err != nil {
		log.Fatalf("could not seed database 2: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Metal", 1), (2, "Concrete", 1), (3, "Glass", 1), (4, "Seeds", 2)
    `); err != nil {
		log.Fatalf("could not seed database 3: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO buildings (id, name, wages_per_hour, admin_per_hour, maintenance_per_hour)
        VALUES (1, "Plantation", 500, 1000, 200), (2, "Factory", 1500, 5000, 500)
    `); err != nil {
		log.Fatalf("could not seed database 4: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO companies_buildings (id, name, company_id, building_id, level, demolished_at)
        VALUES (1, "Plantation", 1, 1, 2, NULL), (2, "Factory", 1, 2, 3, NULL), (3, "Plantation", 1, 1, 1, "2023-10-25 22:36:21")
    `); err != nil {
		log.Fatalf("could not seed database 5: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO buildings_resources (building_id, resource_id, qty_per_hour)
        VALUES (1, 4, 1000), (2, 1, 500), (2, 3, 250)
    `); err != nil {
		log.Fatalf("could not seed database 6: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO buildings_requirements (building_id, resource_id, qty, quality)
        VALUES (1, 1, 50, 0), (2, 1, 150, 0)
    `); err != nil {
		log.Fatalf("could not seed database 7: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO productions (id, resource_id, building_id, qty, quality, finishes_at, created_at, sourcing_cost)
        VALUES (1, 3, 2, 1500, 1, '2050-11-11 11:11:11', '2023-11-09 11:11:11', 1352);
    `); err != nil {
		log.Fatalf("could not seed database 8: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO inventories (company_id, resource_id, quantity, quality, sourcing_cost)
        VALUES (1, 1, 100, 0, 137), (1, 3, 1000, 1, 470), (1, 2, 700, 0, 1553)
    `); err != nil {
		log.Fatalf("could not seed database 9: %s", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	exitCode := t.Run()

	os.Exit(exitCode)
}

func TestBuildingRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	t.Cleanup(func() {
		cancel()

		log.Println("CLEARING UP SEED OF COMPANY BUILDING TEST")

		if _, err := conn.DB.Exec("DELETE FROM inventories"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM productions"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM buildings_requirements"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM buildings_resources"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM companies_buildings"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM buildings"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM resources"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM categories"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM companies"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	resourceRepo := resource.NewRepository(conn)
	warehouseRepo := warehouse.NewRepository(conn)
	repository := companyBuilding.NewBuildingRepository(conn, resourceRepo, warehouseRepo)

	t.Run("GetAll", func(t *testing.T) {
		t.Run("should return empty list when no buildings are found", func(t *testing.T) {
			buildings, err := repository.GetAll(ctx, 100)
			if err != nil {
				t.Fatalf("could not fetch buildings: %s", err)
			}

			if buildings == nil {
				t.Fatal("expected a list, got nil")
			}

			if len(buildings) != 0 {
				t.Errorf("expected empty list, got %d items", len(buildings))
			}
		})

		t.Run("should ignore demolished buildings", func(t *testing.T) {
			buildings, err := repository.GetAll(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch buildings: %s", err)
			}

			if len(buildings) != 2 {
				t.Errorf("expected %d buildings, got %d", 2, len(buildings))
			}

			for _, building := range buildings {
				if building.Id == 0 {
					t.Error("expected an id, got 0")
				}
				if building.Name == "" {
					t.Error("expected a name")
				}
			}
		})

		t.Run("should list buildings with resources", func(t *testing.T) {
			buildings, err := repository.GetAll(ctx, 1)
			if err != nil {
				t.Fatalf("could not get buildings: %s", err)
			}

			if len(buildings) == 0 {
				t.Fatal("expected buildings")
			}

			for _, building := range buildings {
				if building.Id == 1 {
					if len(building.Resources) != 1 {
						t.Errorf("expected %d resources, got %d", 1, len(building.Resources))
					}

					if building.AdminHour != 2000 {
						t.Errorf("expected admin/h of %d, got %d", 2000, building.AdminHour)
					}
					if building.WagesHour != 1000 {
						t.Errorf("expected wages/h of %d, got %d", 1000, building.WagesHour)
					}

					for _, resource := range building.Resources {
						if resource.Resource.Id == 4 && resource.QtyPerHours != 2000 {
							t.Errorf("expected qty per hour %d, got %d", 2000, resource.QtyPerHours)
						}
					}
				}

				if building.Id == 2 {
					if len(building.Resources) != 2 {
						t.Errorf("expected %d resources, got %d", 2, len(building.Resources))
					}

					if building.AdminHour != 15000 {
						t.Errorf("expected admin/h of %d, got %d", 15000, building.AdminHour)
					}
					if building.WagesHour != 4500 {
						t.Errorf("expected wages/h of %d, got %d", 4500, building.WagesHour)
					}

					for _, resource := range building.Resources {
						if resource.Resource.Id == 1 && resource.QtyPerHours != 1500 {
							t.Errorf("expected qty per hour %d, got %d", 1500, resource.QtyPerHours)
						}
						if resource.Resource.Id == 3 && resource.QtyPerHours != 750 {
							t.Errorf("expected qty per hour %d, got %d", 750, resource.QtyPerHours)
						}
					}
				}
			}
		})

		t.Run("should list buildings with busy until", func(t *testing.T) {
			buildings, err := repository.GetAll(ctx, 1)
			if err != nil {
				t.Fatalf("could not get buildings: %s", err)
			}

			for _, building := range buildings {
				if building.Id == 1 && building.BusyUntil != nil {
					t.Errorf("should not be busy, got %+v", *building.BusyUntil)
				}

				if building.Id == 2 && building.BusyUntil == nil {
					t.Error("should be busy")
				}
			}
		})
	})

	t.Run("GetById", func(t *testing.T) {
		t.Run("should get building with resources", func(t *testing.T) {
			building, err := repository.GetById(ctx, 2, 1)
			if err != nil {
				t.Fatalf("could not get building: %s", err)
			}

			if building == nil {
				t.Fatal("could not get building")
			}

			if len(building.Resources) != 2 {
				t.Errorf("expected %d resources, got %d", 2, len(building.Resources))
			}

			if building.AdminHour != 15000 {
				t.Errorf("expected admin/h of %d, got %d", 15000, building.AdminHour)
			}
			if building.WagesHour != 4500 {
				t.Errorf("expected wages/h of %d, got %d", 4500, building.WagesHour)
			}

			for _, resource := range building.Resources {
				if resource.Resource.Id == 1 && resource.QtyPerHours != 1500 {
					t.Errorf("expected qty per hour %d, got %d", 1500, resource.QtyPerHours)
				}
				if resource.Resource.Id == 3 && resource.QtyPerHours != 750 {
					t.Errorf("expected qty per hour %d, got %d", 750, resource.QtyPerHours)
				}
			}
		})

		t.Run("should get building with busy until", func(t *testing.T) {
			building, err := repository.GetById(ctx, 2, 1)
			if err != nil {
				t.Fatalf("could not get building: %s", err)
			}

			if building == nil {
				t.Fatal("could not get building")
			}

			if building.BusyUntil == nil {
				t.Error("should be busy")
			}
		})

		t.Run("should ignore demolished", func(t *testing.T) {
			building, err := repository.GetById(ctx, 3, 1)
			if err != nil {
				t.Fatalf("could not get building: %s", err)
			}

			if building != nil {
				t.Errorf("should not find building: %+v", building)
			}
		})
	})

	t.Run("AddBuilding", func(t *testing.T) {
		t.Run("should insert building", func(t *testing.T) {
			plantation := &building.Building{
				Id:   1,
				Name: "Plantation",
				Requirements: []*resource.Item{
					{ResourceId: 1, Qty: 50, Quality: 0, Resource: &resource.Resource{Id: 1}},
				},
			}

			inventory, err := warehouseRepo.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			if inventory == nil {
				t.Fatal("could not fetch inventory")
			}

			building, err := repository.AddBuilding(ctx, 1, inventory, plantation, 1)
			if err != nil {
				t.Fatalf("could not insert building: %s", err)
			}

			if building == nil {
				t.Fatal("expected building, got nil")
			}
			if *building.Position != 1 {
				t.Errorf("expected position %d, got %d", 1, building.Position)
			}
			if building.Level != 1 {
				t.Errorf("expected level %d, got %d", 1, building.Level)
			}
			if building.Name != "Plantation" {
				t.Errorf("expected name %s, got %s", "Plantation", building.Name)
			}
		})
	})
}
