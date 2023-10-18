package resource_test

import (
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
        INSERT INTO categories (id, name) VALUES (1, "Food");
        INSERT INTO resources (id, name, category_id) VALUES (1, "Water", 1), (2, "Seeds", 1);
    `)
	if err != nil {
		t.Fatal(err)
	}

	repository := resource.NewRepository(conn)

	t.Cleanup(func() {
		if _, err := conn.DB.Exec("DELETE FROM resources; DELETE FROM categories"); err != nil {
			t.Fatalf("could not truncate table: %s", err)
		}
	})

	t.Run("should return nil when not found", func(t *testing.T) {
		resource, err := repository.GetById(5153)
		if err != nil {
			t.Fatalf("could not get by id: %s", err)
		}
		if resource != nil {
			t.Errorf("should not return an instance, got %+v", resource)
		}
	})

	t.Run("should return instance if found", func(t *testing.T) {
		resource, err := repository.GetById(1)
		if err != nil {
			t.Fatalf("could not get by id: %s", err)
		}
		if resource == nil {
			t.Fatal("should return an instance, got nil")
		}
		if resource.Id != 1 {
			t.Errorf("expected id %d, got %d", 1, resource.Id)
		}
		if resource.Name != "Water" {
			t.Errorf("expected name %s, got %s", "Water", resource.Name)
		}
	})

	t.Run("should return resource with ID", func(t *testing.T) {
		resource, err := repository.SaveResource(&resource.Resource{Name: "water", CategoryId: 1})
		if err != nil {
			t.Fatalf("could not save resource: %s", err)
		}

		if resource.Id == 0 {
			t.Error("should add ID after saving")
		}
		if resource.Name != "water" {
			t.Errorf("expected name \"%s\", got \"%s\"", "water", resource.Name)
		}
		if resource.Image != nil {
			t.Errorf("expected no image, got \"%s\"", *resource.Image)
		}
	})
}
