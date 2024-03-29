package company_test

import (
	"api/auth"
	"api/company"
	"api/server"
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

	t.Run("PurchaseTerrain", func(t *testing.T) {
		t.Run("should validate position", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/companies/terrains/aoeu", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})

		t.Run("should validate token", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/companies/terrains/1", nil)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})

		t.Run("should return no content status", func(t *testing.T) {
			newToken, err := auth.GenerateToken(3, "secret")
			if err != nil {
				t.Fatalf("could not generate jwt token: %s", err)
			}

			req := httptest.NewRequest("POST", "/companies/terrains/1", nil)
			req.Header.Set("Authorization", "Bearer "+newToken)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
			}
		})
	})
}
