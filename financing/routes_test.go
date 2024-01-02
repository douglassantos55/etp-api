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

	t.Run("GetRates", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/financing/rates", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response map[string]float64
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not parse json: %s", err)
		}

		if _, ok := response["inflation"]; !ok {
			t.Error("should return inflation rate")
		}

		if _, ok := response["interest"]; !ok {
			t.Error("should return interest rate")
		}
	})
}
