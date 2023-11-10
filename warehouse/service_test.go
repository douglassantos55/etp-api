package warehouse_test

import (
	"api/resource"
	"api/warehouse"
	"testing"
)

func TestInventory(t *testing.T) {
	inventory := &warehouse.Inventory{
		Items: []*warehouse.StockItem{
			{Cost: 1053, Item: &resource.Item{Qty: 100, Quality: 0, Resource: &resource.Resource{Id: 1}}},
			{Cost: 752, Item: &resource.Item{Qty: 50, Quality: 0, Resource: &resource.Resource{Id: 2}}},
			{Cost: 805, Item: &resource.Item{Qty: 50, Quality: 1, Resource: &resource.Resource{Id: 2}}},
			{Item: &resource.Item{Qty: 100, Quality: 2, Resource: &resource.Resource{Id: 3}}},
		},
	}

	t.Run("GetStock", func(t *testing.T) {
		stock := inventory.GetStock(1, 0)
		if stock != 100 {
			t.Errorf("expected stock %d, got %d", 100, stock)
		}

		stock = inventory.GetStock(2, 0)
		if stock != 50 {
			t.Errorf("expected stock %d, got %d", 50, stock)
		}

		stock = inventory.GetStock(2, 1)
		if stock != 50 {
			t.Errorf("expected stock %d, got %d", 50, stock)
		}

		stock = inventory.GetStock(3, 0)
		if stock != 0 {
			t.Errorf("expected stock %d, got %d", 0, stock)
		}

		stock = inventory.GetStock(3, 1)
		if stock != 0 {
			t.Errorf("expected stock %d, got %d", 0, stock)
		}

		stock = inventory.GetStock(3, 2)
		if stock != 100 {
			t.Errorf("expected stock %d, got %d", 100, stock)
		}
	})

	t.Run("HasResources", func(t *testing.T) {
		t.Run("enough same quality", func(t *testing.T) {
			hasResources := inventory.HasResources([]*resource.Item{
				{Qty: 100, Quality: 0, Resource: &resource.Resource{Id: 1}},
			})
			if !hasResources {
				t.Errorf("should have enough resources: %d q%d of id %d", 100, 0, 1)
			}
		})

		t.Run("not enough same quality", func(t *testing.T) {
			hasResources := inventory.HasResources([]*resource.Item{
				{Qty: 200, Quality: 0, Resource: &resource.Resource{Id: 1}},
			})
			if hasResources {
				t.Errorf("should not have enough resources: %d q%d of id %d", 200, 0, 1)
			}
		})

		t.Run("enough different qualities", func(t *testing.T) {
			hasResources := inventory.HasResources([]*resource.Item{
				{Qty: 100, Quality: 0, Resource: &resource.Resource{Id: 2}},
			})
			if !hasResources {
				t.Errorf("should have enough resources: %d q%d of id %d", 100, 0, 2)
			}
		})

		t.Run("not enough different qualities", func(t *testing.T) {
			hasResources := inventory.HasResources([]*resource.Item{
				{Qty: 200, Quality: 0, Resource: &resource.Resource{Id: 2}},
			})
			if hasResources {
				t.Errorf("should not have enough resources: %d q%d of id %d", 200, 0, 2)
			}
		})

		t.Run("enough higher quality", func(t *testing.T) {
			hasResources := inventory.HasResources([]*resource.Item{
				{Qty: 100, Quality: 1, Resource: &resource.Resource{Id: 3}},
			})
			if !hasResources {
				t.Errorf("should have enough resources: %d q%d of id %d", 100, 1, 3)
			}
		})

		t.Run("not enough higher quality", func(t *testing.T) {
			hasResources := inventory.HasResources([]*resource.Item{
				{Qty: 200, Quality: 1, Resource: &resource.Resource{Id: 3}},
			})
			if hasResources {
				t.Errorf("should not have enough resources: %d q%d of id %d", 200, 1, 3)
			}
		})
	})

	t.Run("IncrementStock", func(t *testing.T) {
		t.Run("brand new resource", func(t *testing.T) {
			inventory.IncrementStock([]*warehouse.StockItem{
				{Item: &resource.Item{Qty: 1000, Quality: 0, Resource: &resource.Resource{Id: 6}}},
			})

			stock := inventory.GetStock(6, 0)
			if stock != 1000 {
				t.Errorf("expected stock %d, got %d", 1000, stock)
			}
		})

		t.Run("existing resource, same quality", func(t *testing.T) {
			inventory.IncrementStock([]*warehouse.StockItem{
				{Item: &resource.Item{Qty: 100, Quality: 2, Resource: &resource.Resource{Id: 3}}},
			})

			stock := inventory.GetStock(3, 2)
			if stock != 200 {
				t.Errorf("expected stock %d, got %d", 200, stock)
			}
		})

		t.Run("existing resource, different quality", func(t *testing.T) {
			inventory.IncrementStock([]*warehouse.StockItem{
				{Item: &resource.Item{Qty: 100, Quality: 1, Resource: &resource.Resource{Id: 3}}},
			})

			stock := inventory.GetStock(3, 1)
			if stock != 100 {
				t.Errorf("expected stock %d, got %d", 100, stock)
			}

			stock = inventory.GetStock(3, 2)
			if stock != 200 {
				t.Errorf("expected stock %d, got %d", 200, stock)
			}
		})
	})

	t.Run("ReduceStock", func(t *testing.T) {
		t.Run("same quality", func(t *testing.T) {
			cost := inventory.ReduceStock([]*resource.Item{
				{Qty: 50, Quality: 0, Resource: &resource.Resource{Id: 1}},
			})

			if cost != 1053 {
				t.Errorf("expected cost %d, got %d", 1053, cost)
			}

			stock := inventory.GetStock(1, 0)
			if stock != 50 {
				t.Errorf("expected stock %d, got %d", 50, stock)
			}
		})

		t.Run("different quality", func(t *testing.T) {
			cost := inventory.ReduceStock([]*resource.Item{
				{Qty: 90, Quality: 0, Resource: &resource.Resource{Id: 2}},
			})

			if cost != 775 {
				t.Errorf("expected cost %d, got %d", 775, cost)
			}

			stock := inventory.GetStock(2, 0)
			if stock != 0 {
				t.Errorf("expected stock %d, got %d", 0, stock)
			}

			stock = inventory.GetStock(2, 1)
			if stock != 10 {
				t.Errorf("expected stock %d, got %d", 10, stock)
			}
		})

		t.Run("higher quality", func(t *testing.T) {
			inventory.ReduceStock([]*resource.Item{
				{Qty: 10, Quality: 1, Resource: &resource.Resource{Id: 3}},
			})

			stock := inventory.GetStock(3, 2)
			if stock != 190 {
				t.Errorf("expected stock %d, got %d", 190, stock)
			}
		})
	})
}
