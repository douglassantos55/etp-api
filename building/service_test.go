package building_test

import (
	"api/auth"
	"api/building"
	"api/database"
	"api/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildingService(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	_, err = tx.Exec(`
        INSERT INTO companies (id, name, email, password)
        VALUES (1, "Test", "email", "password");

        INSERT INTO buildings (id, name) VALUES (1, "Factory");

        INSERT INTO companies_buildings (company_id, building_id, name)
        VALUES (1, 1, "Fabrica");
    `)

	if err != nil {
		t.Fatalf("could not setup database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could commit transaction: %s", err)
	}

	t.Cleanup(func() {
		_, err := conn.DB.Exec(`
            DELETE FROM companies_buildings;
            DELETE FROM buildings;
            DELETE FROM companies;
        `)
		if err != nil {
			t.Fatalf("could not clean up database: %s", err)
		}
	})

	t.Setenv(server.JWT_SECRET_KEY, "secret")

	svr := server.NewServer()
	building.CreateEndpoints(svr, conn)

	token, err := auth.GenerateToken(1, "secret")

	t.Run("should return bad request invalid id", func(t *testing.T) {
        t.Parallel()

		req := httptest.NewRequest("GET", "/buildings/someid", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}
