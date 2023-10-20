package company

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		SaveCompany(company *Company) error
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
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
