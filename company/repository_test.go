package company_test

import (
	"api/company"
	"api/database"
	"testing"
)

func TestRepository(t *testing.T) {
	t.Run("should return with id", func(t *testing.T) {
		conn, err := database.GetConnection(database.SQLITE, "../test.db")
		if err != nil {
			t.Fatalf("could not connect to database: %s", err)
		}

		repository := company.NewRepository(conn)

		company := &company.Company{
			Name:  "McDonalds",
			Pass:  "password",
			Email: "contact@mcdonalds.com",
		}

		if err := repository.SaveCompany(company); err != nil {
			t.Fatalf("could not save company: %s", err)
		}

		if company.Id == 0 {
			t.Errorf("expected an id, got %d", company.Id)
		}
	})
}
