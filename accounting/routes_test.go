package accounting_test

import (
	"api/accounting"
	"api/auth"
	"api/scheduler"
	"api/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccountingRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	svc := accounting.NewService(accounting.NewFakeRepository(), scheduler.NewScheduler())

	svr := server.NewServer()
	accounting.CreateEndpoints(svr, svc)

	t.Run("Taxes", func(t *testing.T) {
		t.Run("no token", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/accounting/taxes", nil)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})

		t.Run("invalid token", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/accounting/taxes", nil)
			req.Header.Set("Authorization", "Bearer asotehu.noenthu.sasisis")
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})

		t.Run("not cron token", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/accounting/taxes", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})

		t.Run("cron token", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/accounting/taxes", nil)
			req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjcm9uIiwiYXVkIjoiY3JvbmpvYiIsImlzcyI6ImVudHJlcHJlbmV1ci1hcGkiLCJpYXQiOjE1MTYyMzkwMjJ9.mZmms00XUtw7JbObsSsC9pm40vk0ENdnJFU7Exp9rPU")
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
		})
	})

}
