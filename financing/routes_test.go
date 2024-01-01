package financing_test

import (
	"api/auth"
	"api/company"
	"api/financing"
	"api/financing/bonds"
	"api/financing/loans"
	"api/server"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFinancingRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	svc := financing.NewService(financing.NewFakeRepository())

	companyRepo := company.NewFakeRepository()
	companySvc := company.NewService(companyRepo)

	loansSvc := loans.NewService(loans.NewFakeRepository(companyRepo), companySvc)
	bondsSvc := bonds.NewService(bonds.NewFakeRepository(companyRepo), companySvc)

	svr := server.NewServer()
	financing.CreateEndpoints(svr, svc, loansSvc, bondsSvc)

	t.Run("GetInflation", func(t *testing.T) {
		t.Run("no start", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/financing/inflation?end=2023-12-31", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})

		t.Run("no end", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/financing/inflation?start=2023-12-01", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})

		t.Run("no start/end", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/financing/inflation", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})

		t.Run("get inflation", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/financing/inflation?start=2023-12-01&end=2023-12-31", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			var response map[string]float64
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("could not parse json: %s", err)
			}

			if response["inflation"] != 0.125 {
				t.Errorf("expected %f, got %f", 0.125, response["inflation"])
			}
		})
	})
}
