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
	BOND_INTEREST_PAYMENT = 14
	BOND_PURCHASE         = 15
	BOND_PURCHASED        = 16
	LOAN_BUY_BACK         = 17
	BOND_BUY_BACK         = 18
	TAXES_PAID            = 19
	TAXES_DEFERRED        = 20
)

var INCOME_STATEMENT_CLASSIFICATIONS = []int{
	WAGES,
	TRANSPORT_FEE,
	REFUNDS,
	MARKET_PURCHASE,
	MARKET_SALE,
	MARKET_FEE,
	TAXES_PAID,
	TAXES_DEFERRED,
}

type (
	Repository interface {
		SaveTaxes(ctx context.Context, taxes, companyId int64) error
		GetTransactions(ctx context.Context, companyId int64) ([]*Transaction, error)
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

func (r *goquRepository) GetTransactions(ctx context.Context, companyId int64) ([]*Transaction, error) {
	transactions := make([]*Transaction, 0)

	err := r.builder.
		Select(
			goqu.I("value"),
			goqu.I("created_at"),
			goqu.I("classification_id"),
		).
		From(goqu.T("transactions")).
		Where(goqu.And(
			goqu.I("company_id").Eq(companyId),
			goqu.I("classification_id").In(INCOME_STATEMENT_CLASSIFICATIONS),
		)).
		ScanStructsContext(ctx, &transactions)

	if err != nil {
		return nil, err
	}

	return transactions, nil
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
			goqu.I("classification_id").In(INCOME_STATEMENT_CLASSIFICATIONS),
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
