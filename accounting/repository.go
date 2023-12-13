package accounting

import (
	"api/database"

	"github.com/doug-martin/goqu/v9"
)

const (
	WAGES            = 1
	SOCIAL_CAPITAL   = 2
	TRANSPORT_FEE    = 3
	REFUNDS          = 4
	MARKET_PURCHASE  = 5
	MARKET_SALE      = 6
	MARKET_FEE       = 7
	TERRAIN_PURCHASE = 8
	STAFF_TRAINING   = 9
)

type (
	Repository interface {
		RegisterTransaction(tx *database.DB, transaction Transaction, companyId uint64) error
	}

	Transaction struct {
		Classification uint64
		Description    string
		Value          int
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) RegisterTransaction(tx *database.DB, transaction Transaction, companyId uint64) error {
	_, err := tx.
		Insert(goqu.T("transactions")).
		Rows(goqu.Record{
			"company_id":        companyId,
			"classification_id": transaction.Classification,
			"description":       transaction.Description,
			"value":             transaction.Value,
		}).
		Executor().
		Exec()

	return err
}
