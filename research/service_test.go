package research_test

import (
	"api/company"
	"api/research"
	"context"
	"testing"
	"time"
)

func TestResearchService(t *testing.T) {
	researchRepo := research.NewFakeRepository()
	companySvc := company.NewService(company.NewFakeRepository())
	service := research.NewService(researchRepo, companySvc)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("StartResearch", func(t *testing.T) {
		t.Run("busy staff", func(t *testing.T) {
			_, err := service.StartResearch(ctx, []uint64{1, 2}, 1, 1)
			if err != research.ErrBusyStaff {
				t.Fatalf("expected error %s, got %s", research.ErrBusyStaff, err)
			}

			_, err = service.StartResearch(ctx, []uint64{2, 3}, 1, 1)
			if err != research.ErrBusyStaff {
				t.Fatalf("expected error %s, got %s", research.ErrBusyStaff, err)
			}
		})

		t.Run("not enough cash", func(t *testing.T) {
			_, err := service.StartResearch(ctx, []uint64{3, 4, 5}, 1, 1)
			if err != research.ErrNotEnoughCash {
				t.Fatalf("expected error %s, got %s", research.ErrNotEnoughCash, err)
			}
		})

		t.Run("starts research", func(t *testing.T) {
			research, err := service.StartResearch(ctx, []uint64{3, 4, 5}, 1, 3)
			if err != nil {
				t.Fatalf("could not start research: %s", err)
			}

			expectedInvestment := 40000000
			if research.Investment != expectedInvestment {
				t.Errorf("expected investment %d, got %d", expectedInvestment, research.Investment)
			}

			expectedTime := time.Now().Add(12 * time.Hour)
			diff := research.FinishesAt.Sub(expectedTime)
			if int(diff.Seconds()) != 0 {
				t.Errorf("expected finishes at %+v, got %+v", expectedTime, research.FinishesAt)
			}
		})
	})

	t.Run("CompleteResearch", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			_, err := service.CompleteResearch(ctx, 99999)
			if err != research.ErrResearchNotFound {
				t.Fatalf("expected error %s, got %s", research.ErrResearchNotFound, err)
			}
		})

		t.Run("patents", func(t *testing.T) {
			research, err := service.CompleteResearch(ctx, 1)
			if err != nil {
				t.Fatalf("could not complete research: %s", err)
			}

			if research.CompletedAt.IsZero() {
				t.Error("should have set completed_at")
			}

			if research.Patents < 1 || research.Patents > 3 {
				t.Errorf("expected patents between 1 and 3, got %d", research.Patents)
			}
		})
	})
}
