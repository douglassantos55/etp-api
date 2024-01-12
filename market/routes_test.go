package market_test

import (
	"api/auth"
	"api/company"
	"api/market"
	"api/notification"
	"api/server"
	"api/warehouse"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMarketRoutes(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate jwt token: %s", err)
	}

	svr := server.NewServer()

	companySvc := company.NewService(company.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())

	service := market.NewService(market.NewFakeRepository(), companySvc, warehouseSvc, notification.NoOpNotifier())

	market.CreateEndpoints(svr, service)

	t.Run("PlaceOrder", func(t *testing.T) {
		t.Run("should return bad request", func(t *testing.T) {
			body := strings.NewReader(`{"price":1523.7,"quality":"one","quantity":"751","resource_id":1}`)

			req := httptest.NewRequest("POST", "/market/orders", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})

		t.Run("should validate", func(t *testing.T) {
			body := strings.NewReader(`{"price":0,"quality":0,"quantity":0,"resource_id":0}`)

			req := httptest.NewRequest("POST", "/market/orders", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}

			var response server.ValidationErrors

			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("could not decode: %s", err)
			}

			expectedMessage := "price is a required field"
			if msg, ok := response.Errors["price"]; !ok || msg != expectedMessage {
				t.Errorf("expected validation error for price: %s, got %s", expectedMessage, msg)
			}

			expectedMessage = "quantity is a required field"
			if msg, ok := response.Errors["quantity"]; !ok || msg != expectedMessage {
				t.Errorf("expected validation error for quantity: %s, got %s", expectedMessage, msg)
			}
		})

		t.Run("should return created", func(t *testing.T) {
			body := strings.NewReader(`{"price":10,"quality":0,"quantity":1,"resource_id":1}`)

			req := httptest.NewRequest("POST", "/market/orders", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
			}
		})
	})

	t.Run("CancelOrder", func(t *testing.T) {
		t.Run("should not cancel from other companies", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/market/orders/2", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("expected status %d, got %d: %s", http.StatusForbidden, rec.Code, rec.Body.String())
			}
		})

		t.Run("should not cancel non existing order", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/market/orders/5142532", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("expected status %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
			}
		})

		t.Run("should not cancel invalid order ID", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/market/orders/somethinghere", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})
	})

	t.Run("Purchase", func(t *testing.T) {
		t.Run("should validate bind", func(t *testing.T) {
			body := strings.NewReader(`{"resource_id":1,"quality":"0","quantity":1}`)

			req := httptest.NewRequest("POST", "/market/orders/purchase", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})

		t.Run("should validate struct", func(t *testing.T) {
			body := strings.NewReader(`{"resource_id":0,"quality":0,"quantity":0}`)

			req := httptest.NewRequest("POST", "/market/orders/purchase", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}

			var validationErrors server.ValidationErrors
			if err := json.Unmarshal(rec.Body.Bytes(), &validationErrors); err != nil {
				t.Fatalf("could not parse json: %s", err)
			}

			expectedMessage := "resource_id is a required field"
			if msg, ok := validationErrors.Errors["resource_id"]; !ok || msg != expectedMessage {
				t.Errorf("expected validation error for price: %s, got %s", expectedMessage, msg)
			}

			expectedMessage = "quantity is a required field"
			if msg, ok := validationErrors.Errors["quantity"]; !ok || msg != expectedMessage {
				t.Errorf("expected validation error for quantity: %s, got %s", expectedMessage, msg)
			}
		})

		t.Run("should return ok", func(t *testing.T) {
			body := strings.NewReader(`{"resource_id":2,"quality":0,"quantity":1}`)

			req := httptest.NewRequest("POST", "/market/orders/purchase", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
			}
		})

		t.Run("should return unprocessable entity", func(t *testing.T) {
			body := strings.NewReader(`{"resource_id":10,"quality":0,"quantity":1}`)

			req := httptest.NewRequest("POST", "/market/orders/purchase", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			}
		})
	})

	t.Run("GetByResource", func(t *testing.T) {
		t.Run("should validate quality", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/market/orders?resource=1&quality=aoeu", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})

		t.Run("should validate resource", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/market/orders?resource=1&quality=aoeu", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})

		t.Run("should return orders", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/market/orders?resource=2&quality=0", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
			}

			var orders []*market.Order
			if err := json.Unmarshal(rec.Body.Bytes(), &orders); err != nil {
				t.Fatalf("could not parse json: %s", err)
			}

			if len(orders) != 2 {
				t.Errorf("expected %d orders, got %d", 2, len(orders))
			}
		})

		t.Run("should return empty orders", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/market/orders?resource=2&quality=1", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			svr.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
			}

			var orders []*market.Order
			if err := json.Unmarshal(rec.Body.Bytes(), &orders); err != nil {
				t.Fatalf("could not parse json: %s", err)
			}

			if len(orders) != 0 {
				t.Errorf("expected %d orders, got %d", 0, len(orders))
			}
		})
	})
}
