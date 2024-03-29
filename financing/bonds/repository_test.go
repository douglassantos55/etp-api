package bonds_test

import (
	"api/accounting"
	"api/company"
	"api/database"
	"api/financing/bonds"
	"context"
	"testing"
	"time"
)

func TestBondRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../../test.db")
	if err != nil {
		t.Fatalf("could not connect to database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password) VALUES
        (1, "Coca-Cola", "coke@email.com", "aoeu"),
        (2, "Pepsi", "coke@email.com", "aoeu"),
        (3, "Tesla", "coke@email.com", "aoeu"),
        (4, "Amazon", "coke@email.com", "aoeu")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO bonds (id, company_id, amount, interest_rate) VALUES
        (1, 1, 200000000, 0.15), (2, 2, 100000000, 0.5), (3, 2, 100000000, 0.1)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO bonds_creditors (bond_id, company_id, principal, interest_rate, payable_from, delayed_payments) VALUES
        (1, 2, 100000000, 0.15, "2024-12-12 00:00:00", 1),
        (1, 3, 100000000, 0.15, "2024-12-12 00:00:00", 2),
        (3, 3, 50000000, 0.5, "2024-12-12 00:00:00", 0)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM transactions`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM bonds_creditors`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM bonds`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	accountingRepo := accounting.NewRepository(conn)
	companyRepo := company.NewRepository(conn, accountingRepo)
	repository := bonds.NewRepository(conn, accountingRepo)

	t.Run("GetBonds", func(t *testing.T) {
		t.Run("should paginate", func(t *testing.T) {
			bonds, err := repository.GetBonds(ctx, 0, 1)
			if err != nil {
				t.Fatalf("could not get bonds: %s", err)
			}

			if len(bonds) != 1 {
				t.Fatalf("expected %d bond, got %d", 1, len(bonds))
			}

			if bonds[0].Id != 2 {
				t.Errorf("expected bond ID %d, got %d", 2, bonds[0].Id)
			}

			bonds, err = repository.GetBonds(ctx, 1, 1)
			if err != nil {
				t.Fatalf("could not get bonds: %s", err)
			}

			if len(bonds) != 1 {
				t.Fatalf("expected %d bond, got %d", 1, len(bonds))
			}

			if bonds[0].Id != 3 {
				t.Errorf("expected bond ID %d, got %d", 3, bonds[0].Id)
			}
		})

		t.Run("should return empty list when not found", func(t *testing.T) {
			bonds, err := repository.GetBonds(ctx, 100, 50)
			if err != nil {
				t.Fatalf("could not get bonds: %s", err)
			}

			if len(bonds) != 0 {
				t.Errorf("expected empty bonds, got %d", len(bonds))
			}
		})

		t.Run("should calculate purchased", func(t *testing.T) {
			bonds, err := repository.GetBonds(ctx, 0, 10)
			if err != nil {
				t.Fatalf("could not get bonds: %s", err)
			}

			if len(bonds) != 2 {
				t.Fatalf("expected %d bonds, got %d", 2, len(bonds))
			}

			for _, bond := range bonds {
				if bond.Id == 1 {
					t.Errorf("should ignore purchased bond")
				}

				if bond.Id == 2 && bond.Purchased != 0 {
					t.Errorf("expected purchased %d, got %d", 0, bond.Purchased)
				}

				if bond.Id == 3 && bond.Purchased != 500_000_00 {
					t.Errorf("expected purchased %d, got %d", 500_000_00, bond.Purchased)
				}
			}
		})
	})

	t.Run("SaveBond", func(t *testing.T) {
		bond, err := repository.SaveBond(ctx, &bonds.Bond{
			Amount:       1_000_000_00,
			InterestRate: 0.15,
			CompanyId:    1,
		})

		if err != nil {
			t.Fatalf("could not save bond: %s", err)
		}

		if bond.Id == 0 {
			t.Error("should have set id")
		}

		company, err := companyRepo.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableCash != 1_000_000_00 {
			t.Errorf("expected %d cash, got %d", 1_000_000_00, company.AvailableCash)
		}
	})

	t.Run("GetBond", func(t *testing.T) {
		t.Run("should return error when not found", func(t *testing.T) {
			_, err := repository.GetBond(ctx, 345254)
			if err != bonds.ErrBondNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", bonds.ErrBondNotFound, err)
			}
		})

		t.Run("should include company", func(t *testing.T) {
			bond, err := repository.GetBond(ctx, 1)
			if err != nil {
				t.Fatalf("could not get bond: %s", err)
			}

			if bond.CompanyId != 1 {
				t.Errorf("expected company id %d, got %d", 1, bond.CompanyId)
			}

			if bond.Company.Id != 1 {
				t.Errorf("expected company id %d, got %d", 1, bond.Company.Id)
			}

			if bond.Company.Name != "Coca-Cola" {
				t.Errorf("expected name %s, got %s", "Coca-Cola", bond.Company.Name)
			}
		})

		t.Run("should calculate purchased", func(t *testing.T) {
			bond, err := repository.GetBond(ctx, 1)
			if err != nil {
				t.Fatalf("could not get bond: %s", err)
			}

			if bond.Purchased != 2_000_000_00 {
				t.Errorf("Expected purchased %d, got %d", 2_000_000_00, bond.Purchased)
			}

			bond, err = repository.GetBond(ctx, 2)
			if err != nil {
				t.Fatalf("could not get bond: %s", err)
			}

			if bond.Purchased != 0 {
				t.Errorf("Expected purchased %d, got %d", 0, bond.Purchased)
			}

			bond, err = repository.GetBond(ctx, 3)
			if err != nil {
				t.Fatalf("could not get bond: %s", err)
			}

			if bond.Purchased != 500_000_00 {
				t.Errorf("Expected purchased %d, got %d", 500_000_00, bond.Purchased)
			}
		})

		t.Run("should bring creditors", func(t *testing.T) {
			bond, err := repository.GetBond(ctx, 3)
			if err != nil {
				t.Fatalf("could not get bond: %s", err)
			}

			if len(bond.Creditors) != 1 {
				t.Errorf("expected %d creditor, got %d", 1, len(bond.Creditors))
			}

			if bond.Creditors[0].Name != "Tesla" {
				t.Errorf("Expected name %s, got %s", "Tesla", bond.Creditors[0].Name)
			}
		})
	})

	t.Run("GetCompanyBonds", func(t *testing.T) {
		t.Run("should return empty list when no bonds found", func(t *testing.T) {
			bonds, err := repository.GetCompanyBonds(ctx, 4)
			if err != nil {
				t.Fatalf("could not get bonds: %s", err)
			}

			if bonds == nil {
				t.Fatal("should return an empty slice")
			}

			if len(bonds) != 0 {
				t.Errorf("expected length %d, got %d", 0, len(bonds))
			}
		})

		t.Run("should calculate purchased", func(t *testing.T) {
			bonds, err := repository.GetCompanyBonds(ctx, 2)
			if err != nil {
				t.Fatalf("could not get bonds: %s", err)
			}

			if len(bonds) != 2 {
				t.Fatalf("expeced %d bond, got %d", 2, len(bonds))
			}

			for i, bond := range bonds {
				if i == 0 && bond.Purchased != 0 {
					t.Errorf("expected purchased %d, got %d", 0, bond.Purchased)
				}

				if i == 1 && bond.Purchased != 500_000_00 {
					t.Errorf("expected purchased %d, got %d", 500_000_00, bond.Purchased)
				}
			}
		})

		t.Run("should bring creditors", func(t *testing.T) {
			bonds, err := repository.GetCompanyBonds(ctx, 2)
			if err != nil {
				t.Fatalf("could not get bonds: %s", err)
			}

			for i, bond := range bonds {
				if i == 0 {
					if len(bond.Creditors) != 0 {
						t.Errorf("expected %d creditors, got %d", 0, len(bond.Creditors))
					}
				}
				if i == 1 {
					if len(bond.Creditors) != 1 {
						t.Fatalf("expected %d creditors, got %d", 1, len(bond.Creditors))
					}
					if bond.Creditors[0].Name != "Tesla" {
						t.Errorf("expected name %s, got %s", "Tesla", bond.Creditors[0].Name)
					}
				}
			}
		})
	})

	t.Run("PayBondInterest", func(t *testing.T) {
		err := repository.PayBondInterest(ctx, &bonds.Bond{
			Id:           1,
			Amount:       2_000_000_00,
			CompanyId:    1,
			InterestRate: 0.15,
		}, &bonds.Creditor{
			Company:         &company.Company{Id: 3},
			Principal:       1_000_000_00,
			InterestRate:    0.15,
			PayableFrom:     time.Now().Add(time.Second),
			InterestPaid:    0,
			PrincipalPaid:   0,
			DelayedPayments: 2,
		})

		if err != nil {
			t.Fatalf("could not pay bond interest: %s", err)
		}

		// Make sure emissor loses his money
		emissor, err := companyRepo.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if emissor.AvailableCash != 850_000_00 {
			t.Errorf("expected cash %d, got %d", 850_000_00, emissor.AvailableCash)
		}

		// Make sure creditor gets his money
		creditor, err := companyRepo.GetById(ctx, 3)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if creditor.AvailableCash != 150_000_00 {
			t.Errorf("expected cash %d, got %d", 150_000_00, creditor.AvailableCash)
		}

		// Make sure interest paid is incremented
		bond, err := repository.GetBond(ctx, 1)
		if err != nil {
			t.Fatalf("could not get bond: %s", err)
		}

		if bond.Creditors[1].InterestPaid != 150_000_00 {
			t.Errorf("expected interest paid %d, got %d", 150_000_00, bond.Creditors[1].InterestPaid)
		}

		// Make sure delayed payments is reset
		if bond.Creditors[1].DelayedPayments != 0 {
			t.Errorf("should reset delayed payments, got %d", bond.Creditors[1].DelayedPayments)
		}
	})

	t.Run("SaveCreditor", func(t *testing.T) {
		bond := &bonds.Bond{
			Id:           1,
			Amount:       2_000_000_00,
			InterestRate: 0.15,
			Company:      &company.Company{Id: 1},
		}

		buyer := &bonds.Creditor{
			Company:      &company.Company{Id: 4},
			InterestRate: 0.15,
			Principal:    1_000_000_00,
			PayableFrom:  time.Now().Add(time.Second),
		}

		_, err := repository.SaveCreditor(ctx, bond, buyer)
		if err != nil {
			t.Fatalf("could not save creditor: %s", err)
		}

		// Reduces creditor's money
		creditor, err := companyRepo.GetById(ctx, 4)
		if err != nil {
			t.Fatalf("could not get creditor: %s", err)
		}

		if creditor.AvailableCash != -1_000_000_00 {
			t.Errorf("expected cash %d, got %d", -1_000_000_00, creditor.AvailableCash)
		}

		// Increments issuer's money
		issuer, err := companyRepo.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get issuer: %s", err)
		}

		if issuer.AvailableCash != 1_850_000_00 {
			t.Errorf("expected cash %d, got %d", 1_850_000_00, issuer.AvailableCash)
		}

		// Adds to bond
		bond, err = repository.GetBond(ctx, 1)
		if err != nil {
			t.Fatalf("could not get bond: %s", err)
		}

		if len(bond.Creditors) != 3 {
			t.Errorf("expected %d creditors, got %d", 3, len(bond.Creditors))
		}
	})

	t.Run("BuyBackBond", func(t *testing.T) {
		bond := &bonds.Bond{Id: 3, CompanyId: 2}

		creditor := &bonds.Creditor{
			Company: &company.Company{Id: 3},
		}

		creditor, err := repository.BuyBackBond(ctx, 250_000_00, creditor, bond)
		if err != nil {
			t.Fatalf("could not buy back bond: %s", err)
		}

		// Check issuer transaction
		issuer, err := companyRepo.GetById(ctx, 2)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if issuer.AvailableCash != -250_000_00 {
			t.Errorf("expected cash %d, got %d", -250_000_00, issuer.AvailableCash)
		}

		// Check creditor transaction
		buyer, err := companyRepo.GetById(ctx, 3)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if buyer.AvailableCash != 400_000_00 {
			t.Errorf("expected cash %d, got %d", 400_000_00, buyer.AvailableCash)
		}

		// Check creditor update
		if creditor.Principal != 500_000_00 {
			t.Errorf("expected principal %d, got %d", 500_000_00, creditor.Principal)
		}
		if creditor.PrincipalPaid != 250_000_00 {
			t.Errorf("expected principal paid %d, got %d", 250_000_00, creditor.PrincipalPaid)
		}
		if creditor.GetPrincipal() != 250_000_00 {
			t.Errorf("expected principal %d, got %d", 250_000_00, creditor.GetPrincipal())
		}
	})
}
