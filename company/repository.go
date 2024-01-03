package company

import (
	"api/accounting"
	"api/database"
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type (
	Repository interface {
		Register(ctx context.Context, registration *Registration) (*Company, error)
		GetById(ctx context.Context, id uint64) (*Company, error)
		GetByEmail(ctx context.Context, email string) (*Company, error)
		PurchaseTerrain(ctx context.Context, total int, companyId uint64) error
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

func (r *goquRepository) GetById(ctx context.Context, id uint64) (*Company, error) {
	company := new(Company)

	found, err := r.getSelect().
		Where(r.getCondition().Append(goqu.I("c.id").Eq(id))).
		ScanStructContext(ctx, company)

	if err != nil || !found {
		return nil, err
	}

	return company, err
}

func (r *goquRepository) GetByEmail(ctx context.Context, email string) (*Company, error) {
	company := new(Company)

	found, err := r.getSelect().
		Where(r.getCondition().Append(goqu.I("c.email").Eq(email))).
		ScanStructContext(ctx, company)

	if err != nil || !found {
		return nil, err
	}

	return company, nil
}

func (r *goquRepository) PurchaseTerrain(ctx context.Context, total int, companyId uint64) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err = tx.
		Update(goqu.T("companies")).
		Set(goqu.Record{
			"available_terrains": goqu.L("? + 1", goqu.I("available_terrains")),
		}).
		Where(goqu.I("id").Eq(companyId)).
		Executor().
		Exec(); err != nil {
		return err
	}

	if _, err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -total,
			Description:    "Purchase of terrain",
			Classification: accounting.TERRAIN_PURCHASE,
		},
		companyId,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *goquRepository) Register(ctx context.Context, registration *Registration) (*Company, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	result, err := tx.
		Insert(goqu.T("companies")).
		Rows(goqu.Record{
			"name":     registration.Name,
			"email":    registration.Email,
			"password": registration.Password,
		}).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if _, err = r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Classification: accounting.SOCIAL_CAPITAL,
			Value:          1_000_000 * 100,
			Description:    "Initial capital",
		},
		uint64(id),
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetById(ctx, uint64(id))
}

func (r *goquRepository) getSelect() *goqu.SelectDataset {
	return r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
			goqu.I("c.is_admin"),
			goqu.I("c.available_terrains"),
			goqu.COALESCE(goqu.SUM("t.value"), 0).As("cash"),
		).
		From(goqu.T("companies").As("c")).
		LeftJoin(
			goqu.T("transactions").As("t"),
			goqu.On(goqu.I("t.company_id").Eq(goqu.I("c.id"))),
		).
		GroupBy(goqu.I("c.id"))
}

func (r *goquRepository) getCondition() exp.ExpressionList {
	return goqu.And(
		goqu.I("c.blocked_at").IsNull(),
		goqu.I("c.deleted_at").IsNull(),
	)
}
