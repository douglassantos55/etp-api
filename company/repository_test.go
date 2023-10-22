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
        ("Deleted", "deleted@email.com", "aoeu", "2023-10-22T01:11:53Z", NULL, "2023-10-22T01:11:53Z")
    `)
	if err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
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
}
