package company_test

import (
	"api/building"
	"api/company"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
	"time"
)

func TestProduction(t *testing.T) {
	now := time.Now()
	seed := &resource.Resource{Id: 1, Name: "Test"}

	production := &company.Production{
		Id:             1,
		FinishesAt:     now,
		LastCollection: nil,
		StartedAt:      now.Add(-3 * time.Hour),
		Building: &company.CompanyBuilding{
			Id:        1,
			Name:      "Plantation",
			BusyUntil: &now,
			Resources: []*building.BuildingResource{
				{Resource: seed, QtyPerHours: 50},
			},
		},
		Item: &resource.Item{
			Qty:        150,
			Quality:    0,
			ResourceId: 1,
			Resource:   seed,
		},
	}

	t.Run("no collection", func(t *testing.T) {
		produced, err := production.ProducedUntil(time.Now().Add(-1 * time.Hour))
		if err != nil {
			t.Fatalf("could not get production resources: %s", err)
		}

		if produced.Qty != 100 {
			t.Errorf("expected %d, got %d", 100, produced.Qty)
		}
	})

	t.Run("collection", func(t *testing.T) {
		lastCollection := now.Add(-2 * time.Hour)
		production.LastCollection = &lastCollection

		produced, err := production.ProducedUntil(time.Now().Add(-1 * time.Hour))
		if err != nil {
			t.Fatalf("could not get production resources: %s", err)
		}

		if produced.Qty != 50 {
			t.Errorf("expected %d, got %d", 50, produced.Qty)
		}
	})
}

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

		t.Run("should not cancel production of non existent building", func(t *testing.T) {
			err := service.CancelProduction(ctx, 1, 2, 1)
			expectedError := "building not found"

			if err.Error() != expectedError {
				t.Errorf("expected %s, got %s", expectedError, err)
			}
		})

		t.Run("should not cancel production of non busy building", func(t *testing.T) {
			err := service.CancelProduction(ctx, 1, 3, 1)
			expectedError := "no production in process"

			if err.Error() != expectedError {
				t.Errorf("expected %s, got %s", expectedError, err)
			}
		})
	})
}
