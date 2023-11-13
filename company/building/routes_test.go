package building_test

import (
	"api/auth"
	"api/building"
	"api/company"
	companyBuilding "api/company/building"
	"api/server"
	"api/warehouse"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompanyBuildingRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	companySvc := company.NewService(company.NewFakeRepository())
	buildingSvc := building.NewService(building.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())
	svc := companyBuilding.NewBuildingService(companyBuilding.NewFakeBuildingRepository(), warehouseSvc, buildingSvc)

	svr := server.NewServer()
	companyBuilding.CreateEndpoints(svr, svc, companySvc)

	t.Run("AddBuilding", func(t *testing.T) {
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
	})

	t.Run("Demolish", func(t *testing.T) {
		t.Run("should return unauthorized", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/2/buildings/2", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})

		t.Run("should return 422 when building is not found", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/1/buildings/2", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rec.Code)
			}
		})

		t.Run("should return 422 when building is busy", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/companies/1/buildings/4", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rec.Code)
			}
		})
	})
}
