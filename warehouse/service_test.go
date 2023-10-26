package warehouse_test

import (
	"api/auth"
	"api/database"
	"api/server"
	"api/warehouse"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWarehouseService(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	_, err = tx.Exec(`
        INSERT INTO categories (id, name)
        VALUES (1, "Food"), (2, "Construction");

        INSERT INTO resources (id, name, category_id)
        VALUES (1, "Apple", 1), (2, "Iron", 2), (3, "Seed", 1);

        INSERT INTO inventories (company_id, resource_id, quantity, quality, sourcing_cost)
        VALUES (1, 1, 100, 0, 137), (2, 1, 50, 1, 525), (1, 3, 1000, 2, 47), (1, 2, 700, 0, 1553);
    `)
	if err != nil {
		t.Fatalf("could not seed database: %s", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	t.Cleanup(func() {
		tx, err := conn.DB.Begin()
		if err != nil {
			t.Fatalf("could not start transaction: %s", err)
		}

		_, err = tx.Exec(`
            DELETE FROM inventories;
            DELETE FROM resources;
            DELETE FROM companies;
            DELETE FROM categories;
        `)

		if err != nil {
			t.Fatalf("could not clean up database: %s", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("could not clean up database: %s", err)
		}
	})

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate token: %s", err)
	}

	svr := server.NewServer()
	warehouse.CreateEndpoints(svr, conn)

	t.Run("should return authenticated company's inventory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/warehouse", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		response := make([]map[string]any, 0)
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not parse response: %s", err)
		}

		if len(response) != 3 {
			t.Errorf("expected %d items, got %d", 3, len(response))
		}
	})
}
