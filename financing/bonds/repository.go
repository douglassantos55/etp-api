package bonds

import (
	"api/accounting"
	"api/database"
	"api/server"
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
)

var (
	ErrBondNotFound = server.NewBusinessRuleError("bond not found")
)

type (
	Repository interface {
		GetBonds(ctx context.Context, page, limit uint) ([]*Bond, error)
		GetCompanyBonds(ctx context.Context, companyId int64) ([]*Bond, error)
		GetBond(ctx context.Context, bondId int64) (*Bond, error)
		SaveBond(ctx context.Context, bond *Bond) (*Bond, error)
		PayBondInterest(ctx context.Context, bond *Bond, creditor *Creditor) error
		SaveCreditor(ctx context.Context, bond *Bond, creditor *Creditor) (*Creditor, error)
		BuyBackBond(ctx context.Context, amount int64, creditor *Creditor, bond *Bond) (*Creditor, error)
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

func (r *goquRepository) GetBonds(ctx context.Context, page, limit uint) ([]*Bond, error) {
	bonds := make([]*Bond, 0)

	err := r.builder.
		Select(
			goqu.I("b.id"),
			goqu.I("b.amount"),
			goqu.I("b.company_id"),
			goqu.I("b.interest_rate"),
			goqu.COALESCE(goqu.SUM("bc.principal"), 0).As("purchased"),
		).
		From(goqu.T("bonds").As("b")).
		LeftJoin(
			goqu.T("bonds_creditors").As("bc"),
			goqu.On(goqu.I("b.id").Eq(goqu.I("bc.bond_id"))),
		).
		GroupBy(goqu.I("b.id")).
		Having(goqu.COALESCE(goqu.SUM("bc.principal"), 0).Lt(goqu.I("b.amount"))).
		Order(goqu.I("b.interest_rate").Desc()).
		Limit(limit).
		Offset(page*limit).
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

func (r *goquRepository) GetCompanyBonds(ctx context.Context, companyId int64) ([]*Bond, error) {
	bonds := make([]*Bond, 0)

	err := r.builder.
		Select(
			goqu.I("b.id"),
			goqu.I("b.amount"),
			goqu.I("b.company_id"),
			goqu.I("b.interest_rate"),
			goqu.COALESCE(goqu.SUM("bc.principal"), 0).As("purchased"),
		).
		From(goqu.T("bonds").As("b")).
		LeftJoin(
			goqu.T("bonds_creditors").As("bc"),
			goqu.On(goqu.I("b.id").Eq(goqu.I("bc.bond_id"))),
		).
		Where(goqu.I("b.company_id").Eq(companyId)).
		GroupBy(goqu.I("b.id")).
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

func (r *goquRepository) GetBond(ctx context.Context, bondId int64) (*Bond, error) {
	bond := new(Bond)

	found, err := r.builder.
		Select(
			goqu.I("b.id"),
			goqu.I("b.amount"),
			goqu.I("b.company_id"),
			goqu.I("b.interest_rate"),
			goqu.I("c.id").As(goqu.C("company.id")),
			goqu.I("c.name").As(goqu.C("company.name")),
			goqu.COALESCE(goqu.SUM("bc.principal"), 0).As("purchased"),
		).
		From(goqu.T("bonds").As("b")).
		LeftJoin(
			goqu.T("bonds_creditors").As("bc"),
			goqu.On(goqu.I("b.id").Eq(goqu.I("bc.bond_id"))),
		).
		InnerJoin(
			goqu.T("companies").As("c"),
			goqu.On(goqu.I("b.company_id").Eq(goqu.I("c.id"))),
		).
		Where(goqu.I("b.id").Eq(bondId)).
		GroupBy(goqu.I("b.id")).
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
			goqu.I("bc.interest_rate"),
			goqu.I("bc.interest_paid"),
			goqu.I("bc.principal"),
			goqu.I("bc.principal_paid"),
			goqu.I("bc.payable_from"),
			goqu.I("bc.delayed_payments"),
			goqu.I("c.id").As(goqu.C("company.id")),
			goqu.I("c.name").As(goqu.C("company.name")),
		).
		From(goqu.T("bonds_creditors").As("bc")).
		InnerJoin(
			goqu.T("companies").As("c"),
			goqu.On(goqu.I("bc.company_id").Eq(goqu.I("c.id"))),
		).
		Where(goqu.I("bc.bond_id").Eq(bondId)).
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
				"Issuance of $ %.2f in bonds (%.2f%% interest rate)",
				float64(bond.Amount/100),
				bond.InterestRate*100,
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
			Description: fmt.Sprintf(
				"Bond interest payment over principal $ %.2f",
				float64(creditor.GetPrincipal()/100),
			),
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
			Description: fmt.Sprintf(
				"Bond interest payment over principal $ %.2f",
				float64(creditor.GetPrincipal()/100),
			),
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

func (r *goquRepository) SaveCreditor(ctx context.Context, bond *Bond, creditor *Creditor) (*Creditor, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	// Transfer to issuer
	if err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          int(creditor.Principal),
			Classification: accounting.BOND_PURCHASED,
			Description: fmt.Sprintf(
				"Purchase of $ %.2f in bonds (%.2f%% interest rate)",
				float64(creditor.Principal/100),
				creditor.InterestRate*100,
			),
		},
		uint64(bond.Company.Id),
	); err != nil {
		return nil, err
	}

	// Remove from creditor
	if err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(creditor.Principal),
			Classification: accounting.BOND_PURCHASE,
			Description: fmt.Sprintf(
				"Purchase of $ %.2f in bonds (%.2f%% interest rate)",
				float64(creditor.Principal/100),
				creditor.InterestRate*100,
			),
		},
		creditor.Id,
	); err != nil {
		return nil, err
	}

	_, err = tx.
		Insert(goqu.T("bonds_creditors")).
		Rows(goqu.Record{
			"bond_id":       bond.Id,
			"company_id":    creditor.Id,
			"principal":     creditor.Principal,
			"interest_rate": creditor.InterestRate,
			"payable_from":  creditor.PayableFrom,
		}).
		Executor().
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return creditor, nil
}

func (r *goquRepository) BuyBackBond(ctx context.Context, amount int64, creditor *Creditor, bond *Bond) (*Creditor, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(amount),
			Description:    "Bond buy back",
			Classification: accounting.BOND_BUY_BACK,
		},
		uint64(bond.CompanyId),
	)

	if err != nil {
		return nil, err
	}

	err = r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          int(amount),
			Description:    "Bond buy back",
			Classification: accounting.BOND_BUY_BACK,
		},
		creditor.Id,
	)

	if err != nil {
		return nil, err
	}

	_, err = tx.
		Update(goqu.T("bonds_creditors")).
		Set(goqu.Record{
			"principal_paid": goqu.L("? + ?", goqu.I("principal_paid"), amount),
		}).
		Where(goqu.And(
			goqu.I("bond_id").Eq(bond.Id),
			goqu.I("company_id").Eq(creditor.Id),
		)).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	bond, err = r.GetBond(ctx, bond.Id)
	if err != nil {
		return nil, err
	}

	return bond.GetCreditor(int64(creditor.Id))
}
