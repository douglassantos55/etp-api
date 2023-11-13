package building_test

import (
	"api/building"
	companyBuilding "api/company/building"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
)

func TestCompanyBuilding(t *testing.T) {
	companyBuilding := &companyBuilding.CompanyBuilding{
		WagesHour: 225000,
		AdminHour: 413500,
		Resources: []*building.BuildingResource{
			{
				QtyPerHours: 50,
				Resource: &resource.Resource{
					Id:           1,
					Requirements: []*resource.Item{},
				},
			},
			{
				QtyPerHours: 100,
				Resource: &resource.Resource{
					Id: 2,
					Requirements: []*resource.Item{
						{Qty: 10, Quality: 0, ResourceId: 1},
					},
				},
			},
		},
	}

	t.Run("GetResource", func(t *testing.T) {
		t.Run("get existing", func(t *testing.T) {
			found, err := companyBuilding.GetResource(1)
			if err != nil {
				t.Fatalf("could not find resource, got %s", err)
			}
			if found.Id != 1 {
				t.Errorf("expected resource ID %d, got %d", 1, found.Id)
			}

			found, err = companyBuilding.GetResource(2)
			if err != nil {
				t.Fatalf("could not find resource, got %s", err)
			}
			if found.Id != 2 {
				t.Errorf("expected resource ID %d, got %d", 2, found.Id)
			}
		})

		t.Run("should error if not found", func(t *testing.T) {
			found, err := companyBuilding.GetResource(3)
			if err == nil {
				t.Error("should not find resource")
			}
			if found != nil {
				t.Errorf("should not find resource, got %+v", found)
			}
		})
	})

	t.Run("GetProductionRequirements", func(t *testing.T) {
		t.Run("should consider qty", func(t *testing.T) {
			resourceToProduce := &resource.Item{Qty: 10, Quality: 0, ResourceId: 2}
			requirements, err := companyBuilding.GetProductionRequirements(resourceToProduce)
			if err != nil {
				t.Fatalf("could not get requirements: %s", err)
			}

			for _, req := range requirements {
				if req.Qty != 100 {
					t.Errorf("expected qty %d, got %d", 100, req.Qty)
				}
			}
		})

		t.Run("should error if resource not found", func(t *testing.T) {
			resourceToProduce := &resource.Item{Qty: 10, Quality: 0, ResourceId: 5}
			_, err := companyBuilding.GetProductionRequirements(resourceToProduce)

			expectedError := "resource not found"
			if err.Error() != expectedError {
				t.Errorf("expected error %s, got %s", expectedError, err)
			}
		})
	})

	t.Run("GetProductionTime", func(t *testing.T) {
		t.Run("should error if resource not found", func(t *testing.T) {
			resourceToProduce := &resource.Item{Qty: 10, Quality: 0, ResourceId: 5}
			_, err := companyBuilding.GetProductionTime(resourceToProduce)

			expectedError := "resource not found"
			if err.Error() != expectedError {
				t.Errorf("expected error %s, got %s", expectedError, err)
			}
		})

		t.Run("should return duration", func(t *testing.T) {
			resourceToProduce := &resource.Item{Qty: 200, Quality: 0, ResourceId: 2}
			duration, err := companyBuilding.GetProductionTime(resourceToProduce)

			if err != nil {
				t.Fatalf("could not get production duration: %s", err)
			}

			if duration != 120.0 {
				t.Errorf("expected duration %f, got %f", 120.0, duration)
			}

			resourceToProduce = &resource.Item{Qty: 25, Quality: 0, ResourceId: 1}
			duration, err = companyBuilding.GetProductionTime(resourceToProduce)

			if err != nil {
				t.Fatalf("could not get production duration: %s", err)
			}

			if duration != 30.0 {
				t.Errorf("expected duration %f, got %f", 30.0, duration)
			}
		})
	})

	t.Run("GetProductionCost", func(t *testing.T) {
		t.Run("should error if resource not found", func(t *testing.T) {
			resourceToProduce := &resource.Item{Qty: 10, Quality: 0, ResourceId: 5}
			_, err := companyBuilding.GetProductionCost(resourceToProduce)

			expectedError := "resource not found"
			if err.Error() != expectedError {
				t.Errorf("expected error %s, got %s", expectedError, err)
			}
		})

		t.Run("should return cost", func(t *testing.T) {
			resourceToProduce := &resource.Item{Qty: 100, Quality: 0, ResourceId: 2}
			cost, err := companyBuilding.GetProductionCost(resourceToProduce)

			if err != nil {
				t.Fatalf("could not get production cost: %s", err)
			}

			expectedCost := uint64(638500)
			if cost != expectedCost {
				t.Errorf("expected cost %d, got %d", expectedCost, cost)
			}

			resourceToProduce = &resource.Item{Qty: 25, Quality: 0, ResourceId: 1}
			cost, err = companyBuilding.GetProductionCost(resourceToProduce)

			if err != nil {
				t.Fatalf("could not get production cost: %s", err)
			}

			if cost != expectedCost/2 {
				t.Errorf("expected cost %d, got %d", expectedCost/2, cost)
			}
		})
	})
}

func TestBuildingService(t *testing.T) {
	repository := companyBuilding.NewFakeBuildingRepository()
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())
	buildingSvc := building.NewService(building.NewFakeRepository())
	service := companyBuilding.NewBuildingService(repository, warehouseSvc, buildingSvc)

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

		t.Run("should reduce stocks and set construction downtime", func(t *testing.T) {
			_, err := service.AddBuilding(ctx, 1, 1, 1)
			if err != nil {
				t.Fatalf("could not add building: %s", err)
			}

			inventory, err := warehouseSvc.GetInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			for _, item := range inventory.Items {
				if item.Resource.Id == 1 && item.Qty != 50 {
					t.Errorf("expected stock %d, got %d", 50, item.Qty)
				}

				if item.Resource.Id == 2 && item.Qty != 700 {
					t.Errorf("expected stock %d, got %d", 700, item.Qty)
				}

				if item.Resource.Id == 3 && item.Qty != 1000 {
					t.Errorf("expected stock %d, got %d", 1000, item.Qty)
				}
			}
		})
	})

	t.Run("demolish", func(t *testing.T) {
		t.Run("cannot demolish non existing building", func(t *testing.T) {
			err := service.Demolish(ctx, 1, 452)
			expectedError := "building not found"
			if err.Error() != expectedError {
				t.Errorf("expected error %s, got %s", expectedError, err)
			}
        })

		t.Run("cannot demolish busy building", func(t *testing.T) {
			err := service.Demolish(ctx, 1, 4)
			expectedError := "cannot demolish busy building"
			if err.Error() != expectedError {
				t.Errorf("expected error %s, got %s", expectedError, err)
			}
		})
	})
}
