package financing_test

import (
	"api/auth"
	"api/company"
	"api/financing"
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

	companyRepo := company.NewFakeRepository()
	companySvc := company.NewService(companyRepo)
	svc := financing.NewService(financing.NewFakeRepository(companyRepo), companySvc)

	svr := server.NewServer()
	financing.CreateEndpoints(svr, svc)

	t.Run("BondRoutes", func(t *testing.T) {
		t.Run("GetAll", func(t *testing.T) {
			t.Run("no pagination", func(t *testing.T) {
				req := httptest.NewRequest("GET", "/financing/bonds", nil)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
				}

				var response []*financing.Bond
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("could not parse json: %s", err)
				}

				if len(response) != 2 {
					t.Errorf("expected %d bonds, got %d", 2, len(response))
				}
			})

			t.Run("limit", func(t *testing.T) {
				req := httptest.NewRequest("GET", "/financing/bonds?limit=1", nil)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
				}

				var response []*financing.Bond
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("could not parse json: %s", err)
				}

				if len(response) != 1 {
					t.Fatalf("expected %d bonds, got %d", 1, len(response))
				}
			})
		})

		t.Run("GetCompanyBonds", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/financing/bonds?company=3", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
			}

			var response []*financing.Bond
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("could not parse json: %s", err)
			}

			if len(response) != 1 {
				t.Errorf("expected %d bonds, got %d", 1, len(response))
			}
		})
	})
}
