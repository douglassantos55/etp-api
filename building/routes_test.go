package building_test

import (
	"api/auth"
	"api/building"
	"api/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildingRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	svr := server.NewServer()
	svc := building.NewService(building.NewFakeRepository())
	building.CreateEndpoints(svr, svc)

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate token: %s", err)
	}

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
