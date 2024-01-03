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
	financing.CreateEndpoints(svr, svc, loansSvc, bondsSvc, companySvc)

	t.Run("GetRates", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/financing/rates", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response *financing.Rates
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not parse json: %s", err)
		}

		if response.Inflation == 0 {
			t.Error("should return inflation rate")
		}

		if response.Interest == 0 {
			t.Error("should return interest rate")
		}
	})

	t.Run("SaveRates", func(t *testing.T) {
		t.Run("no token", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/financing/rates", nil)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})

		t.Run("invalid token", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/financing/rates", nil)
			req.Header.Set("Authorization", "Bearer asotehu.noenthu.sasisis")
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})

		t.Run("not admin", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/financing/rates", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
			}
		})

		t.Run("admin", func(t *testing.T) {
			adminToken, err := auth.GenerateToken(3, "secret")
			if err != nil {
				t.Fatalf("could not generate jwt token: %s", err)
			}

			req := httptest.NewRequest("POST", "/financing/rates", nil)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
		})
	})
}
