package production_test

import (
	"api/auth"
	"api/building"
	"api/company"
	companyBuilding "api/company/building"
	"api/company/building/production"
	"api/research"
	"api/server"
	"api/warehouse"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductionRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	companySvc := company.NewService(company.NewFakeRepository())
	buildingSvc := building.NewService(building.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())

	researchSvc := research.NewService(research.NewFakeRepository(), companySvc)
	companyBuildingSvc := companyBuilding.NewBuildingService(companyBuilding.NewFakeBuildingRepository(), warehouseSvc, buildingSvc)
	svc := production.NewProductionService(production.NewFakeProductionRepository(), companySvc, companyBuildingSvc, warehouseSvc, researchSvc)

	svr := server.NewServer()
	production.CreateEndpoints(svr, svc, companyBuildingSvc, companySvc)

	t.Run("should return 422 producing on building that does not exist", func(t *testing.T) {
		body := strings.NewReader(`{"resource_id":1,"quantity":100,"quality":0}`)

		req := httptest.NewRequest("POST", "/companies/1/buildings/152/productions", body)
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
