package company_test

import (
	"api/company"
	"api/database"
	"api/server"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompanyService(t *testing.T) {
	server := server.NewServer()
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	company.CreateEndpoints(server, conn)

	t.Run("should not return password", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"Coca-Cola","email":"coke@coke.com","password":"password"}`)

		req := httptest.NewRequest("POST", "/companies/register", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected code %d, got %d", http.StatusCreated, rec.Code)
		}

		company := new(company.Company)
		if err := json.Unmarshal(rec.Body.Bytes(), company); err != nil {
			t.Fatalf("could not decode: %s", err)
		}

		if company.Pass != "" {
			t.Errorf("should not return password, got %s", company.Pass)
		}
	})

	t.Run("hash password", func(t *testing.T) {
		t.Parallel()

		hashed, err := company.HashPassword("password")
		if err != nil {
			t.Fatalf("could not hash password: %s", err)
		}

		if hashed == "password" {
			t.Errorf("should hash password, got %s", hashed)
		}
	})

	t.Run("compare password", func(t *testing.T) {
		t.Parallel()

		hash := "$2a$10$OBo6gtRDtR2g8X6S9Qn/Z.1r33jf6QYRSxavEIjG8UfrJ8MLQWRzy"
		err := company.ComparePassword(hash, "password")

		if err != nil {
			t.Errorf("error comparing passwords: %s", err)
		}
	})
}
