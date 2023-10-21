package company

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		SaveCompany(company *Company) error

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

func (r *goquRepository) GetByEmail(email string) (*Company, error) {
	company := new(Company)

	found, err := r.builder.Select(
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
				goqu.I("c.bloked_at").IsNull(),
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

func (r *goquRepository) SaveCompany(company *Company) error {
	record := goqu.Record{
		"name":     company.Name,
		"email":    company.Email,
		"password": company.Pass,
	}

	result, err := r.builder.
		Insert(goqu.T("companies")).
		Rows(record).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	company.Id = uint64(id)
	return nil
}
