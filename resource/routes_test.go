package resource_test

import (
	"api/auth"
	"api/resource"
	"api/server"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepository struct {
	data map[uint64]*resource.Resource
}

func NewFakeRepository() resource.Repository {
	data := map[uint64]*resource.Resource{
		1: {Id: 1, Name: "Water", Category: &resource.Category{Id: 1, Name: "Food"}},
		2: {Id: 2, Name: "Seeds", Category: &resource.Category{Id: 1, Name: "Food"}},
	}
	return &fakeRepository{data}
}

func (r *fakeRepository) FetchResources(ctx context.Context) ([]*resource.Resource, error) {
	items := make([]*resource.Resource, 0)
	for _, item := range r.data {
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeRepository) GetById(ctx context.Context, id uint64) (*resource.Resource, error) {
	return r.data[id], nil
}

func (r *fakeRepository) GetRequirements(ctx context.Context, resourceId uint64) ([]*resource.Item, error) {
	return nil, nil
}

func (r *fakeRepository) SaveResource(ctx context.Context, resource *resource.Resource) (*resource.Resource, error) {
	id := uint64(len(r.data) + 1)
	resource.Id = id
	r.data[id] = resource
	return resource, nil
}

func (r *fakeRepository) UpdateResource(ctx context.Context, resource *resource.Resource) (*resource.Resource, error) {
	r.data[resource.Id] = resource
	return resource, nil
}

func TestService(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")

	svr := server.NewServer()
	svc := resource.NewService(NewFakeRepository())
	resource.CreateEndpoints(svr, svc)

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	t.Run("should return 201 when creating resource", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/resources/", strings.NewReader(`{"name":"Wood","category_id":1,"image":"http://placeimg.com/10","requirements":[]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
		}
	})

	t.Run("should validate input when creating resource", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("POST", "/resources/", strings.NewReader(`{"name":"","image":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("should validate requirements", func(t *testing.T) {
        t.Parallel()

		body := strings.NewReader(`{"name":"Wood","category_id":1,"image":"http://placeimg.com/10","requirements":[{"quantity":0,"quality":0,"resource_id":0}]}`)

		req := httptest.NewRequest("POST", "/resources/", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("should validate input when updating resource", func(t *testing.T) {
        t.Parallel()

		req := httptest.NewRequest("PUT", "/resources/1", strings.NewReader(`{"name":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return 404 when updating non existing resource", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("PUT", "/resources/1052", strings.NewReader(`{"name":"Iron"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("should return 400 when updating invalid id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("PUT", "/resources/someid", strings.NewReader(`{"name":"Iron"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return 200 when updating resource", func(t *testing.T) {
		body := strings.NewReader(`{"name":"Iron","category_id":1,"requirements":[]}`)

		req := httptest.NewRequest("PUT", "/resources/1", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("should return 404 if resource is not found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/resources/2355", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("should return 400 if id is invalid", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/resources/someid", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return resource", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/resources/2", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

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
