package company_test

import (
	"api/building"
	"api/company"
	"api/resource"
	"api/warehouse"
	"testing"
)

func TestCompanyService(t *testing.T) {
	buildingSvc := building.NewService(building.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())
	service := company.NewService(NewFakeRepository(), buildingSvc, warehouseSvc)

	t.Run("add building", func(t *testing.T) {
		t.Run("should error if building is not found", func(t *testing.T) {
			_, err := service.AddBuilding(1, 3, 0)
			if err.Error() != "building not found" {
				t.Errorf("should not find building: %s", err)
			}
		})

		t.Run("should error if cannot find company's inventory", func(t *testing.T) {
			_, err := service.AddBuilding(3, 1, 0)
			if err.Error() != "inventory not found" {
				t.Errorf("should not find inventory: %s", err)
			}
		})

		t.Run("should error if not enough resources", func(t *testing.T) {
			_, err := service.AddBuilding(1, 2, 1)
			if err.Error() != "not enough resources" {
				t.Errorf("should not have enough resources: %s", err)
			}
		})

		t.Run("should reduce stocks", func(t *testing.T) {
			building, err := service.AddBuilding(1, 1, 1)
			if err != nil {
				t.Errorf("should not fail: %s", err)
			}

			inventory, err := warehouseSvc.GetInventory(1)
			if err != nil {
				t.Fatalf("could not fetch inventory: %s", err)
			}

			if inventory.Items[0].Qty != 50 {
				t.Errorf("expected stock %d, got %d", 50, inventory.Items[0].Qty)
			}

			if building == nil {
				t.Fatal("should add building")
			}
			if *building.Position != 1 {
				t.Errorf("expected position %d, got %d", 1, *building.Position)
			}
			if building.Name != "Plantation" {
				t.Errorf("expected name %s, got %s", "Plantation", building.Name)
			}
		})
	})

	t.Run("produce", func(t *testing.T) {
		t.Run("should not produce on building from other companies", func(t *testing.T) {
			_, err := service.Produce(1, 2, &resource.Item{Qty: 10, Quality: 0, ResourceId: 1})
			if err.Error() != "building not found" {
				t.Errorf("expected building not found got %s", err)
			}
		})

		t.Run("should not produce resource that is not in building", func(t *testing.T) {
			_, err := service.Produce(1, 1, &resource.Item{Qty: 10, Quality: 0, ResourceId: 2})
			if err.Error() != "resource not found" {
				t.Errorf("expected resource not found, got %s", err)
			}
		})

		t.Run("should not produce without enough resources", func(t *testing.T) {
			_, err := service.Produce(1, 1, &resource.Item{Qty: 100, Quality: 0, ResourceId: 1})
			if err.Error() != "not enough resources" {
				t.Errorf("expected not enough resources, got %s", err)
			}
		})

		t.Run("should not produce without enough cash", func(t *testing.T) {
			_, err := service.Produce(1, 1, &resource.Item{Qty: 10, Quality: 0, ResourceId: 1})
			if err.Error() != "not enough cash" {
				t.Errorf("expected not enough cash, got %s", err)
			}
		})

		t.Run("should reduce cash", func(t *testing.T) {
			production, err := service.Produce(1, 1, &resource.Item{Qty: 1, Quality: 0, ResourceId: 1})
			if err != nil {
				t.Fatalf("could not produce: %s", err)
			}

			if production == nil {
				t.Fatal("should return a production instance")
			}

			company, err := service.GetById(1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			expectedCash := 360
			if company.AvailableCash != expectedCash {
				t.Errorf("expected %d cash, got %d", expectedCash, company.AvailableCash)
			}
		})

		t.Run("should not produce on busy building", func(t *testing.T) {
			_, err := service.Produce(1, 1, &resource.Item{Qty: 1, Quality: 0, ResourceId: 1})
			if err.Error() != "building is busy" {
				t.Errorf("expected building is busy, got %s", err)
			}
		})
	})
}
