package loans

import (
	"api/accounting"
	"api/database"
	"api/server"
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
)

var ErrLoanNotFound = server.NewBusinessRuleError("loan not found")

type (
	Repository interface {
		GetLoans(ctx context.Context, companyId int64) ([]*Loan, error)
		GetLoan(ctx context.Context, loanId, companyId int64) (*Loan, error)
		SaveLoan(ctx context.Context, loan *Loan) (*Loan, error)
		UpdateLoan(ctx context.Context, loan *Loan) (*Loan, error)
		PayLoanInterest(ctx context.Context, loan *Loan) error
		ForcePrincipalPayment(ctx context.Context, terrains []int8, loan *Loan) error
		BuyBackLoan(ctx context.Context, amount int64, loan *Loan) (*Loan, error)
	}

	goquRepository struct {
		builder        *goqu.Database
		accountingRepo accounting.Repository
	}
)

func NewRepository(conn *database.Connection, accountingRepo accounting.Repository) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder, accountingRepo}
}

func (r *goquRepository) GetLoans(ctx context.Context, companyId int64) ([]*Loan, error) {
	loans := make([]*Loan, 0)

	err := r.builder.
		Select(
			goqu.I("id"),
			goqu.I("interest_rate"),
			goqu.I("interest_paid"),
			goqu.I("payable_from"),
			goqu.I("principal"),
			goqu.I("principal_paid"),
			goqu.I("delayed_payments"),
			goqu.I("company_id"),
		).
		From(goqu.T("loans")).
		Where(goqu.I("company_id").Eq(companyId)).
		ScanStructsContext(ctx, &loans)

	if err != nil {
		return nil, err
	}

	return loans, nil
}

func (r *goquRepository) BuyBackLoan(ctx context.Context, amount int64, loan *Loan) (*Loan, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if _, err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(amount),
			Description:    "Loan buy back",
			Classification: accounting.LOAN_BUY_BACK,
		},
		uint64(loan.CompanyId),
	); err != nil {
		return nil, err
	}

	_, err = tx.
		Update(goqu.T("loans")).
		Set(goqu.Record{
			"principal_paid": goqu.L("? + ?", goqu.I("principal_paid"), amount),
		}).
		Where(goqu.And(
			goqu.I("id").Eq(loan.Id),
			goqu.I("company_id").Eq(loan.CompanyId),
		)).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetLoan(ctx, loan.Id, loan.CompanyId)
}

func (r *goquRepository) SaveLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if _, err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Classification: accounting.LOAN,
			Value:          int(loan.Principal),
			Description: fmt.Sprintf(
				"Loan of $ %.2f (%.2f%% interest rate)",
				float64(loan.Principal/100),
				loan.InterestRate,
			),
		},
		uint64(loan.CompanyId),
	); err != nil {
		return nil, err
	}

	result, err := tx.Insert(goqu.T("loans")).Rows(loan).Executor().Exec()
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	loan.Id = id
	return loan, nil
}

func (r *goquRepository) UpdateLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	_, err := r.builder.Update(goqu.T("loans")).Set(loan).Executor().ExecContext(ctx)
	if err != nil {
		return nil, err
	}
	return loan, nil
}

func (r *goquRepository) GetLoan(ctx context.Context, loanId, companyId int64) (*Loan, error) {
	loan := new(Loan)

	found, err := r.builder.
		Select(
			goqu.I("id"),
			goqu.I("interest_rate"),
			goqu.I("interest_paid"),
			goqu.I("payable_from"),
			goqu.I("principal"),
			goqu.I("principal_paid"),
			goqu.I("delayed_payments"),
			goqu.I("company_id"),
		).
		From(goqu.T("loans")).
		Where(goqu.And(
			goqu.I("id").Eq(loanId),
			goqu.I("company_id").Eq(companyId),
		)).
		ScanStructContext(ctx, loan)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrLoanNotFound
	}

	return loan, nil
}

func (r *goquRepository) PayLoanInterest(ctx context.Context, loan *Loan) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	interest := loan.GetInterest()
	principal := loan.GetPrincipal()

	if _, err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(interest),
			Classification: accounting.LOAN_INTEREST_PAYMENT,
			Description: fmt.Sprintf(
				"Interest payment over principal $ %.2f",
				float64(principal/100),
			),
		},
		uint64(loan.CompanyId),
	); err != nil {
		return err
	}

	tx.
		Update(goqu.T("loans")).
		Set(goqu.Record{
			"delayed_payments": 0,
			"interest_paid": goqu.L(
				"? + ?",
				goqu.I("interest_paid"),
				interest,
			),
		}).
		Where(goqu.And(
			goqu.I("id").Eq(loan.Id),
			goqu.I("company_id").Eq(loan.CompanyId),
		)).
		Executor().
		Exec()

	return tx.Commit()
}

func (r *goquRepository) ForcePrincipalPayment(ctx context.Context, terrains []int8, loan *Loan) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Demolish buildings in terrains
	_, err = tx.
		Update(goqu.T("companies_buildings")).
		Set(goqu.Record{"demolished_at": time.Now()}).
		Where(goqu.And(
			goqu.I("position").In(terrains),
			goqu.I("company_id").Eq(loan.CompanyId),
		)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	// Reduce the number of available terrains
	_, err = tx.Update(goqu.T("companies")).
		Set(goqu.Record{
			"available_terrains": goqu.L(
				"? - ?",
				goqu.I("available_terrains"),
				len(terrains),
			),
		}).
		Where(goqu.I("id").Eq(loan.CompanyId)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	tx.
		Update(goqu.T("loans")).
		Set(goqu.Record{
			"principal_paid": loan.GetPrincipal(),
		}).
		Where(goqu.And(
			goqu.I("id").Eq(loan.Id),
			goqu.I("company_id").Eq(loan.CompanyId),
		)).
		Executor().
		Exec()

	return tx.Commit()
}
