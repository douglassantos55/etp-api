package warehouse_test

import (
	"api/auth"
	"api/server"
	"api/warehouse"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWarehouseService(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate token: %s", err)
	}

	svr := server.NewServer()
	svc := warehouse.NewService(warehouse.NewFakeRepository())
	warehouse.CreateEndpoints(svr, svc)

	t.Run("should return authenticated company's inventory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/warehouse", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response *warehouse.Inventory
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not parse response: %s", err)
		}

		if len(response.Items) != 3 {
			t.Errorf("expected %d items, got %d", 3, len(response.Items))
		}
	})
}
