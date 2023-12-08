package staff_test

import (
	"api/auth"
	"api/research/staff"
	"api/scheduler"
	"api/server"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResearchRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	svr := server.NewServer()
	svc := staff.NewService(staff.NewFakeRepository(), scheduler.NewScheduler())

	staff.CreateEndpoints(svr, svc)
	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate token: %s", err)
	}

	t.Run("FindGraduate", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/research/staff/graduate", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var search *staff.Search
		if err := json.Unmarshal(rec.Body.Bytes(), &search); err != nil {
			t.Fatalf("could not parse json: %s", err)
		}

		duration := search.FinishesAt.Sub(time.Now())
		diff := staff.SEARCH_DURATION - duration
		if int(diff.Seconds()) != 0 {
			t.Errorf("expected duration %+v, got %d", search.FinishesAt, int(diff.Seconds()))
		}
	})

	t.Run("FindExperienced", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/research/staff/experienced", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var search *staff.Search
		if err := json.Unmarshal(rec.Body.Bytes(), &search); err != nil {
			t.Fatalf("could not parse json: %s", err)
		}

		duration := search.FinishesAt.Sub(time.Now())
		diff := staff.SEARCH_DURATION - duration
		if int(diff.Seconds()) != 0 {
			t.Errorf("expected duration %+v, got %d", search.FinishesAt, int(diff.Seconds()))
		}
	})

	t.Run("Raise", func(t *testing.T) {
		body := strings.NewReader(`{"salary":5100022}`)

		req := httptest.NewRequest("POST", "/research/staff/2/raise", body)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		staff := new(staff.Staff)
		if err := json.Unmarshal(rec.Body.Bytes(), staff); err != nil {
			t.Fatalf("could not parse json: %s", err)
		}

		if staff.Salary != 5100022 {
			t.Errorf("expected salary %d, got %d", 5100022, staff.Salary)
		}
	})
}
