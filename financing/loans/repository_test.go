package loans_test

import (
	"api/accounting"
	"api/company"
	"api/company/building"
	"api/database"
	"api/financing/loans"
	"api/resource"
	"api/warehouse"
	"context"
	"testing"
	"time"
)

func TestLoansRepository(t *testing.T) {
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
		if _, err := conn.DB.Exec(`DELETE FROM buildings`); err != nil {
			t.Fatalf("could not cleanup database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM loans`); err != nil {
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
	repository := loans.NewRepository(conn, accountingRepo)

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
		loan, err := repository.SaveLoan(ctx, &loans.Loan{
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
		loan := &loans.Loan{
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
		err := repository.ForcePrincipalPayment(ctx, []int8{0, 1, 2}, &loans.Loan{
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

	t.Run("BuyBackLoan", func(t *testing.T) {
		loan, err := repository.BuyBackLoan(ctx, 500_000_00, &loans.Loan{
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
}
