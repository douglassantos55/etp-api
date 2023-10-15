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

	t.Cleanup(func() {
		if _, err := conn.DB.Exec("DELETE FROM resources"); err != nil {
			t.Fatalf("could not truncate table: %s", err)
		}
	})

	t.Run("should return resource with ID", func(t *testing.T) {
		repository := resource.NewRepository(conn)
		resource, err := repository.SaveResource(&resource.Resource{Name: "water"})
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
