package staff_test

import (
	"api/accounting"
	"api/company"
	"api/database"
	"api/research/staff"
	"context"
	"testing"
	"time"
)

func TestResearchRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../../test.db")
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
        INSERT INTO research_staff (id, name, salary, company_id, poacher_id, skill)
        VALUES (1, "Test", 200000, 1, null, 0), (2, "Other", 100000, 2, 1, 90), (3, "T", 500000, 2, null, 0)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO trainings (id, staff_id, company_id, investment, finishes_at)
        VALUES (1, 2, 1, 10000000, '2023-12-31 13:25:22');
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO transactions (company_id, value)
        VALUES (1, 1000000);
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM transactions`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM trainings`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM research_staff`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	accountingRepo := accounting.NewRepository(conn)
	repository := staff.NewRepository(conn, accountingRepo)
	companyRepo := company.NewRepository(conn, accountingRepo)

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
			if err != staff.ErrStaffNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", staff.ErrStaffNotFound, err)
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

	t.Run("SaveTraining", func(t *testing.T) {
		t.Run("should save a transaction", func(t *testing.T) {
			training, err := repository.SaveTraining(ctx, &staff.Training{
				Investment: 500000,
				StaffId:    1,
				CompanyId:  1,
				FinishesAt: time.Now().Add(time.Second),
			})

			if err != nil {
				t.Fatalf("could not save training: %s", err)
			}

			if training.Id == 0 {
				t.Error("should have set an id")
			}

			company, err := companyRepo.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			expectedCash := 500000
			if company.AvailableCash != expectedCash {
				t.Errorf("expected cash %d, got %d", expectedCash, company.AvailableCash)
			}
		})
	})

	t.Run("UpdateTraining", func(t *testing.T) {
		t.Run("update staff skill", func(t *testing.T) {
			err := repository.UpdateTraining(ctx, &staff.Training{
				Id:          1,
				Investment:  10000000,
				StaffId:     2,
				CompanyId:   1,
				Result:      15,
				CompletedAt: time.Now(),
			})

			if err != nil {
				t.Fatalf("could not update training: %s", err)
			}

			training, err := repository.GetTraining(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not get training: %s", err)
			}

			if training.CompletedAt.IsZero() {
				t.Error("should have updated completed_at")
			}

			if training.Result == 0 {
				t.Error("should have updated result")
			}

			staff, err := repository.GetStaffById(ctx, 2)
			if err != nil {
				t.Fatalf("could not get staff: %s", err)
			}

			if staff.Skill != 100 {
				t.Errorf("expected skill %d, got %d", 100, staff.Skill)
			}
		})
	})
}
