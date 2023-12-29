package financing_test

import (
	"api/accounting"
	"api/company"
	"api/company/building"
	"api/database"
	"api/financing"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
	"time"
)

func TestFinancingRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
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
        INSERT INTO buildings (id, name) VALUES (1, "Mill"), (2, "Plantation")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO companies_buildings (id, name, company_id, building_id, position)
        VALUES (1, "Mill", 2, 1, 0), (2, "Plantation", 2, 2, 2), (3, "Store", 2, 2, 3)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO loans (id, company_id, principal, interest_rate, payable_from, interest_paid, delayed_payments) VALUES
        (1, 2, 100000000, 0.15, '2024-12-12 00:00:00', 15000000, 2),
        (2, 2, 100000000, 0.15, '2024-12-12 00:00:00', 15000000, 2),
        (3, 1, 100000000, 0.15, '2024-12-12 00:00:00', 15000000, 2)
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
		if _, err := conn.DB.Exec(`DELETE FROM companies_buildings`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM bonds_creditors`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM buildings`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM loans`); err != nil {
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
	repository := financing.NewRepository(conn, accountingRepo)

	t.Run("GetLoans", func(t *testing.T) {
		t.Run("should return empty list when not found", func(t *testing.T) {
			loans, err := repository.GetLoans(ctx, 3)
			if err != nil {
				t.Fatalf("could not get loans: %s", err)
			}

			if len(loans) != 0 {
				t.Errorf("expected %d loans, got %d", 0, len(loans))
			}
		})

		t.Run("should list company's loans", func(t *testing.T) {
			loans, err := repository.GetLoans(ctx, 2)
			if err != nil {
				t.Fatalf("could not get loans: %s", err)
			}

			if len(loans) != 2 {
				t.Errorf("expected %d loans, got %d", 2, len(loans))
			}
		})
	})

	t.Run("SaveLoan", func(t *testing.T) {
		loan, err := repository.SaveLoan(ctx, &financing.Loan{
			Principal:    1_000_000_00,
			CompanyId:    1,
			PayableFrom:  time.Now().Add(time.Second),
			InterestRate: 0.15,
		})

		if err != nil {
			t.Fatalf("could not save loan: %s", err)
		}

		if loan.Id == 0 {
			t.Error("should set an id")
		}

		company, err := companyRepo.GetById(ctx, 1)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableCash != 1_000_000_00 {
			t.Errorf("expected cash %d, got %d", 1_000_000_00, company.AvailableCash)
		}
	})

	t.Run("PayInterest", func(t *testing.T) {
		loan := &financing.Loan{
			Id:              1,
			InterestRate:    0.15,
			CompanyId:       2,
			Principal:       1_000_000_00,
			DelayedPayments: 2,
		}

		if err := repository.PayLoanInterest(ctx, loan); err != nil {
			t.Fatalf("could not pay interest: %s", err)
		}

		company, err := companyRepo.GetById(ctx, 2)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableCash != -150_000_00 {
			t.Errorf("expected cash %d, got %d", -150_000_00, company.AvailableCash)
		}

		loan, err = repository.GetLoan(ctx, 1, 2)
		if err != nil {
			t.Fatalf("could not get loan: %s", err)
		}

		if loan.InterestPaid != 300_000_00 {
			t.Errorf("expected interest paid %d, got %d", 300_000_00, loan.InterestPaid)
		}

		if loan.DelayedPayments != 0 {
			t.Errorf("shoud have reset delayed payments, got %d", loan.DelayedPayments)
		}
	})

	t.Run("ForcePrincipalPayment", func(t *testing.T) {
		err := repository.ForcePrincipalPayment(ctx, []int8{0, 1, 2}, &financing.Loan{
			Id:        1,
			CompanyId: 2,
			Principal: 1_000_000_00,
		})

		if err != nil {
			t.Fatalf("could not force payment: %s", err)
		}

		company, err := companyRepo.GetById(ctx, 2)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableTerrains != 0 {
			t.Errorf("expected 0 terrains left, got %d", company.AvailableTerrains)
		}

		loan, err := repository.GetLoan(ctx, 1, 2)
		if err != nil {
			t.Fatalf("could not get loan: %s", err)
		}

		if loan.PrincipalPaid != 1_000_000_00 {
			t.Errorf("expected principal paid %d, got %d", 1_000_000_00, loan.PrincipalPaid)
		}

		buildingsRepo := building.NewBuildingRepository(
			conn,
			resource.NewRepository(conn),
			warehouse.NewRepository(conn),
		)

		buildings, err := buildingsRepo.GetAll(ctx, 2)
		if err != nil {
			t.Fatalf("could not get buildings: %s", err)
		}

		if len(buildings) != 1 {
			t.Errorf("should have demolished buildings, got %d", len(buildings))
		}
	})

	t.Run("SaveBond", func(t *testing.T) {
		bond, err := repository.SaveBond(ctx, &financing.Bond{
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

		if company.AvailableCash != 2_000_000_00 {
			t.Errorf("expected %d cash, got %d", 2_000_000_00, company.AvailableCash)
		}
	})

	t.Run("GetBond", func(t *testing.T) {
		t.Run("should return error when not found", func(t *testing.T) {
			_, err := repository.GetBond(ctx, 345254)
			if err != financing.ErrBondNotFound {
				t.Errorf("expected error \"%s\", got \"%s\"", financing.ErrBondNotFound, err)
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

	t.Run("GetBonds", func(t *testing.T) {
		t.Run("should return empty list when no bonds found", func(t *testing.T) {
			bonds, err := repository.GetBonds(ctx, 4)
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
			bonds, err := repository.GetBonds(ctx, 2)
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
			bonds, err := repository.GetBonds(ctx, 2)
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
		err := repository.PayBondInterest(ctx, &financing.Bond{
			Id:           1,
			Amount:       2_000_000_00,
			CompanyId:    1,
			InterestRate: 0.15,
		}, &financing.Creditor{
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

		if emissor.AvailableCash != 1_850_000_00 {
			t.Errorf("expected cash %d, got %d", 1_850_000_00, emissor.AvailableCash)
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
		bond := &financing.Bond{
			Id:           1,
			Amount:       2_000_000_00,
			InterestRate: 0.15,
			Company:      &company.Company{Id: 1},
		}

		buyer := &financing.Creditor{
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

		if issuer.AvailableCash != 2_850_000_00 {
			t.Errorf("expected cash %d, got %d", 2_850_000_00, issuer.AvailableCash)
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

	t.Run("BuyBackLoan", func(t *testing.T) {
		loan, err := repository.BuyBackLoan(ctx, 500_000_00, &financing.Loan{
			Id:          2,
			CompanyId:   2,
			Principal:   1_000_000_00,
			PayableFrom: time.Now().Add(time.Second),
		})

		if err != nil {
			t.Fatalf("could not pay back loan: %s", err)
		}

		company, err := companyRepo.GetById(ctx, 2)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableCash != -650_000_00 {
			t.Errorf("expected cash %d, got %d", -650_000_00, company.AvailableCash)
		}

		if loan.PrincipalPaid != 500_000_00 {
			t.Errorf("expected principal paid %d, got %d", 500_000_00, loan.PrincipalPaid)
		}

		if loan.GetPrincipal() != 500_000_00 {
			t.Errorf("expected principal %d, got %d", 500_000_00, loan.GetPrincipal())
		}
	})

	t.Run("BuyBackBond", func(t *testing.T) {
		bond := &financing.Bond{Id: 3, CompanyId: 2}

		creditor := &financing.Creditor{
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

		if issuer.AvailableCash != -900_000_00 {
			t.Errorf("expected cash %d, got %d", -900_000_00, issuer.AvailableCash)
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
