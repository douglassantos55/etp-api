package research_test

import (
	"api/research"
	"api/scheduler"
	"context"
	"testing"
	"time"
)

func TestResearchService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	service := research.NewService(research.NewFakeRepository(), scheduler.NewScheduler())

	t.Run("GetGraduate", func(t *testing.T) {
		staff, err := service.GetGraduate(ctx, 1)
		if err != nil {
			t.Fatalf("could not find graduate: %s", err)
		}

		if staff.Id == 0 {
			t.Error("should have saved staff member")
		}

		if staff.Employer != 1 {
			t.Errorf("expected employer %d, got %d", 1, staff.Employer)
		}

		if staff.Status != research.PENDING {
			t.Errorf("expected status %d, got %d", research.PENDING, staff.Status)
		}

		if staff.Skill > 100 {
			t.Errorf("should not have skill higher than 100, got %d", staff.Skill)
		}

		if staff.Talent > 100 {
			t.Errorf("should not have talent higher than 100, got %d", staff.Talent)
		}

		if staff.Salary < 1000 || staff.Salary > 2000 {
			t.Errorf("salary should be between 1000 and 2000, got %d", staff.Salary)
		}
	})

	t.Run("GetExperienced", func(t *testing.T) {
		t.Run("should get from other companies", func(t *testing.T) {
			staff, err := service.GetExperienced(ctx, 1)
			if err != nil {
				t.Fatalf("could not find experienced: %s", err)
			}

			if staff.Id != 1 {
				t.Errorf("should have id %d, got %d", 1, staff.Id)
			}

			if staff.Employer != 2 {
				t.Error("should not have the same employer that is looking for a candidate")
			}

			if *staff.Poacher != 1 {
				t.Errorf("expected poacher %d, got %d", 1, *staff.Poacher)
			}
		})
	})

	t.Run("HireStaff", func(t *testing.T) {
		t.Run("graduate", func(t *testing.T) {
			staff, err := service.GetGraduate(ctx, 1)
			if err != nil {
				t.Fatalf("could not get graduate: %s", err)
			}

			staff, err = service.HireStaff(ctx, staff.Id)
			if err != nil {
				t.Fatalf("could not hire staff: %s", err)
			}

			if staff.Status != research.HIRED {
				t.Errorf("expected status %d, got %d", research.HIRED, staff.Status)
			}
		})

		t.Run("experienced", func(t *testing.T) {
			staff, err := service.GetExperienced(ctx, 1)
			if err != nil {
				t.Fatalf("could not get experienced: %s", err)
			}

			staff, err = service.HireStaff(ctx, staff.Id)
			if err != nil {
				t.Fatalf("could not hire staff: %s", err)
			}

			if staff.Status != research.HIRED {
				t.Errorf("expected status %d, got %d", research.HIRED, staff.Status)
			}

			if staff.Poacher != nil {
				t.Errorf("expected nil poacher, got %d", *staff.Poacher)
			}

			if staff.Salary != 2000000 {
				t.Errorf("expected salary %d, got %d", 2000000, staff.Salary)
			}

			if staff.Offer != 0 {
				t.Errorf("expected zero offer, got %d", staff.Offer)
			}

			if staff.Employer != 1 {
				t.Errorf("expected employer ID %d, got %d", 1, staff.Employer)
			}
		})
	})

	t.Run("MakeOffer", func(t *testing.T) {
		t.Run("should refuse lower than salary", func(t *testing.T) {
			offer := uint64(1000000)
			_, err := service.MakeOffer(ctx, offer, 1)

			expectedError := "offer is too low"
			if err.Error() != expectedError {
				t.Fatalf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("should save offer", func(t *testing.T) {
			offer := uint64(3000000)

			staff, err := service.MakeOffer(ctx, offer, 1)
			if err != nil {
				t.Fatalf("could not make offer: %s", err)
			}

			if staff.Offer != offer {
				t.Errorf("expected offer %d, got %d", offer, staff.Offer)
			}
		})
	})
}
