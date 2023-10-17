package resource_test

import (
	"api/database"
	"api/resource"
	"api/server"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestService(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.DB.Exec(`INSERT INTO resources (id, name) VALUES (1, "Water"), (2, "Seeds")`)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec("DELETE FROM resources"); err != nil {
			t.Fatalf("could not truncate table: %s", err)
		}
	})

	server := server.NewServer()
	resource.CreateEndpoints(server, conn)

	t.Run("should return 201 when creating resource", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/resources/", strings.NewReader(`{"name":"Wood","image":"http://placeimg.com/10"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
		}
	})

	t.Run("should validate input when creating resource", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/resources/", strings.NewReader(`{"name":"","image":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should validate input when updating resource", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/resources/1", strings.NewReader(`{"name":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return 404 when updating non existing resource", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/resources/1052", strings.NewReader(`{"name":"Iron"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}
