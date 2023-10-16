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
	t.Run("create resource validation", func(t *testing.T) {

		conn, err := database.GetConnection(database.SQLITE, "../test.db")
		if err != nil {
			t.Fatal(err)
		}

		server := server.NewServer()
		resource.CreateEndpoints(server, conn)

		req := httptest.NewRequest("POST", "/resources/", strings.NewReader(`{"name":"","image":""}`))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
		expected := `{"message":"{\"name\":{\"required\":\"name is required to not be empty\"}}"}`
		if rec.Body.String() != expected {
			t.Errorf("expected error message %s, got %s", expected, rec.Body.String())
		}
	})
}
