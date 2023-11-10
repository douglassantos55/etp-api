package company_test

import (
	"api/building"
	"api/company"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
)

func TestProductionService(t *testing.T) {
	companySvc := company.NewService(company.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())

	buildingSvc := building.NewService(building.NewFakeRepository())
	companyBuildingSvc := company.NewBuildingService(company.NewFakeBuildingRepository(), warehouseSvc, buildingSvc)

	repository := company.NewFakeProductionRepository()
	service := company.NewProductionService(repository, companySvc, companyBuildingSvc, warehouseSvc)

	ctx := context.Background()

	t.Run("Produce", func(t *testing.T) {
		t.Run("should not produce on building from other companies", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 2, &resource.Item{Qty: 10, Quality: 0, ResourceId: 1})
			if err.Error() != "building not found" {
				t.Errorf("expected building not found got %s", err)
			}
		})

		t.Run("should not produce resource that is not in building", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 1, &resource.Item{Qty: 10, Quality: 0, Resource: &resource.Resource{Id: 2}})
			if err.Error() != "resource not found" {
				t.Errorf("expected resource not found, got %s", err)
			}
		})

		t.Run("should not produce without enough cash", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 3, &resource.Item{
				Qty:      10,
				Quality:  0,
				Resource: &resource.Resource{Id: 5},
			})

			if err.Error() != "not enough cash" {
				t.Errorf("expected not enough cash, got %s", err)
			}
		})

		t.Run("should not produce without enough resources", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 1, &resource.Item{
				Qty:      100,
				Quality:  0,
				Resource: &resource.Resource{Id: 1},
			})

			if err.Error() != "not enough resources" {
				t.Errorf("expected not enough resources, got %s", err)
			}
		})

		t.Run("should set finishes at and reduce stock", func(t *testing.T) {
			production, err := service.Produce(ctx, 1, 1, &resource.Item{
				Qty:      1,
				Quality:  0,
				Resource: &resource.Resource{Id: 1},
			})
			if err != nil {
				t.Fatalf("could not produce: %s", err)
			}

			if production.FinishesAt.IsZero() {
				t.Error("should set finishes at")
			}

			inventory, err := warehouseSvc.GetInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			for _, item := range inventory.Items {
				if item.Resource.Id == 1 && item.Qty != 100 {
					t.Errorf("expected stock %d, got %d", 100, item.Qty)
				}

				if item.Resource.Id == 2 && item.Qty != 685 {
					t.Errorf("expected stock %d, got %d", 685, item.Qty)
				}

				if item.Resource.Id == 3 && item.Qty != 1000 {
					t.Errorf("expected stock %d, got %d", 1000, item.Qty)
				}
			}
		})

		t.Run("should not produce on busy building", func(t *testing.T) {
			_, err := service.Produce(ctx, 1, 1, &resource.Item{Qty: 1, Quality: 0, Resource: &resource.Resource{Id: 1}})

			if err.Error() != "building is busy" {
				t.Errorf("expected building is busy, got %s", err)
			}
		})
	})

	t.Run("CancelProduction", func(t *testing.T) {
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

		t.Run("should not cancel non existing production", func(t *testing.T) {
			err := service.CancelProduction(ctx, 1, 4, 6)
			expectedError := "production not found"

			if err.Error() != expectedError {
				t.Errorf("Expected %s, got %s", expectedError, err)
			}
		})

		t.Run("should set canceled at and increment inventory", func(t *testing.T) {
			err := service.CancelProduction(ctx, 1, 4, 1)
			if err != nil {
				t.Fatalf("could not cancel production: %s", err)
			}

			production, err := repository.GetProduction(ctx, 1, 4, 1)
			if err != nil {
				t.Fatalf("could not get production: %s", err)
			}

			if production.CanceledAt == nil {
				t.Error("should have set canceled at")
			}

			inventory, err := warehouseSvc.GetInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			var qty uint64
			var found bool
			for _, item := range inventory.Items {
				if item.Resource.Id == 1 && item.Qty != 100 {
					t.Errorf("expected stock %d, got %d", 100, item.Qty)
				}

				if item.Resource.Id == 2 && item.Qty != 685 {
					t.Errorf("expected stock %d, got %d", 685, item.Qty)
				}

				if item.Resource.Id == 3 && item.Qty != 1000 {
					t.Errorf("expected stock %d, got %d", 1000, item.Qty)
				}

				if item.Resource.Id == 6 {
					found = true
					qty = item.Qty
				}
			}

			if !found {
				t.Fatal("resource not found")
			}
			if qty != 1000 {
				t.Errorf("expected qty %d, got %d", 1000, qty)
			}
		})
	})

	t.Run("CollectResource", func(t *testing.T) {
		t.Run("should not collect from non existent building", func(t *testing.T) {
			_, err := service.CollectResource(ctx, 1, 2, 1)
			expectedError := "building not found"

			if err.Error() != expectedError {
				t.Errorf("expected %s, got %s", expectedError, err)
			}
		})

		t.Run("should not collect from non existing production", func(t *testing.T) {
			_, err := service.CollectResource(ctx, 1, 4, 6)
			expectedError := "production not found"

			if err.Error() != expectedError {
				t.Errorf("expected %s, got %s", expectedError, err)
			}
		})

		t.Run("should not collect from non busy building", func(t *testing.T) {
			err := service.CancelProduction(ctx, 1, 3, 1)
			expectedError := "no production in process"

			if err.Error() != expectedError {
				t.Errorf("expected %s, got %s", expectedError, err)
			}
		})

		t.Run("should set collected at and increment stock", func(t *testing.T) {
			collected, err := service.CollectResource(ctx, 1, 4, 2)
			if err != nil {
				t.Fatalf("could not cancel production: %s", err)
			}

			if collected.Qty != 1000 {
				t.Errorf("should have collected %d, got %d", 1000, collected.Qty)
			}

			production, err := repository.GetProduction(ctx, 2, 4, 1)
			if err != nil {
				t.Fatalf("could not get production: %s", err)
			}

			if production.LastCollection == nil {
				t.Error("should have set last collection")
			}

			inventory, err := warehouseSvc.GetInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			for _, item := range inventory.Items {
				if item.Resource.Id == 1 && item.Qty != 100 {
					t.Errorf("expected stock %d, got %d", 100, item.Qty)
				}

				if item.Resource.Id == 2 && item.Qty != 685 {
					t.Errorf("expected stock %d, got %d", 685, item.Qty)
				}

				if item.Resource.Id == 3 && item.Qty != 1000 {
					t.Errorf("expected stock %d, got %d", 1000, item.Qty)
				}

				if item.Resource.Id == 6 && item.Qty != 2000 {
					t.Errorf("expected stock %d, got %d", 2000, item.Qty)
				}
			}
		})
	})
}
