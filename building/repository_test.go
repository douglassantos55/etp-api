package building_test

import (
	"api/building"
	"api/database"
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

	tx.Exec(`INSERT INTO buildings (id, name) VALUES (1, "Plantation"), (2, "Factory")`)

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		_, err := conn.DB.Exec(`DELETE FROM buildings`)

		if err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	repository := building.NewRepository(conn)

	t.Run("should list all", func(t *testing.T) {
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
	})

	t.Run("should return nil if not found", func(t *testing.T) {
		building, err := repository.GetById(999)
		if err != nil {
			t.Fatalf("could not get building: %s", err)
		}

		if building != nil {
			t.Errorf("expected nil, got %+v", building)
		}
	})
}
