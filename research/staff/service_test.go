package staff_test

import (
	"api/research/staff"
	"api/scheduler"
	"context"
	"testing"
	"time"
)

func TestResearchService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	repository := staff.NewFakeRepository()
	service := staff.NewService(repository, scheduler.NewScheduler())

	t.Run("GetGraduate", func(t *testing.T) {
		employee, err := service.GetGraduate(ctx, 1)
		if err != nil {
			t.Fatalf("could not find graduate: %s", err)
		}

		if employee.Id == 0 {
			t.Error("should have saved staff member")
		}

		if employee.Employer != 1 {
			t.Errorf("expected employer %d, got %d", 1, employee.Employer)
		}

		if employee.Status != staff.PENDING {
			t.Errorf("expected status %d, got %d", staff.PENDING, employee.Status)
		}

		if employee.Skill > 100 {
			t.Errorf("should not have skill higher than 100, got %d", employee.Skill)
		}

		if employee.Talent > 100 {
			t.Errorf("should not have talent higher than 100, got %d", employee.Talent)
		}

		if employee.Salary < 1000 || employee.Salary > 2000 {
			t.Errorf("salary should be between 1000 and 2000, got %d", employee.Salary)
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
			t.Run("for company", func(t *testing.T) {
				employee, err := service.GetGraduate(ctx, 1)
				if err != nil {
					t.Fatalf("could not get graduate: %s", err)
				}

				employee, err = service.HireStaff(ctx, employee.Id, 1)
				if err != nil {
					t.Fatalf("could not hire staff: %s", err)
				}

				if employee.Status != staff.HIRED {
					t.Errorf("expected status %d, got %d", staff.HIRED, employee.Status)
				}
			})

			t.Run("other company", func(t *testing.T) {
				employee, err := service.GetGraduate(ctx, 1)
				if err != nil {
					t.Fatalf("could not get graduate: %s", err)
				}

				employee, err = service.HireStaff(ctx, employee.Id, 2)
				if err != staff.ErrStaffNotFound {
					t.Errorf("expected \"%s\", got \"%s\"", staff.ErrStaffNotFound, err)
				}
			})
		})

		t.Run("experienced", func(t *testing.T) {
			t.Run("other company", func(t *testing.T) {
				employee, err := service.GetExperienced(ctx, 1)
				if err != nil {
					t.Fatalf("could not get experienced: %s", err)
				}

				employee, err = service.HireStaff(ctx, employee.Id, 2)
				if err != staff.ErrStaffNotFound {
					t.Errorf("expected \"%s\", got \"%s\"", staff.ErrStaffNotFound, err)
				}
			})

			t.Run("for company", func(t *testing.T) {
				employee, err := service.GetExperienced(ctx, 1)
				if err != nil {
					t.Fatalf("could not get experienced: %s", err)
				}

				employee, err = service.HireStaff(ctx, employee.Id, 1)
				if err != nil {
					t.Fatalf("could not hire staff: %s", err)
				}

				if employee.Status != staff.HIRED {
					t.Errorf("expected status %d, got %d", staff.HIRED, employee.Status)
				}

				if employee.Poacher != nil {
					t.Errorf("expected nil poacher, got %d", *employee.Poacher)
				}

				if employee.Salary != 2000000 {
					t.Errorf("expected salary %d, got %d", 2000000, employee.Salary)
				}

				if employee.Offer != 0 {
					t.Errorf("expected zero offer, got %d", employee.Offer)
				}

				if employee.Employer != 1 {
					t.Errorf("expected employer ID %d, got %d", 1, employee.Employer)
				}
			})
		})
	})

	t.Run("MakeOffer", func(t *testing.T) {
		t.Run("other company", func(t *testing.T) {
			employee, err := service.GetExperienced(ctx, 2)
			if err != nil {
				t.Fatalf("could not get experienced: %s", err)
			}

			_, err = service.MakeOffer(ctx, 15235, employee.Id, 1)
			if err != staff.ErrStaffNotFound {
				t.Fatalf("expected error \"%s\", got \"%s\"", staff.ErrStaffNotFound, err)
			}
		})

		t.Run("should refuse lower than salary", func(t *testing.T) {
			_, err := service.GetExperienced(ctx, 1)
			if err != nil {
				t.Fatalf("could not get experienced: %s", err)
			}

			offer := uint64(1000000)
			_, err = service.MakeOffer(ctx, offer, 1, 1)

			expectedError := "offer is too low"
			if err.Error() != expectedError {
				t.Fatalf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("should save offer", func(t *testing.T) {
			offer := uint64(3000000)

			staff, err := service.MakeOffer(ctx, offer, 1, 1)
			if err != nil {
				t.Fatalf("could not make offer: %s", err)
			}

			if staff.Offer != offer {
				t.Errorf("expected offer %d, got %d", offer, staff.Offer)
			}
		})
	})

	t.Run("IncreaseSalary", func(t *testing.T) {
		t.Run("less than current salary", func(t *testing.T) {
			_, err := service.IncreaseSalary(ctx, 2000000, 1, 1)
			expectedError := "new salary must be higher than current salary"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("less than offer", func(t *testing.T) {
			_, err := service.GetExperienced(ctx, 1)
			if err != nil {
				t.Fatalf("could not get experienced: %s", err)
			}

			newSalary := uint64(3000000)
			_, err = service.IncreaseSalary(ctx, newSalary, 1, 1)

			expectedError := "new salary must be higher than current offer"
			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})

		t.Run("removes offer", func(t *testing.T) {
			newSalary := uint64(3500000)
			staff, err := service.IncreaseSalary(ctx, newSalary, 1, 1)
			if err != nil {
				t.Fatalf("could not increase salary: %s", err)
			}

			if staff.Salary != newSalary {
				t.Errorf("expected salary %d, got %d", newSalary, staff.Salary)
			}

			if staff.Poacher != nil {
				t.Errorf("should have removed poacher, got %d", *staff.Poacher)
			}
		})
	})

	t.Run("Train", func(t *testing.T) {
		t.Run("other company", func(t *testing.T) {
			_, err := service.Train(ctx, 1, 3)

			if err != staff.ErrStaffNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", staff.ErrStaffNotFound, err)
			}
		})

		t.Run("duration", func(t *testing.T) {
			training, err := service.Train(ctx, 2, 1)
			if err != nil {
				t.Fatalf("could not train: %s", err)
			}

			expected := time.Now().Add(13 * time.Hour)
			diff := training.FinishesAt.Sub(expected)

			if int(diff.Seconds()) != 0 {
				t.Errorf("expected finishes at %+v, got %+v", expected, training.FinishesAt)
			}
		})

		t.Run("investment", func(t *testing.T) {
			training, err := service.Train(ctx, 2, 1)
			if err != nil {
				t.Fatalf("could not train: %s", err)
			}

			expected := uint64(6000000)
			if training.Investment != expected {
				t.Errorf("expected investiment %d, got %d", expected, training.Investment)
			}
		})
	})

	t.Run("FinishTraining", func(t *testing.T) {
		t.Run("result", func(t *testing.T) {
			if err := service.FinishTraining(ctx, 1, 1); err != nil {
				t.Fatalf("could not finish training: %s", err)
			}

			training, err := repository.GetTraining(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not get training: %s", err)
			}

			if training.Result == 0 {
				t.Error("should have set result")
			}

			if training.CompletedAt.IsZero() {
				t.Error("should have set completed at")
			}
		})
	})
}
