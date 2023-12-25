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
        (2, "Coca-Cola", "coke@email.com", "aoeu"),
        (3, "Coca-Cola", "coke@email.com", "aoeu")
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
        (1, 2, 100000000, 0.15, '2024-12-12 00:00:00', 15000000, 2)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO bonds (id, company_id, amount, interest_rate, purchased) VALUES
        (1, 1, 200000000, 0.15, 100000000)
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO bonds_creditors (bond_id, company_id, principal, interest_rate, payable_from, delayed_payments) VALUES
        (1, 2, 100000000, 0.15, "2024-12-12 00:00:00", 1), (1, 3, 100000000, 0.15, "2024-12-12 00:00:00", 2)
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
		creditor, err := companyRepo.GetById(ctx, 2)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if creditor.AvailableCash != -150_000_00 {
			t.Errorf("expected cash %d, got %d", -150_000_00, creditor.AvailableCash)
		}

		// Make sure interest paid is incremented
		bond, err := repository.GetBond(ctx, 1, 1)
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
}
