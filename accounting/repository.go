package accounting

import (
	"api/database"
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

const (
	WAGES                 = 1
	SOCIAL_CAPITAL        = 2
	TRANSPORT_FEE         = 3
	REFUNDS               = 4
	MARKET_PURCHASE       = 5
	MARKET_SALE           = 6
	MARKET_FEE            = 7
	TERRAIN_PURCHASE      = 8
	STAFF_TRAINING        = 9
	RESEARCH              = 10
	LOAN                  = 11
	LOAN_INTEREST_PAYMENT = 12
	BOND_EMISSION         = 13
	BOND_INTEREST_EXPENSE = 14
	BOND_INTEREST_INCOME  = 15
	BOND_PURCHASE         = 16
	BOND_PURCHASED        = 17
	LOAN_BUY_BACK         = 18
	BOND_BUY_BACK         = 19
	TAXES_PAID            = 20
	TAXES_DEFERRED        = 21
)

var INCOME_STATEMENT_CLASSIFICATIONS = []int{
	WAGES,
	TRANSPORT_FEE,
	REFUNDS,
	MARKET_PURCHASE,
	MARKET_SALE,
	MARKET_FEE,
	TAXES_PAID,
	BOND_INTEREST_EXPENSE,
	BOND_INTEREST_INCOME,
}

type (
	Repository interface {
		SaveTaxes(ctx context.Context, taxes, companyId int64) error
		GetPeriodResults(ctx context.Context, start, end time.Time) ([]*IncomeResult, error)
		RegisterTransaction(tx *database.DB, transaction Transaction, companyId uint64) (int64, error)
		GetIncomeTransactions(ctx context.Context, start, end time.Time, companyId int64) ([]*Transaction, error)
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) GetPeriodResults(ctx context.Context, start, end time.Time) ([]*IncomeResult, error) {
	results := make([]*IncomeResult, 0)

	err := r.builder.
		Select(
			goqu.I("c.id").As("company_id"),
			goqu.COALESCE(goqu.SUM(goqu.I("t.value")), 0).As("taxable_income"),
			goqu.
				Select(goqu.COALESCE(goqu.SUM(goqu.I("value")), 0)).
				From(goqu.T("transactions")).
				Where(goqu.And(
					goqu.I("company_id").Eq(goqu.I("c.id")),
					goqu.I("classification_id").Eq(TAXES_DEFERRED),
				)).
				As("deferred_taxes"),
		).
		From(goqu.T("companies").As("c")).
		LeftJoin(
			goqu.T("transactions").As("t"),
			goqu.On(
				goqu.And(
					goqu.I("t.company_id").Eq(goqu.I("c.id")),
					goqu.I("t.classification_id").In(INCOME_STATEMENT_CLASSIFICATIONS),
					goqu.I("t.created_at").Between(exp.NewRangeVal(
						start.Format(time.DateTime),
						end.Format(time.DateTime),
					)),
				),
			),
		).
		GroupBy(goqu.I("c.id")).
		ScanStructsContext(ctx, &results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *goquRepository) GetIncomeTransactions(ctx context.Context, start, end time.Time, companyId int64) ([]*Transaction, error) {
	transactions := make([]*Transaction, 0)

	err := r.builder.
		Select(
			goqu.I("classification_id"),
			goqu.SUM(goqu.I("value")).As("value"),
		).
		From(goqu.T("transactions")).
		Where(goqu.And(
			goqu.I("company_id").Eq(companyId),
			goqu.I("classification_id").In(append(
				INCOME_STATEMENT_CLASSIFICATIONS,
				TAXES_DEFERRED,
			)),
			goqu.Or(
				goqu.I("classification_id").Eq(TAXES_DEFERRED),
				goqu.I("created_at").Between(exp.NewRangeVal(
					start.Format(time.DateTime),
					end.Format(time.DateTime),
				)),
			),
		)).
		GroupBy(goqu.I("classification_id")).
		ScanStructsContext(ctx, &transactions)

	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *goquRepository) SaveTaxes(ctx context.Context, taxes int64, companyId int64) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	description := "Taxes paid"
	classification := TAXES_PAID

	if taxes < 0 {
		description = "Deferred taxes"
		classification = TAXES_DEFERRED
	} else {
		// If there are taxes to be paid, remove deferred cause they are
		// included on the taxes to be paid
		_, err = tx.
			Delete(goqu.T("transactions")).
			Where(goqu.And(
				goqu.I("company_id").Eq(companyId),
				goqu.I("classification_id").Eq(TAXES_DEFERRED),
			)).
			Executor().
			Exec()

		if err != nil {
			return err
		}
	}

	if _, err := r.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		Transaction{
			Value:          -int(taxes),
			Classification: uint64(classification),
			Description:    description,
		},
		uint64(companyId),
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *goquRepository) RegisterTransaction(tx *database.DB, transaction Transaction, companyId uint64) (int64, error) {
	result, err := tx.
		Insert(goqu.T("transactions")).
		Rows(goqu.Record{
			"company_id":        companyId,
			"classification_id": transaction.Classification,
			"description":       transaction.Description,
			"value":             transaction.Value,
		}).
		Executor().
		Exec()

	if err != nil {
		return -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}
