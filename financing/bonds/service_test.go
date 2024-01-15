package bonds_test

import (
	"api/company"
	"api/financing/bonds"
	"api/notification"
	"context"
	"log"
	"testing"
	"time"
)

func TestBondService(t *testing.T) {
	companyRepo := company.NewFakeRepository()
	companySvc := company.NewService(companyRepo)
	service := bonds.NewService(bonds.NewFakeRepository(companyRepo), companySvc, notification.NoOpNotifier(), log.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("EmitBond", func(t *testing.T) {
		t.Run("should not allow bond greater than total terrain value", func(t *testing.T) {
			_, err := service.EmitBond(ctx, 0.15, 1_500_000_00, 1)
			expectedError := "amount must not be higher than 1300000.00"

			if err.Error() != expectedError {
				t.Errorf("expected error \"%s\", got \"%s\"", expectedError, err)
			}
		})
	})

	t.Run("PayBondInterest", func(t *testing.T) {
		t.Run("should skip if not enough cash", func(t *testing.T) {
			issuer, err := companyRepo.GetById(ctx, 1)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			initialCash := issuer.AvailableCash

			err = service.PayBondInterest(ctx, &bonds.Creditor{
				InterestRate:  0.1,
				Principal:     500_000_00,
				PrincipalPaid: 100_000_00,
				Company:       &company.Company{Name: "Bar"},
			}, &bonds.Bond{CompanyId: 1})

			if err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			if issuer.AvailableCash != initialCash {
				t.Errorf("expected cash %d, got %d", initialCash, issuer.AvailableCash)
			}
		})

		t.Run("should pay interest", func(t *testing.T) {
			company, err := companyRepo.GetById(ctx, 3)
			if err != nil {
				t.Fatalf("could not get company: %s", err)
			}

			initialCash := company.AvailableCash

			err = service.PayBondInterest(ctx, &bonds.Creditor{
				Principal:       1_500_000_00,
				InterestRate:    0.1,
				InterestPaid:    0,
				DelayedPayments: 0,
				PrincipalPaid:   0,
			}, &bonds.Bond{CompanyId: 3})

			if err != nil {
				t.Fatalf("could not pay interest: %s", err)
			}

			expectedCash := initialCash - 150_000_00
			if company.AvailableCash != expectedCash {
				t.Errorf("expected cash %d, got %d", expectedCash, company.AvailableCash)
			}
		})
	})

	t.Run("BuyBond", func(t *testing.T) {
		t.Run("should validate cash", func(t *testing.T) {
			_, _, err := service.BuyBond(ctx, 500_000_00, 1, 2)
			if err != bonds.ErrNotEnoughCash {
				t.Errorf("expected error \"%s\", got \"%s\"", bonds.ErrNotEnoughCash, err)
			}
		})

		t.Run("should not buy more than available", func(t *testing.T) {
			_, _, err := service.BuyBond(ctx, 600_000_00, 1, 3)
			if err != bonds.ErrAmountHigherThanAvailable {
				t.Errorf("expected error \"%s\", got \"%s\"", bonds.ErrAmountHigherThanAvailable, err)
			}
		})
	})

	t.Run("BuyBackBond", func(t *testing.T) {
		t.Run("should not buy back from other companies", func(t *testing.T) {
			_, err := service.BuyBackBond(ctx, 500_000_00, 1, 2, 2)
			if err != bonds.ErrBondNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", bonds.ErrBondNotFound, err)
			}
		})

		t.Run("should not buy back from creditor that does not exist", func(t *testing.T) {
			_, err := service.BuyBackBond(ctx, 500_000_00, 1, 3, 1)
			if err != bonds.ErrCreditorNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", bonds.ErrCreditorNotFound, err)
			}
		})

		t.Run("should not buy back more than available", func(t *testing.T) {
			_, err := service.BuyBackBond(ctx, 500_000_00, 1, 2, 1)
			if err != bonds.ErrAmountHigherThanPrincipal {
				t.Errorf("expected error \"%s\", got \"%s\"", bonds.ErrAmountHigherThanPrincipal, err)
			}
		})

		t.Run("should not buy back without enough money", func(t *testing.T) {
			_, err := service.BuyBackBond(ctx, 100_000_00, 1, 2, 1)
			if err != bonds.ErrNotEnoughCash {
				t.Errorf("expected error \"%s\", got \"%s\"", bonds.ErrNotEnoughCash, err)
			}
		})
	})
}
