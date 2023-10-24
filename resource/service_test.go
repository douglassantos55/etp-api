package resource_test

import (
	"api/auth"
	"api/database"
	"api/resource"
	"api/server"
	"encoding/json"
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

	_, err = conn.DB.Exec(`
        INSERT INTO categories (id, name) VALUES (1, "Food");
        INSERT INTO resources (id, name, category_id) VALUES (1, "Water", 1), (2, "Seeds", 1);
    `)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec("DELETE FROM resources; DELETE FROM categories;"); err != nil {
			t.Fatalf("could not truncate table: %s", err)
		}
	})

	server := server.NewServer("localhost", "secret")
	resource.CreateEndpoints(server, conn)

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	t.Run("should return 201 when creating resource", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/resources/", strings.NewReader(`{"name":"Wood","category_id":1,"image":"http://placeimg.com/10"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

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
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("should validate input when updating resource", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/resources/1", strings.NewReader(`{"name":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

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
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("should return 400 when updating invalid id", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/resources/someid", strings.NewReader(`{"name":"Iron"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return 200 when updating resource", func(t *testing.T) {
		body := strings.NewReader(`{"name":"Iron"}`)

		req := httptest.NewRequest("PUT", "/resources/1", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("should return 404 if resource is not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/resources/2355", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("should return 400 if id is invalid", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/resources/someid", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return resource", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/resources/2", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resource resource.Resource
		if err := json.Unmarshal(rec.Body.Bytes(), &resource); err != nil {
			t.Fatalf("could not parse json: %s", err)
		}

		if resource.Id != 2 {
			t.Errorf("expected id %d, got %d", 2, resource.Id)
		}
		if resource.Name != "Seeds" {
			t.Errorf("expected name %s, got %s", "Seeds", resource.Name)
		}
	})
}
