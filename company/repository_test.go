package company_test

import (
	"api/company"
	"api/database"
	"context"
	"testing"
	"time"
)

func TestRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	productionEnd := time.Now().Add(1 * time.Hour).UTC().Format("2006-01-02 15:04:05")
	productionStart := time.Now().Add(-1 * time.Hour).UTC().Format("2006-01-02 15:04:05")

	_, err = conn.DB.Exec(`
        INSERT INTO companies (id, name, email, password, created_at, blocked_at, deleted_at) VALUES
        (1, "Coca-Cola", "coke@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, NULL),
        (2, "Blocked", "blocked@email.com", "aoeu", "2023-10-22T01:11:53Z", "2023-10-22T01:11:53Z", NULL),
        (3, "Deleted", "deleted@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, "2023-10-22T01:11:53Z");

        INSERT INTO categories (id, name) VALUES (1, "Construction"), (2, "Food");

        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Metal", 1), (2, "Concrete", 1), (3, "Glass", 1), (4, "Seeds", 2);

        INSERT INTO buildings (id, name, wages_per_hour, admin_per_hour, maintenance_per_hour)
        VALUES (1, "Plantation", 500, 1000, 200), (2, "Factory", 1500, 5000, 500);

        INSERT INTO companies_buildings (id, name, company_id, building_id, level, demolished_at)
        VALUES (1, "Plantation", 1, 1, 2, NULL), (2, "Factory", 1, 2, 3, NULL), (3, "Plantation", 1, 1, 1, "2023-10-25 22:36:21");

        INSERT INTO buildings_resources (building_id, resource_id, qty_per_hour)
        VALUES (1, 4, 1000), (2, 1, 500), (2, 3, 250);

        INSERT INTO buildings_requirements (building_id, resource_id, qty, quality)
        VALUES (1, 1, 50, 0), (2, 1, 150, 0);

        INSERT INTO transactions (company_id, value)
        VALUES (1, 1000000);

        INSERT INTO productions (id, resource_id, building_id, qty, quality, finishes_at, created_at, sourcing_cost)
        VALUES (1, 3, 2, 1500, 1, '` + productionEnd + `', '` + productionStart + `', 1352);

        INSERT INTO inventories (company_id, resource_id, quantity, quality, sourcing_cost)
        VALUES (1, 1, 100, 0, 137), (1, 3, 1000, 1, 470), (1, 2, 700, 0, 1553);
    `)
	if err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	repository := company.NewRepository(conn)

	t.Run("should return with cash", func(t *testing.T) {
		company, err := repository.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableCash != 1_000_000 {
			t.Errorf("expected cash %d, got %d", 1_000_000, company.AvailableCash)
		}
	})

	t.Run("should return with id", func(t *testing.T) {
		registration := &company.Registration{
			Name:     "McDonalds",
			Password: "password",
			Email:    "contact@mcdonalds.com",
			Confirm:  "password",
		}

		company, err := repository.Register(ctx, registration)
		if err != nil {
			t.Fatalf("could not save company: %s", err)
		}

		if company.Id == 0 {
			t.Errorf("expected an id, got %d", company.Id)
		}
	})

	t.Run("should return nil when not found by id", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetById(ctx, 51245)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find company, got %+v", company)
		}
	})

	t.Run("should ignore deleted", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetById(ctx, 3)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find company, got %+v", company)
		}
	})

	t.Run("should return nil when not found by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail(ctx, "test@test.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find company, got %+v", company)
		}
	})

	t.Run("should return company by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail(ctx, "coke@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company == nil {
			t.Error("should find company")
		}
	})

	t.Run("should not return blocked company by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail(ctx, "blocked@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find blocked company, got %+v", company)
		}
	})

	t.Run("should not return deleted company by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail(ctx, "deleted@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find deleted company, got %+v", company)
		}
	})

	cancel()

	if _, err := conn.DB.Exec(`
            DELETE FROM inventories;
            DELETE FROM productions;
            DELETE FROM transactions;
            DELETE FROM buildings_requirements;
            DELETE FROM buildings_resources;
            DELETE FROM companies_buildings;
            DELETE FROM buildings;
            DELETE FROM resources;
            DELETE FROM categories;
            DELETE FROM companies;
        `); err != nil {
		t.Fatalf("could not cleanup: %s", err)
	}
}
