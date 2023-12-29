package financing_test

import (
	"api/auth"
	"api/company"
	"api/financing"
	"api/server"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

		t.Run("EmitBond", func(t *testing.T) {
			t.Run("validation", func(t *testing.T) {
				body := strings.NewReader(`{"rate":0.7,"amount":100000000.00}`)

				req := httptest.NewRequest("POST", "/financing/bonds", body)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
				}
			})

			t.Run("success", func(t *testing.T) {
				body := strings.NewReader(`{"rate":0.17,"amount":100000000}`)

				req := httptest.NewRequest("POST", "/financing/bonds", body)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
				}
			})
		})

		t.Run("BuyBackBond", func(t *testing.T) {
			t.Run("invalid bond", func(t *testing.T) {
				body := strings.NewReader(`{"creditor_id":2,"amount":100}`)

				req := httptest.NewRequest("POST", "/financing/bonds/aoeu", body)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
				}
			})

			t.Run("non-existing bond", func(t *testing.T) {
				body := strings.NewReader(`{"creditor_id":2,"amount":100}`)

				req := httptest.NewRequest("POST", "/financing/bonds/1523", body)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusUnprocessableEntity {
					t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
				}
			})

			t.Run("validation", func(t *testing.T) {
				body := strings.NewReader(`{"creditor_id":2,"amount":"100"}`)

				req := httptest.NewRequest("POST", "/financing/bonds/1", body)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
				}
			})

			t.Run("success", func(t *testing.T) {
				body := strings.NewReader(`{"creditor_id":2,"amount":100}`)

				req := httptest.NewRequest("POST", "/financing/bonds/1", body)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				rec := httptest.NewRecorder()
				svr.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
				}
			})
		})
	})
}
