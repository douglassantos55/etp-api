package financing

import (
	"api/accounting"
	"api/database"
	"api/server"
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
)

var (
	ErrLoanNotFound = server.NewBusinessRuleError("loan not found")
	ErrBondNotFound = server.NewBusinessRuleError("bond not found")
)

type (
	Repository interface {
		GetLoan(ctx context.Context, loanId, companyId int64) (*Loan, error)
		SaveLoan(ctx context.Context, loan *Loan) (*Loan, error)
		UpdateLoan(ctx context.Context, loan *Loan) (*Loan, error)
		PayLoanInterest(ctx context.Context, loan *Loan) error
		ForcePrincipalPayment(ctx context.Context, terrains []int8, loan *Loan) error

		GetBonds(ctx context.Context, companyId int64) ([]*Bond, error)
		GetBond(ctx context.Context, bondId, companyId int64) (*Bond, error)
		SaveBond(ctx context.Context, bond *Bond) (*Bond, error)
		PayBondInterest(ctx context.Context, bond *Bond, creditor *Creditor) error
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
			Description:    fmt.Sprintf("Loan of %.2f", float64(loan.Principal/100)),
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

	if r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(interest),
			Classification: accounting.LOAN_INTEREST_PAYMENT,
			Description:    fmt.Sprintf("Interest payment over principal %.2f", float64(principal)/100),
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

func (r *goquRepository) GetBonds(ctx context.Context, companyId int64) ([]*Bond, error) {
	bonds := make([]*Bond, 0)

	err := r.builder.
		Select(
			goqu.I("id"),
			goqu.I("amount"),
			goqu.I("interest_rate"),
		).
		From(goqu.T("bonds")).
		Where(goqu.I("company_id").Eq(companyId)).
		ScanStructsContext(ctx, &bonds)

	if err != nil {
		return nil, err
	}

	for _, bond := range bonds {
		creditors, err := r.getCreditors(ctx, bond.Id)
		if err != nil {
			return nil, err
		}
		bond.Creditors = creditors
	}

	return bonds, nil
}

func (r *goquRepository) GetBond(ctx context.Context, bondId, companyId int64) (*Bond, error) {
	bond := new(Bond)

	found, err := r.builder.
		Select(
			goqu.I("id"),
			goqu.I("amount"),
			goqu.I("interest_rate"),
		).
		From(goqu.T("bonds")).
		Where(goqu.And(
			goqu.I("id").Eq(bondId),
			goqu.I("company_id").Eq(companyId),
		)).
		ScanStructContext(ctx, bond)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrBondNotFound
	}

	creditors, err := r.getCreditors(ctx, bond.Id)
	if err != nil {
		return nil, err
	}

	bond.Creditors = creditors
	return bond, nil
}

func (r *goquRepository) getCreditors(ctx context.Context, bondId int64) ([]*Creditor, error) {
	creditors := make([]*Creditor, 0)

	err := r.builder.
		Select(
			goqu.I("company_id").As(goqu.C("company.id")),
			goqu.I("interest_rate"),
			goqu.I("interest_paid"),
			goqu.I("principal"),
			goqu.I("principal_paid"),
			goqu.I("payable_from"),
			goqu.I("delayed_payments"),
		).
		From(goqu.T("bonds_creditors")).
		Where(goqu.I("bond_id").Eq(bondId)).
		ScanStructsContext(ctx, &creditors)

	if err != nil {
		return nil, err
	}

	return creditors, nil
}

func (r *goquRepository) SaveBond(ctx context.Context, bond *Bond) (*Bond, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          int(bond.Amount),
			Classification: accounting.BOND_EMISSION,
			Description: fmt.Sprintf(
				"Emission of %.2f in bonds (%.2f interest rate)",
				float64(bond.Amount)/100,
				bond.InterestRate,
			),
		},
		uint64(bond.CompanyId),
	)

	if err != nil {
		return nil, err
	}

	result, err := tx.Insert(goqu.T("bonds")).Rows(bond).Executor().Exec()
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	bond.Id = id
	return bond, nil
}

func (r *goquRepository) PayBondInterest(ctx context.Context, bond *Bond, creditor *Creditor) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(creditor.GetInterest()),
			Classification: accounting.BOND_INTEREST_PAYMENT,
			Description:    fmt.Sprintf("Bond interest payment over principal %.2f", float64(creditor.GetPrincipal()/100)),
		},
		uint64(bond.CompanyId),
	)

	if err != nil {
		return err
	}

	err = r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          int(creditor.GetInterest()),
			Classification: accounting.BOND_INTEREST_PAYMENT,
			Description:    fmt.Sprintf("Bond interest payment over principal %.2f", float64(creditor.GetPrincipal()/100)),
		},
		uint64(creditor.Id),
	)

	_, err = tx.
		Update(goqu.T("bonds_creditors")).
		Set(goqu.Record{
			"delayed_payments": 0,
			"interest_paid":    goqu.L("? + ?", goqu.I("interest_paid"), creditor.GetInterest()),
		}).
		Where(goqu.And(
			goqu.I("bond_id").Eq(bond.Id),
			goqu.I("company_id").Eq(creditor.Id),
		)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	return tx.Commit()
}
