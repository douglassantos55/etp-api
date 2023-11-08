package company_test

import (
	"api/auth"
	"api/building"
	"api/company"
	"api/database"
	"api/resource"
	"api/server"
	"api/warehouse"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeRepository struct {
	data      map[uint64]*company.Company
	buildings map[uint64]map[uint64]*company.CompanyBuilding
}

func NewFakeRepository() company.Repository {
	data := map[uint64]*company.Company{
		1: {Id: 1, Name: "Test", Email: "admin@test.com", Pass: "$2a$10$OBo6gtRDtR2g8X6S9Qn/Z.1r33jf6QYRSxavEIjG8UfrJ8MLQWRzy", AvailableCash: 720},
	}

	busyUntil := time.Now().Add(time.Minute)
	buildings := map[uint64]map[uint64]*company.CompanyBuilding{
		1: {
			1: {
				Id:        1,
				Name:      "Plantation",
				Level:     1,
				WagesHour: 100,
				AdminHour: 500,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 1000,
						Resource: &resource.Resource{
							Id:   1,
							Name: "Seeds",
							Requirements: []*resource.Item{
								{Qty: 15, Quality: 0, Resource: &resource.Resource{Id: 2}},
							},
						},
					},
				},
			},
			3: {
				Id:        3,
				Name:      "Laboratory",
				Level:     1,
				WagesHour: 1000000,
				AdminHour: 5000000,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 100,
						Resource: &resource.Resource{
							Id:   5,
							Name: "Vaccine",
							Requirements: []*resource.Item{
								{Qty: 15, Quality: 0, Resource: &resource.Resource{Id: 2}},
							},
						},
					},
				},
			},
			4: {
				Id:        4,
				Name:      "Factory",
				Level:     1,
				WagesHour: 10,
				AdminHour: 50,
				BusyUntil: &busyUntil,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 1000,
						Resource: &resource.Resource{
							Id:   6,
							Name: "Iron bar",
							Requirements: []*resource.Item{
								{Qty: 1500, Quality: 0, Resource: &resource.Resource{Id: 3}},
							},
						},
					},
				},
			},
		},
		2: {
			2: {
				Id:        2,
				Name:      "Plantation",
				Level:     1,
				WagesHour: 100,
				AdminHour: 500,
				Resources: []*building.BuildingResource{
					{
						QtyPerHours: 1000,
						Resource: &resource.Resource{
							Id:   1,
							Name: "Seeds",
							Requirements: []*resource.Item{
								{Qty: 15, Quality: 0, Resource: &resource.Resource{Id: 2}},
							},
						},
					},
				},
			},
		},
	}
	return &fakeRepository{data, buildings}
}

func (r *fakeRepository) Register(ctx context.Context, registration *company.Registration) (*company.Company, error) {
	id := uint64(len(r.data) + 1)
	company := &company.Company{
		Id:    id,
		Name:  registration.Name,
		Email: registration.Email,
		Pass:  registration.Password,
	}
	r.data[id] = company
	return r.GetById(ctx, id)
}

func (r *fakeRepository) GetById(ctx context.Context, id uint64) (*company.Company, error) {
	return r.data[id], nil
}

func (r *fakeRepository) GetByEmail(ctx context.Context, email string) (*company.Company, error) {
	for _, company := range r.data {
		if company.Email == email {
			return company, nil
		}
	}
	return nil, nil
}

func (r *fakeRepository) GetBuildings(ctx context.Context, companyId uint64) ([]*company.CompanyBuilding, error) {
	buildings := make([]*company.CompanyBuilding, 0)
	for _, building := range r.buildings[companyId] {
		buildings = append(buildings, building)
	}
	return buildings, nil
}

