package production_test

import (
	"api/company"
	companyBuilding "api/company/building"
	"api/company/building/production"
	"api/database"
	"api/resource"
	"api/warehouse"
	"context"
	"log"
	"math"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.M) {
	conn, err := database.GetConnection(database.SQLITE, "../../../test.db")
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		log.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	productionEnd := time.Now().Add(1 * time.Hour).UTC().Format("2006-01-02 15:04:05")
	productionStart := time.Now().Add(-1 * time.Hour).UTC().Format("2006-01-02 15:04:05")

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password, created_at, blocked_at, deleted_at) VALUES
        (1, "Coca-Cola", "coke@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, NULL),
        (2, "Blocked", "blocked@email.com", "aoeu", "2023-10-22T01:11:53Z", "2023-10-22T01:11:53Z", NULL),
        (3, "Deleted", "deleted@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, "2023-10-22T01:11:53Z");
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`INSERT INTO categories (id, name) VALUES (1, "Construction"), (2, "Food")`); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Metal", 1), (2, "Concrete", 1), (3, "Glass", 1), (4, "Seeds", 2)
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO buildings (id, name, wages_per_hour, admin_per_hour, maintenance_per_hour)
        VALUES (1, "Plantation", 500, 1000, 200), (2, "Factory", 1500, 5000, 500)
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO companies_buildings (id, name, company_id, building_id, level, demolished_at)
        VALUES (1, "Plantation", 1, 1, 2, NULL), (2, "Factory", 1, 2, 3, NULL), (3, "Plantation", 1, 1, 1, "2023-10-25 22:36:21")
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO buildings_resources (building_id, resource_id, qty_per_hour)
        VALUES (1, 4, 1000), (2, 1, 500), (2, 3, 250)
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO buildings_requirements (building_id, resource_id, qty)
        VALUES (1, 1, 50), (2, 1, 150)
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`INSERT INTO transactions (id, company_id, value) VALUES (1, 1, 1000000)`); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO productions (id, resource_id, building_id, qty, quality, finishes_at, created_at, sourcing_cost)
        VALUES (1, 3, 2, 1500, 1, '` + productionEnd + `', '` + productionStart + `', 1352)
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO inventories (company_id, resource_id, quantity, quality, sourcing_cost)
        VALUES (1, 1, 100, 0, 137), (1, 3, 1000, 1, 470), (1, 2, 700, 0, 1553)
    `); err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	exitCode := t.Run()

	os.Exit(exitCode)
}

func TestProductionRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../../../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	t.Cleanup(func() {
		cancel()

		if _, err := conn.DB.Exec("DELETE FROM inventories"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM productions"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM transactions"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM buildings_requirements"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM buildings_resources"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM companies_buildings"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM buildings"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM resources"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM categories"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec("DELETE FROM companies"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}
	})

	companyRepo := company.NewRepository(conn)
	warehouseRepo := warehouse.NewRepository(conn)
	resourceRepo := resource.NewRepository(conn)
	buildingRepo := companyBuilding.NewBuildingRepository(conn, resourceRepo, warehouseRepo)

	repository := production.NewProductionRepository(conn, companyRepo, buildingRepo, warehouseRepo)

	t.Run("Produce", func(t *testing.T) {
		t.Run("should set sourcing cost and register transaction", func(t *testing.T) {
			companyBuilding, err := buildingRepo.GetById(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not get building: %s", err)
			}

			inventory, err := warehouseRepo.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			production := &production.Production{
				Item:           &resource.Item{Qty: 2000, Quality: 0, Resource: &resource.Resource{Id: 4, Name: "Test"}},
				Building:       companyBuilding,
				ProductionCost: 5000 * 100,
				FinishesAt:     time.Now().Add(time.Hour),
				StartedAt:      time.Now(),
			}

			production, err = repository.SaveProduction(ctx, production, inventory, 1)
			if err != nil {
				t.Fatalf("could not produce: %s", err)
			}

			if production == nil {
				t.Fatal("production not found")
			}

			if production.SourcingCost != 250 {
				t.Errorf("expected sourcing cost %d, got %d", 250, production.SourcingCost)
			}

			diff := production.FinishesAt.Sub(time.Now())
			if math.Round(diff.Minutes()) != 60 {
				t.Errorf("expected 60, got %f", math.Round(diff.Minutes()))
			}

			company, err := companyRepo.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			if company == nil {
				t.Fatal("could not find company")
			}

			expectedCash := 5000 * 100
			if company.AvailableCash != expectedCash {
				t.Errorf("expected %d cash, got %d", expectedCash, company.AvailableCash)
			}
		})
	})

	t.Run("CollectResource", func(t *testing.T) {
		t.Run("should set collected at", func(t *testing.T) {
			inventory, err := warehouseRepo.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			production, err := repository.GetProduction(ctx, 1, 2, 1)
			if err != nil {
				t.Fatalf("could not get production: %s", err)
			}

			if production == nil {
				t.Fatal("could not find production")
			}

			now := time.Now()
			production.LastCollection = &now

			err = repository.CollectResource(ctx, production, inventory)
			if err != nil {
				t.Fatalf("could not cancel production: %s", err)
			}

			production, err = repository.GetProduction(ctx, 1, 2, 1)
			if err != nil {
				t.Fatalf("could not get production: %s", err)
			}

			if production == nil {
				t.Fatal("could not find production")
			}

			if production.CanceledAt != nil {
				t.Errorf("should not set canceled at, got %+v", production.CanceledAt)
			}
			if production.LastCollection == nil {
				t.Error("should set collected at")
			}
		})
	})

	t.Run("CancelProduction", func(t *testing.T) {
		t.Run("should set canceled at and collected at", func(t *testing.T) {
			inventory, err := warehouseRepo.FetchInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			production, err := repository.GetProduction(ctx, 1, 2, 1)
			if err != nil {
				t.Fatalf("could not get production: %s", err)
			}

			if production == nil {
				t.Fatal("could not find production")
			}

			now := time.Now()
			production.CanceledAt = &now

			err = repository.CancelProduction(ctx, production, inventory)
			if err != nil {
				t.Fatalf("could not cancel production: %s", err)
			}

			production, err = repository.GetProduction(ctx, 1, 2, 1)
			if err != nil {
				t.Fatalf("could not get production: %s", err)
			}

			if production == nil {
				t.Fatal("could not find production")
			}

			if production.CanceledAt == nil {
				t.Error("should set canceled at")
			}
			if production.LastCollection == nil {
				t.Error("should set collected at")
			}
		})
	})
}
