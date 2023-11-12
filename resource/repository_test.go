package resource_test

import (
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

	tx.Exec(`INSERT INTO categories (id, name) VALUES (1, "Food")`)

	tx.Exec(`
        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Water", 1), (2, "Seeds", 1), (3, "Apple", 1)
    `)

	tx.Exec(`
        INSERT INTO resources_requirements (resource_id, requirement_id, qty, quality)
        VALUES (2, 1, 5, 0), (3, 1, 10, 0), (3, 2, 2, 0)
    `)

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	exitCode := t.Run()

	tx, err = conn.DB.Begin()
	if err != nil {
		log.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	tx.Exec("DELETE FROM resources_requirements")
	tx.Exec("DELETE FROM resources")
	tx.Exec("DELETE FROM categories")

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	os.Exit(exitCode)
}

func TestResourceRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	repository := resource.NewRepository(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	t.Run("should list with requirements", func(t *testing.T) {
		resources, err := repository.FetchResources(ctx)
		if err != nil {
			t.Fatalf("could not fetch resources: %s", err)
		}

		for _, resource := range resources {
			if resource.Id == 1 && len(resource.Requirements) != 0 {
				t.Errorf("expected %d requirements, got %d", 0, len(resource.Requirements))
			}
			if resource.Id == 2 && len(resource.Requirements) != 1 {
				t.Errorf("expected %d requirements, got %d", 1, len(resource.Requirements))
			}
			if resource.Id == 3 && len(resource.Requirements) != 2 {
				t.Errorf("expected %d requirements, got %d", 2, len(resource.Requirements))
			}
		}
	})

	t.Run("should return nil when not found", func(t *testing.T) {
		resource, err := repository.GetById(ctx, 5153)
		if err != nil {
			t.Fatalf("could not get by id: %s", err)
		}
		if resource != nil {
			t.Errorf("should not return an instance, got %+v", resource)
		}
	})

	t.Run("should return instance if found", func(t *testing.T) {
		resource, err := repository.GetById(ctx, 3)
		if err != nil {
			t.Fatalf("could not get by id: %s", err)
		}
		if resource == nil {
			t.Fatal("should return an instance, got nil")
		}
		if resource.Id != 3 {
			t.Errorf("expected id %d, got %d", 3, resource.Id)
		}
		if resource.Name != "Apple" {
			t.Errorf("expected name %s, got %s", "Apple", resource.Name)
		}
		if len(resource.Requirements) != 2 {
			t.Errorf("expected %d requirements, got %d", 2, len(resource.Requirements))
		}
	})

	t.Run("should return resource with ID", func(t *testing.T) {
		resource, err := repository.SaveResource(ctx, &resource.Resource{Name: "water", CategoryId: 1})
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

	t.Run("should save requirements", func(t *testing.T) {
		resource, err := repository.SaveResource(ctx, &resource.Resource{
			Name:       "water",
			CategoryId: 1,
			Requirements: []*resource.Item{
				{Qty: 10, Quality: 0, ResourceId: 1},
				{Qty: 20, Quality: 1, ResourceId: 2},
			},
		})

		if err != nil || resource == nil {
			t.Fatalf("could not save resource: %s", err)
		}

		resource, err = repository.GetById(ctx, resource.Id)
		if err != nil {
			t.Fatalf("could not save resource: %s", err)
		}

		if len(resource.Requirements) != 2 {
			t.Errorf("expected %d requirements, got %d", 2, len(resource.Requirements))
		}
	})
}
