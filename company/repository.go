package company

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		Register(registration *Registration) (*Company, error)

		GetById(id uint64) (*Company, error)

		GetByEmail(email string) (*Company, error)
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) GetById(id uint64) (*Company, error) {
	company := new(Company)

	found, err := r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
		).
		From(goqu.T("companies").As("c")).
		Where(
			goqu.And(
				goqu.I("c.id").Eq(id),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		ScanStruct(company)

	if err != nil || !found {
		return nil, err
	}

	return company, err
}

func (r *goquRepository) GetByEmail(email string) (*Company, error) {
	company := new(Company)

	found, err := r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
		).
		From(goqu.T("companies").As("c")).
		Where(
			goqu.And(
				goqu.I("email").Eq(email),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		ScanStruct(company)

	if err != nil || !found {
		return nil, err
	}

	return company, nil
}

func (r *goquRepository) Register(registration *Registration) (*Company, error) {
	record := goqu.Record{
		"name":     registration.Name,
		"email":    registration.Email,
		"password": registration.Password,
	}

	result, err := r.builder.
		Insert(goqu.T("companies")).
		Rows(record).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetById(uint64(id))
}

func (r *goquRepository) getById(id uint64) (*Company, error) {
	company := new(Company)

	found, err := r.builder.
		Select(
			goqu.I("c.id"),
			goqu.I("c.name"),
			goqu.I("c.email"),
			goqu.I("c.password"),
			goqu.I("c.last_login"),
			goqu.I("c.created_at"),
		).
		From(goqu.T("companies").As("c")).
		Where(
			goqu.And(
				goqu.I("c.id").Eq(id),
				goqu.I("c.blocked_at").IsNull(),
				goqu.I("c.deleted_at").IsNull(),
			),
		).
		ScanStruct(company)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return company, nil
}
