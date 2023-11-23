package company

import (
	"api/database"
	"context"

	"github.com/doug-martin/goqu/v9"
)

const (
	WAGES          = 1
	SOCIAL_CAPITAL = 2
	TRANSPORT_FEE  = 3
	REFUNDS        = 4
)

type (
	Repository interface {
		Register(ctx context.Context, registration *Registration) (*Company, error)
		GetById(ctx context.Context, id uint64) (*Company, error)
		GetByEmail(ctx context.Context, email string) (*Company, error)
		RegisterTransaction(tx *database.DB, companyId, classificationId uint64, amount int, description string) error
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) GetById(ctx context.Context, id uint64) (*Company, error) {
	company := new(Company)

	found, err := r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
			goqu.COALESCE(goqu.SUM("t.value"), 0).As("cash"),
		).
		From(goqu.T("companies").As("c")).
		LeftJoin(
			goqu.T("transactions").As("t"),
			goqu.On(goqu.I("t.company_id").Eq(goqu.I("c.id"))),
		).
		Where(
			goqu.And(
				goqu.I("c.id").Eq(id),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		GroupBy(goqu.I("c.id")).
		ScanStructContext(ctx, company)

	if err != nil || !found {
		return nil, err
	}

	return company, err
}

func (r *goquRepository) GetByEmail(ctx context.Context, email string) (*Company, error) {
	company := new(Company)

	found, err := r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
			goqu.COALESCE(goqu.SUM("t.value"), 0).As("cash"),
		).
		From(goqu.T("companies").As("c")).
		LeftJoin(
			goqu.T("transactions").As("t"),
			goqu.On(goqu.I("t.company_id").Eq(goqu.I("c.id"))),
		).
		Where(
			goqu.And(
				goqu.I("email").Eq(email),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		GroupBy(goqu.I("c.id")).
		ScanStructContext(ctx, company)

	if err != nil || !found {
		return nil, err
	}

	return company, nil
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

	if err = r.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		uint64(id),
		SOCIAL_CAPITAL,
		1_000_000*100,
		"Initial capital",
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetById(ctx, uint64(id))
}

func (r *goquRepository) RegisterTransaction(tx *database.DB, companyId, classificationId uint64, amount int, description string) error {
	_, err := tx.
		Insert(goqu.T("transactions")).
		Rows(goqu.Record{
			"company_id":        companyId,
			"classification_id": classificationId,
			"description":       description,
			"value":             amount,
		}).
		Executor().
		Exec()

	return err
}
