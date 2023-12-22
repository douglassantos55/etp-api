package financing

import (
	"api/accounting"
	"api/database"
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		SaveLoan(ctx context.Context, loan *Loan) (*Loan, error)
		UpdateLoan(ctx context.Context, loan *Loan) (*Loan, error)

		PayInterest(ctx context.Context, loan *Loan) error
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

func (r *goquRepository) SaveLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Classification: accounting.LOAN,
			Value:          int(loan.Principal),
			Description:    fmt.Sprintf("Loan of %f.2", float64(loan.Principal/100)),
		},
		uint64(loan.CompanyId),
	); err != nil {
		return nil, err
	}

	result, err := tx.
		Insert(goqu.T("loans")).
		Rows(loan).
		Executor().
		Exec()

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

func (r *goquRepository) PayInterest(ctx context.Context, loan *Loan) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	interest := loan.GetInterest()
	principal := loan.GetPrincipal()

	if r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(interest),
			Classification: accounting.LOAN_INTEREST_PAYMENT,
			Description:    fmt.Sprintf("Interest payment over principal %f.2", float64(principal)/100),
		},
		uint64(loan.CompanyId),
	); err != nil {
		return err
	}

	loan.InterestPaid += int64(interest)

	tx.
		Update(goqu.T("loans")).
		Set(loan).
		Where(goqu.And(
			goqu.I("id").Eq(loan.Id),
			goqu.I("company_id").Eq(loan.CompanyId),
		)).
		Executor().
		Exec()

	return tx.Commit()
}
