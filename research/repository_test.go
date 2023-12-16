package research_test

import (
	"api/accounting"
	"api/company"
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

	defer tx.Rollback()

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password)
        VALUES (1, "Foo", "", ""), (2, "Bar", "", "")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO resources (id, name)
        VALUES (1, "Meat"), (2, "Milk")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO researches (id, patents, investment, finishes_at, completed_at, resource_id, company_id) VALUES
        (1, 5, 10000000, '2024-12-12 12:12:12', '2023-12-12 12:12:12', 1, 1),
        (2, 0, 10000000, '2024-12-12 12:12:12', NULL, 2, 1),
        (3, 0, 10000000, '2024-12-12 12:12:12', NULL, 1, 1)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO research_staff (id, name, company_id)
        VALUES (1, "john", 1), (2, "jane", 1), (3, "james", 1), (4, "mark", 1)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO assigned_staff (staff_id, research_id)
        VALUES (1, 1), (2, 1), (3, 2), (4, 2)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO transactions (id, value, company_id)
        VALUES (1, 40000000, 1)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO resources_qualities (resource_id, company_id, quality, patents)
        VALUES (1, 1, 0, 99), (2, 1, 2, 1)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM resources_qualities`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM transactions`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM assigned_staff`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM research_staff`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM researches`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM resources`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	accountingRepo := accounting.NewRepository(conn)
	companyRepo := company.NewRepository(conn, accountingRepo)
	repository := research.NewRepository(conn, accountingRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("IsStaffBusy", func(t *testing.T) {
		t.Run("completed", func(t *testing.T) {
			busy, err := repository.IsStaffBusy(ctx, []uint64{1, 2}, 1)
			if err != nil {
				t.Fatalf("could not check busy: %s", err)
			}
			if busy {
				t.Error("should not be busy once research is completed")
			}
		})

		t.Run("not completed", func(t *testing.T) {
			busy, err := repository.IsStaffBusy(ctx, []uint64{3, 4}, 1)
			if err != nil {
				t.Fatalf("could not check busy: %s", err)
			}
			if !busy {
				t.Error("should be busy since research is not completed")
			}
		})
	})

	t.Run("GetResearch", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			_, err := repository.GetResearch(ctx, 53462)
			if err != research.ErrResearchNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", research.ErrResearchNotFound, err)
			}
		})

		t.Run("brings assigned staff", func(t *testing.T) {
			research, err := repository.GetResearch(ctx, 1)
			if err != nil {
				t.Fatalf("could not get research: %s", err)
			}
			if len(research.AssignedStaff) != 2 {
				t.Errorf("expected %d assigned staff, got %d", 2, len(research.AssignedStaff))
			}
		})
	})

	t.Run("SaveResearch", func(t *testing.T) {
		t.Run("registers staff", func(t *testing.T) {
			research, err := repository.SaveResearch(ctx, time.Now(), 20000000, []uint64{2}, 1, 1)
			if err != nil {
				t.Fatalf("could not save research: %s", err)
			}

			if len(research.AssignedStaff) != 1 {
				t.Errorf("expected %d assigned staff, got %d", 1, len(research.AssignedStaff))
			}
		})

		t.Run("registers transaction", func(t *testing.T) {
			_, err := repository.SaveResearch(ctx, time.Now(), 20000000, []uint64{2}, 1, 1)
			if err != nil {
				t.Fatalf("could not save research: %s", err)
			}

			company, err := companyRepo.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			if company.AvailableCash != 0 {
				t.Errorf("expected cash %d, got %d", 0, company.AvailableCash)
			}
		})
	})

	t.Run("CompleteResearch", func(t *testing.T) {
		t.Run("no level up", func(t *testing.T) {
			now := time.Now()

			research := &research.Research{
				Id:          2,
				Patents:     15,
				Investment:  10000000,
				CompanyId:   1,
				ResourceId:  2,
				CompletedAt: &now,
			}

			research, err := repository.CompleteResearch(ctx, research)
			if err != nil {
				t.Fatalf("could not complete research: %s", err)
			}

			if research.CompletedAt == nil {
				t.Error("should have set completed_at")
			}

			quality, err := repository.GetQuality(ctx, 2, 1)
			if err != nil {
				t.Fatalf("could not get quality: %s", err)
			}

			if quality.Quality != 2 {
				t.Errorf("expected quality %d, got %d", 2, quality.Quality)
			}

			if quality.Patents != 16 {
				t.Errorf("expected patents %d, got %d", 16, quality.Patents)
			}
		})

		t.Run("level up", func(t *testing.T) {
			now := time.Now()

			research := &research.Research{
				Id:          3,
				Patents:     15,
				Investment:  10000000,
				CompanyId:   1,
				ResourceId:  1,
				CompletedAt: &now,
			}

			research, err := repository.CompleteResearch(ctx, research)
			if err != nil {
				t.Fatalf("could not complete research: %s", err)
			}

			if research.CompletedAt == nil {
				t.Error("should have set completed_at")
			}

			quality, err := repository.GetQuality(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not get quality: %s", err)
			}

			if quality.Quality != 1 {
				t.Errorf("expected quality %d, got %d", 1, quality.Quality)
			}

			if quality.Patents != 14 {
				t.Errorf("expected patents %d, got %d", 14, quality.Patents)
			}
		})
	})
}
