package company_test

import (
	"api/building"
	"api/company"
	"api/database"
	"api/resource"
	"testing"
)

func TestRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

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

        INSERT INTO transactions (company_id, value)
        VALUES (1, 1000000);
    `)
	if err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`
            DELETE FROM buildings_resources;
            DELETE FROM companies_buildings;
            DELETE FROM buildings;
            DELETE FROM resources;
            DELETE FROM categories;
            DELETE FROM companies;
        `); err != nil {
			t.Fatalf("could not cleanup: %s", err)
		}
	})

	resourcesRepository := resource.NewRepository(conn)
	repository := company.NewRepository(conn, resourcesRepository)

	t.Run("should return with cash", func(t *testing.T) {
		company, err := repository.GetById(1)
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

		if len(buildings) != 3 {
			t.Errorf("expected %d buildings, got %d", 3, len(buildings))
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
		t.Parallel()

		buildings, err := repository.GetBuildings(1)
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

	t.Run("should get building with resources", func(t *testing.T) {
		t.Parallel()

		building, err := repository.GetBuilding(2, 1)
		if err != nil {
			t.Fatalf("could not get building: %s", err)
		}

		if building == nil {
			t.Fatal("could not get building")
		}

		if len(building.Resources) != 2 {
			t.Errorf("expected %d resources, got %d", 2, len(building.Resources))
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

	t.Run("should return nil if not found", func(t *testing.T) {
		building, err := repository.GetBuilding(3, 1)
		if err != nil {
			t.Fatalf("could not get building: %s", err)
		}

		if building != nil {
			t.Errorf("should not find building: %+v", building)
		}
	})

	t.Run("should insert building", func(t *testing.T) {
		building, err := repository.AddBuilding(1, &building.Building{Id: 1, Name: "Plantation"}, 1)
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
}
