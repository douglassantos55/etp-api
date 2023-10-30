package building_test

import (
	"api/auth"
	"api/building"
	"api/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeRepository struct {
	data map[uint64]*building.Building
}

func NewFakeRepository() building.Repository {
	data := map[uint64]*building.Building{
		1: {Id: 1, Name: "Plantation"},
		2: {Id: 2, Name: "Factory"},
	}
	return &fakeRepository{data}
}

func (r *fakeRepository) GetAll() ([]*building.Building, error) {
	items := make([]*building.Building, 0)
	for _, item := range r.data {
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeRepository) GetById(id uint64) (*building.Building, error) {
	return r.data[id], nil
}

func TestBuildingService(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	svr := server.NewServer()
	svc := building.NewService(NewFakeRepository())
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
