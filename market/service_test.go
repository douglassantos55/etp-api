package market_test

import (
	"api/company"
	"api/market"
	"api/notification"
	"api/warehouse"
	"context"
	"log"
	"testing"
)

func TestMarketService(t *testing.T) {
	companySvc := company.NewService(company.NewFakeRepository())
	warehouseSvc := warehouse.NewService(warehouse.NewFakeRepository())

	service := market.NewService(market.NewFakeRepository(), companySvc, warehouseSvc, notification.NoOpNotifier(), log.Default())

	ctx := context.Background()

	t.Run("PlaceOrder", func(t *testing.T) {
		t.Run("should not place order without items in stock", func(t *testing.T) {
			_, err := service.PlaceOrder(ctx, &market.Order{
				CompanyId:  1,
				Quality:    1,
				Quantity:   5000,
				Price:      53025,
				ResourceId: 3,
			})

			expectedError := "not enough resources"
			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", err, expectedError)
			}
		})

		t.Run("should not place order without cash for transport fee", func(t *testing.T) {
			_, err := service.PlaceOrder(ctx, &market.Order{
				CompanyId:  1,
				Quantity:   1,
				ResourceId: 4,
				Quality:    2,
				Price:      200000,
			})

			expectedError := "not enough cash to pay transport fee"
			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", err, expectedError)
			}
		})

		t.Run("should set transport fee and sourcing cost", func(t *testing.T) {
			order, err := service.PlaceOrder(ctx, &market.Order{
				CompanyId:  1,
				Quality:    0,
				Quantity:   50,
				Price:      5,
				ResourceId: 2,
			})

			if err != nil {
				t.Fatalf("could not place order: %s", err)
			}

			expectedFee := uint64(388)
			if order.TransportFee != expectedFee {
				t.Errorf("expected transport fee %d, got %d", expectedFee, order.TransportFee)
			}

			expectedSourcingCost := uint64(1553)
			if order.SourcingCost != expectedSourcingCost {
				t.Errorf("expected sourcing cost %d, got %d", expectedSourcingCost, order.SourcingCost)
			}
		})

		t.Run("should reduce stocks", func(t *testing.T) {
			_, err := service.PlaceOrder(ctx, &market.Order{
				CompanyId:  1,
				Quality:    0,
				Quantity:   50,
				Price:      5,
				ResourceId: 2,
			})

			inventory, err := warehouseSvc.GetInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			stock := inventory.GetStock(2, 0)
			if stock != 600 {
				t.Errorf("expected stock %d, got %d", 600, stock)
			}
		})
	})

	t.Run("CancelOrder", func(t *testing.T) {
		t.Run("should increment stocks", func(t *testing.T) {
			order, err := service.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get order: %s", err)
			}

			if err := service.CancelOrder(ctx, order); err != nil {
				t.Fatalf("could not cancel order: %s", err)
			}

			inventory, err := warehouseSvc.GetInventory(ctx, 1)
			if err != nil {
				t.Fatalf("could not get inventory: %s", err)
			}

			stock := inventory.GetStock(2, 0)
			if stock != 750 {
				t.Errorf("expected stock %d, got %d", 750, stock)
			}
		})
	})
}