func (r *fakeRepository) GetBuilding(ctx context.Context, buildingId, companyId uint64) (*company.CompanyBuilding, error) {
	buildings, ok := r.buildings[companyId]
	if !ok {
		return nil, nil
	}

	companyBuilding, ok := buildings[buildingId]
	if !ok {
		return nil, nil
	}

	resources := make([]*building.BuildingResource, 0)
	for _, buildingResource := range companyBuilding.Resources {
		requirements := make([]*resource.Item, 0)
		for _, req := range buildingResource.Requirements {
			requirements = append(requirements, &resource.Item{
				Qty:        req.Qty,
				Quality:    req.Quality,
				ResourceId: req.ResourceId,
				Resource:   req.Resource,
			})
		}

		resources = append(resources, &building.BuildingResource{
			Resource: &resource.Resource{
				Id:           buildingResource.Id,
				Name:         buildingResource.Name,
				Requirements: requirements,
			},
			QtyPerHours: buildingResource.QtyPerHours,
		})
	}

	return &company.CompanyBuilding{
		Id:              companyBuilding.Id,
		Name:            companyBuilding.Name,
		WagesHour:       companyBuilding.WagesHour,
		AdminHour:       companyBuilding.AdminHour,
		MaintenanceHour: companyBuilding.MaintenanceHour,
		Level:           companyBuilding.Level,
		Position:        companyBuilding.Position,
		BusyUntil:       companyBuilding.BusyUntil,
		Resources:       resources,
	}, nil
}

func (r *fakeRepository) AddBuilding(ctx context.Context, companyId uint64, inventory *warehouse.Inventory, building *building.Building, position uint8) (*company.CompanyBuilding, error) {
	id := uint64(len(r.buildings) + 1)
	companyBuilding := &company.CompanyBuilding{
		Id:              id,
		Name:            building.Name,
		Position:        &position,
		Level:           1,
		WagesHour:       building.WagesHour,
		AdminHour:       building.AdminHour,
		MaintenanceHour: building.MaintenanceHour,
	}
	r.buildings[companyId][id] = companyBuilding
	return companyBuilding, nil
}

func (r *fakeRepository) Produce(ctx context.Context, companyId uint64, inventory *warehouse.Inventory, building *company.CompanyBuilding, item *resource.Item, totalCost int) (*company.Production, error) {
	finishesAt := time.Now().Add(time.Hour)
	r.buildings[companyId][building.Id].BusyUntil = &finishesAt

	return &company.Production{
		Item:       item,
		Id:         1,
		FinishesAt: finishesAt,
	}, nil
}

func (r *fakeRepository) RegisterTransaction(tx *database.DB, companyId, classificationId uint64, amount int, description string) error {
	r.data[companyId].AvailableCash += amount
	return nil
}

func (r *fakeRepository) CancelProduction(ctx context.Context, productionId, buildingId, companyId uint64) error {
	return nil
}

func TestCompanyRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	svr := server.NewServer()
	buildingSvc := building.NewService(building.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())
	svc := company.NewService(NewFakeRepository(), buildingSvc, warehouseSvc)
	company.CreateEndpoints(svr, svc)

	t.Run("should validate registration", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"","email":"coke","password":""}`)

		req := httptest.NewRequest("POST", "/companies/register", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected code %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response server.ValidationErrors

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not decode: %s", err)
		}

		if _, ok := response.Errors["name"]; !ok {
			t.Error("expected validation error for name")
		}
		if _, ok := response.Errors["email"]; !ok {
			t.Error("expected validation error for email")
		}
		if _, ok := response.Errors["password"]; !ok {
			t.Error("expected validation error for password")
		}
		if _, ok := response.Errors["confirm_password"]; !ok {
			t.Error("expected validation error for confirm_password")
		}
	})

	t.Run("should validate if passwords match", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"Test","email":"test@email.com","password":"123","confirm_password":"122"}`)

		req := httptest.NewRequest("POST", "/companies/register", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected code %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response server.ValidationErrors

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not decode: %s", err)
		}

		expectedMessage := "confirm_password must be equal to Password"
		if msg, ok := response.Errors["confirm_password"]; !ok || msg != expectedMessage {
			t.Errorf("expected validation error for confirm_password: %s, got %s", expectedMessage, msg)
		}
	})

	t.Run("should not return password", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"Coca-Cola","email":"coke@coke.com","password":"password","confirm_password":"password"}`)

		req := httptest.NewRequest("POST", "/companies/register", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

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

	t.Run("should validate login", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`email=test&password=`)

		req := httptest.NewRequest("POST", "/companies/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response server.ValidationErrors
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not parse json: %s", err)
		}

		if _, ok := response.Errors["email"]; !ok {
			t.Error("expected validation error for email")
		}
		if _, ok := response.Errors["password"]; !ok {
			t.Error("expected validation error for password")
		}
	})

	t.Run("should return bad request when email not found", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`email=test@test.com&password=123`)

		req := httptest.NewRequest("POST", "/companies/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should return bad request when password does not match", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`email=admin@test.com&password=123`)

		req := httptest.NewRequest("POST", "/companies/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("should send token when authenticated", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`email=admin@test.com&password=password`)

		req := httptest.NewRequest("POST", "/companies/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d. %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		response := make(map[string]string)
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not parse response: %s", err)
		}

		if _, ok := response["token"]; !ok {
			t.Error("expected token")
		}
	})

	t.Run("should return bad request invalid data", func(t *testing.T) {
		body := strings.NewReader(`{"building_id":"a","position":"b"}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("should validate building", func(t *testing.T) {
		body := strings.NewReader(`{"building_id":0,"position":0}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}

	})

	t.Run("should return unauthorized", func(t *testing.T) {
		body := strings.NewReader(`{"building_id":1,"position":1}`)

		req := httptest.NewRequest("POST", "/companies/2/buildings", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d: %s", http.StatusUnauthorized, rec.Code, rec.Body.String())
		}
	})

	t.Run("should return 422 not enough resources", func(t *testing.T) {
		body := strings.NewReader(`{"building_id":2,"position":1}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		}
		if strings.TrimSpace(rec.Body.String()) != "{\"message\":\"not enough resources\"}" {
			t.Errorf("expected not enough resources, got %s", rec.Body.String())
		}
	})

	t.Run("should return 422 producing on building that does not exist", func(t *testing.T) {
		body := strings.NewReader(`{"resource_id":1,"quantity":100,"quality":0}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings/5/productions", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		}
		if strings.TrimSpace(rec.Body.String()) != "{\"message\":\"building not found\"}" {
			t.Errorf("expected building not found, got %s", rec.Body.String())
		}
	})

	t.Run("should return 422 producing on busy building", func(t *testing.T) {
		body := strings.NewReader(`{"resource_id":6,"quantity":100,"quality":0}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings/4/productions", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		}
		if strings.TrimSpace(rec.Body.String()) != "{\"message\":\"building is busy\"}" {
			t.Errorf("expected building is busy, got %s", rec.Body.String())
		}
	})

	t.Run("should return 422 producing resource not available for building", func(t *testing.T) {
		body := strings.NewReader(`{"resource_id":5,"quantity":100,"quality":0}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings/1/productions", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		}
		if strings.TrimSpace(rec.Body.String()) != "{\"message\":\"resource not found\"}" {
			t.Errorf("expected resource not found, got %s", rec.Body.String())
		}
	})

	t.Run("should return 422 if not enough cash", func(t *testing.T) {
		body := strings.NewReader(`{"resource_id":5,"quantity":1,"quality":0}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings/3/productions", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		}
		if strings.TrimSpace(rec.Body.String()) != "{\"message\":\"not enough cash\"}" {
			t.Errorf("expected not enough cash, got %s", rec.Body.String())
		}
	})

	t.Run("cancel production", func(t *testing.T) {
		t.Run("should return 400 when invalid company id", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/a/buildings/1/productions/1", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})

		t.Run("should return 400 when invalid building id", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/1/buildings/a/productions/1", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})

		t.Run("should return 400 when invalid production id", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/1/buildings/2/productions/a", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})

		t.Run("should return 401 when other company", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/2/buildings/2/productions/1", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d: %s", http.StatusUnauthorized, rec.Code, rec.Body.String())
			}
		})

		t.Run("should return 422 building not found", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/1/buildings/2/productions/1", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			}
			if strings.TrimSpace(rec.Body.String()) != "{\"message\":\"building not found\"}" {
				t.Errorf("expected building not found, got %s", rec.Body.String())
			}
		})

		t.Run("should return 422 building not producing", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/1/buildings/3/productions/1", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			}
			if strings.TrimSpace(rec.Body.String()) != "{\"message\":\"no production in process\"}" {
				t.Errorf("expected no production in process, got %s", rec.Body.String())
			}
		})

		t.Run("should return 204 when canceled", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/1/buildings/4/productions/1", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Errorf("expected status %d, got %d: %s", http.StatusNoContent, rec.Code, rec.Body.String())
			}
		})
	})
}
