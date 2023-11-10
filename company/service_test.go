package company_test

import (
	"api/building"
	"api/company"
	"api/resource"
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
