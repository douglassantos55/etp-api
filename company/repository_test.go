package company_test

import (
	"api/company"
	"api/database"
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

	_, err = tx.Exec(`
        INSERT INTO companies (id, name, email, password, created_at, blocked_at, deleted_at) VALUES
        (1, "Coca-Cola", "coke@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, NULL),
        (2, "Blocked", "blocked@email.com", "aoeu", "2023-10-22T01:11:53Z", "2023-10-22T01:11:53Z", NULL),
        (3, "Deleted", "deleted@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, "2023-10-22T01:11:53Z");
    `)

	if err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	_, err = tx.Exec(`INSERT INTO transactions (id, company_id, value) VALUES (1, 1, 100000000)`)
	if err != nil {
		log.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}

	exitCode := t.Run()

	os.Exit(exitCode)
}

func TestRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	t.Cleanup(func() {
		cancel()

		log.Println("STARTING CLEANUP OF COMPANY TEST")

		if _, err := conn.DB.Exec("DELETE FROM companies"); err != nil {
			log.Fatalf("could not cleanup database: %s", err)
		}

		if _, err := conn.DB.Exec("DELETE FROM transactions"); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	repository := company.NewRepository(conn)

	t.Run("should return with cash", func(t *testing.T) {
		company, err := repository.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableCash != 1_000_000_00 {
			t.Errorf("expected cash %d, got %d", 1_000_000_00, company.AvailableCash)
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
		company, err := repository.GetById(ctx, 51245)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find company, got %+v", company)
		}
	})

	t.Run("should ignore deleted", func(t *testing.T) {
		company, err := repository.GetById(ctx, 3)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find company, got %+v", company)
		}
	})

	t.Run("should return nil when not found by email", func(t *testing.T) {
		company, err := repository.GetByEmail(ctx, "test@test.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find company, got %+v", company)
		}
	})

	t.Run("should return company by email", func(t *testing.T) {
		company, err := repository.GetByEmail(ctx, "coke@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company == nil {
			t.Error("should find company")
		}
	})

	t.Run("should not return blocked company by email", func(t *testing.T) {
		company, err := repository.GetByEmail(ctx, "blocked@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find blocked company, got %+v", company)
		}
	})

	t.Run("should not return deleted company by email", func(t *testing.T) {
		company, err := repository.GetByEmail(ctx, "deleted@email.com")
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}
		if company != nil {
			t.Errorf("should not find deleted company, got %+v", company)
		}
	})

	t.Run("PurchaseTerrain", func(t *testing.T) {
		t.Run("should increment available terrains", func(t *testing.T) {
			if err := repository.PurchaseTerrain(ctx, 0, 1); err != nil {
				t.Fatalf("could not purchase terrain: %s", err)
			}

			company, err := repository.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			if company.AvailableTerrains != 4 {
				t.Errorf("expected %d terrains, got %d", 4, company.AvailableTerrains)
			}
		})

		t.Run("should reduce cash", func(t *testing.T) {
			if err := repository.PurchaseTerrain(ctx, 500_000_00, 1); err != nil {
				t.Fatalf("could not purchase terrain: %s", err)
			}

			company, err := repository.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			expectedCash := 500_000_00
			if company.AvailableCash != expectedCash {
				t.Errorf("expected cash %d, got %d", expectedCash, company.AvailableCash)
			}
		})
	})
}
