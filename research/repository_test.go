package research_test

import (
	"api/database"
	"api/research"
	"context"
	"testing"
	"time"
)

func TestResearchRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password)
        VALUES (1, "Test", "test", "test"), (2, "Other", "other", "other")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO research_staff (id, name, salary, company_id, poacher_id)
        VALUES (1, "Test", 200000, 1, null), (2, "Other", 100000, 2, 1), (3, "T", 500000, 2, null)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM research_staff`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	repository := research.NewRepository(conn)

	t.Run("GetStaff", func(t *testing.T) {
		t.Run("brings poached as well", func(t *testing.T) {
			staff, err := repository.GetStaff(ctx, 1)
			if err != nil {
				t.Fatalf("could not get staff: %s", err)
			}

			if len(staff) != 2 {
				t.Errorf("expected %d staff, got %d", 2, len(staff))
			}
		})
	})

	t.Run("GetStaffById", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			_, err := repository.GetStaffById(ctx, 1543)
			if err != research.ErrStaffNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", research.ErrStaffNotFound, err)
			}
		})

		t.Run("found", func(t *testing.T) {
			staff, err := repository.GetStaffById(ctx, 3)
			if err != nil {
				t.Fatalf("could not find staff: %s", err)
			}

			if staff.Id != 3 {
				t.Errorf("expected id %d, got %d", 3, staff.Id)
			}
		})
	})

	t.Run("RandomStaff", func(t *testing.T) {
		t.Run("should not bring currently poached", func(t *testing.T) {
			staff, err := repository.RandomStaff(ctx, 1)
			if err != nil {
				t.Fatalf("could not get random staff: %s", err)
			}

			if staff.Id != 3 {
				t.Errorf("expected id %d, got %d", 3, staff.Id)
			}
		})
	})
}
