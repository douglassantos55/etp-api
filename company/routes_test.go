package company_test

import (
	"api/auth"
	"api/building"
	"api/company"
	"api/server"
	"api/warehouse"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompanyRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	svr := server.NewServer()
	svc := company.NewService(company.NewFakeRepository())

	buildingSvc := building.NewService(building.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())

	companyBuildingSvc := company.NewBuildingService(company.NewFakeBuildingRepository(), warehouseSvc, buildingSvc)
	productionSvc := company.NewProductionService(company.NewFakeProductionRepository(), svc, companyBuildingSvc, warehouseSvc)

	company.CreateEndpoints(svr, svc, companyBuildingSvc, productionSvc)

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
