package financing_test

import (
	"api/accounting"
	"api/company"
	"api/database"
	"api/financing"
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
        (1, "Coca-Cola", "coke@email.com", "aoeu"), (2, "Coca-Cola", "coke@email.com", "aoeu")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO loans (id, company_id, principal, interest_rate, payable_from) VALUES
        (1, 1, 100000000, 0.15, '2024-12-12 00:00:00')
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
			Id:           1,
			InterestRate: 0.15,
			CompanyId:    2,
			Principal:    1_000_000_00,
		}

		if err := repository.PayInterest(ctx, loan); err != nil {
			t.Fatalf("could not pay interest: %s", err)
		}

		company, err := companyRepo.GetById(ctx, 2)
		if err != nil {
			t.Fatalf("could not get company: %s", err)
		}

		if company.AvailableCash != -150_000_00 {
			t.Errorf("expected cash %d, got %d", -150_000_00, company.AvailableCash)
		}
	})
}
