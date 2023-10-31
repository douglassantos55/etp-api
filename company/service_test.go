package company_test

import (
	"api/building"
	"api/company"
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
}
