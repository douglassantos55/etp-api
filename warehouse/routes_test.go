package warehouse_test

import (
	"api/auth"
	"api/resource"
	"api/server"
	"api/warehouse"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeRepository struct {
	data map[uint64]*warehouse.Inventory
}

func NewFakeRepository() warehouse.Repository {
	data := map[uint64]*warehouse.Inventory{
		1: {Items: []*warehouse.StockItem{
			{Cost: 137, Item: &resource.Item{Quality: 0, Qty: 100, Resource: &resource.Resource{Id: 1}}},
			{Cost: 47, Item: &resource.Item{Quality: 1, Qty: 1000, Resource: &resource.Resource{Id: 3}}},
			{Cost: 1553, Item: &resource.Item{Quality: 0, Qty: 700, Resource: &resource.Resource{Id: 2}}},
		}},
		2: {Items: []*warehouse.StockItem{
			{Cost: 525, Item: &resource.Item{Quality: 1, Qty: 50, Resource: &resource.Resource{Id: 1}}},
		}},
	}
	return &fakeRepository{data}
}

func (r *fakeRepository) FetchInventory(companyId uint64) (*warehouse.Inventory, error) {
	return r.data[companyId], nil
}

func (r *fakeRepository) ReduceStock(companyId uint64, inventory *warehouse.Inventory, items []*resource.Item) error {
	return nil
}

func TestWarehouseService(t *testing.T) {
	t.Setenv(server.JWT_SECRET_KEY, "secret")

	token, err := auth.GenerateToken(1, "secret")
	if err != nil {
		t.Fatalf("could not generate token: %s", err)
	}

	svr := server.NewServer()
	svc := warehouse.NewService(NewFakeRepository())
	warehouse.CreateEndpoints(svr, svc)

	t.Run("should return authenticated company's inventory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/warehouse", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		svr.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response *warehouse.Inventory
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("could not parse response: %s", err)
		}

		if len(response.Items) != 3 {
			t.Errorf("expected %d items, got %d", 3, len(response.Items))
		}
	})
}
