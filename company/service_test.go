package company_test

import (
	"api/building"
	"api/company"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
)

func TestCompanyService(t *testing.T) {
	buildingSvc := building.NewService(building.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())
	service := company.NewService(NewFakeRepository(), buildingSvc, warehouseSvc)

	ctx := context.Background()

	t.Run("add building", func(t *testing.T) {
		t.Run("should error if building is not found", func(t *testing.T) {
			_, err := service.AddBuilding(ctx, 1, 3, 0)
			if err.Error() != "building not found" {
				t.Errorf("should not find building: %s", err)
			}
		})

		t.Run("should error if cannot find company's inventory", func(t *testing.T) {
			_, err := service.AddBuilding(ctx, 3, 1, 0)
			if err.Error() != "inventory not found" {
				t.Errorf("should not find inventory: %s", err)
			}
		})

		t.Run("should error if not enough resources", func(t *testing.T) {
			_, err := service.AddBuilding(ctx, 1, 2, 1)
			if err.Error() != "not enough resources" {
				t.Errorf("should not have enough resources: %s", err)
			}
		})
	})

	t.Run("produce", func(t *testing.T) {
		t.Run("should not produce on building from other companies", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 2, &resource.Item{Qty: 10, Quality: 0, ResourceId: 1})
			if err.Error() != "building not found" {
				t.Errorf("expected building not found got %s", err)
			}
		})

		t.Run("should not produce resource that is not in building", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 1, &resource.Item{Qty: 10, Quality: 0, ResourceId: 2})
			if err.Error() != "resource not found" {
				t.Errorf("expected resource not found, got %s", err)
			}
		})

		t.Run("should not produce without enough resources", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 1, &resource.Item{Qty: 100, Quality: 0, ResourceId: 1})
			if err.Error() != "not enough resources" {
				t.Errorf("expected not enough resources, got %s", err)
			}
		})

		t.Run("should not produce without enough cash", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 1, &resource.Item{Qty: 10, Quality: 0, ResourceId: 1})
			if err.Error() != "not enough cash" {
				t.Errorf("expected not enough cash, got %s", err)
			}
		})

		t.Run("should not produce on busy building", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 1, &resource.Item{Qty: 1, Quality: 0, ResourceId: 1})
			_, err = service.Produce(ctx, 1, 1, &resource.Item{Qty: 1, Quality: 0, ResourceId: 1})
			if err.Error() != "building is busy" {
				t.Errorf("expected building is busy, got %s", err)
			}
		})
	})
}
