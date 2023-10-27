package company_test

import (
	"api/company"
	"api/database"
	"testing"
)

func TestRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	_, err = conn.DB.Exec(`
        INSERT INTO companies (name, email, password, created_at, blocked_at, deleted_at) VALUES
        ("Coca-Cola", "coke@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, NULL),
        ("Blocked", "blocked@email.com", "aoeu", "2023-10-22T01:11:53Z", "2023-10-22T01:11:53Z", NULL),
        ("Deleted", "deleted@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, "2023-10-22T01:11:53Z");

        INSERT INTO buildings (id, name) VALUES (1, "Plantation"), (2, "Factory");

        INSERT INTO companies_buildings (id, name, company_id, building_id, demolished_at)
        VALUES (1, "Plantation", 1, 1, NULL), (2, "Factory", 1, 2, NULL), (3, "Plantation", 1, 1, "2023-10-25 22:36:21");
    `)
	if err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`
            DELETE FROM companies_buildings;
            DELETE FROM buildings;
            DELETE FROM companies;
        `); err != nil {
			t.Fatalf("could not cleanup: %s", err)
		}
	})

	repository := company.NewRepository(conn)

	t.Run("should return with id", func(t *testing.T) {
		t.Parallel()

		registration := &company.Registration{
			Name:     "McDonalds",
			Password: "password",
			Email:    "contact@mcdonalds.com",
			Confirm:  "password",
		}

		company, err := repository.Register(registration)
		if err != nil {
			t.Fatalf("could not save company: %s", err)
		}

		if company.Id == 0 {
			t.Errorf("expected an id, got %d", company.Id)
		}
	})

	t.Run("should return nil when not found by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail("test@test.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find company, got %+v", company)
		}
	})

	t.Run("should return company by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail("coke@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company == nil {
			t.Error("should find company")
		}
	})

	t.Run("should not return blocked company by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail("blocked@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find blocked company, got %+v", company)
		}
	})

	t.Run("should not return deleted company by email", func(t *testing.T) {
		t.Parallel()

		company, err := repository.GetByEmail("deleted@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find deleted company, got %+v", company)
		}
	})

	t.Run("should return empty list when no buildings are found", func(t *testing.T) {
		t.Parallel()

		buildings, err := repository.GetBuildings(100)
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
		t.Parallel()

		buildings, err := repository.GetBuildings(1)
		if err != nil {
			t.Fatalf("could not fetch buildings: %s", err)
		}

		if len(buildings) != 2 {
			t.Errorf("expected %d buildings, got %d", 2, len(buildings))
			t.Log(buildings)
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
}
