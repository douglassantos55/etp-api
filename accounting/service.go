package accounting

import (
	"context"
	"time"
)

const TAX_RATE = 0.15

type (
	IncomeStatement struct {
		categories map[uint64]int
	}

	Transaction struct {
		Value          int       `db:"value"`
		Description    string    `db:"description"`
		Classification uint64    `db:"classification_id"`
		CreatedAt      time.Time `db:"created_at" goqu:"skipinsert,skipupdate"`
	}

	Service interface {
		PayTaxes(ctx context.Context, start, end time.Time, companyId int64) (int64, error)
	}

	service struct {
		repository Repository
	}
)

func NewIncomeStatement(transactions []*Transaction) *IncomeStatement {
	categories := make(map[uint64]int)
	for _, transaction := range transactions {
		categories[transaction.Classification] = transaction.Value
	}
	return &IncomeStatement{categories}
}

func (s *IncomeStatement) GetTaxableIncome() int64 {
	var total int64
	for _, value := range s.categories {
		total += int64(value)
	}
	return total - s.GetDeferredTaxes()
}

func (s *IncomeStatement) GetDeferredTaxes() int64 {
	return int64(s.categories[TAXES_DEFERRED])
}

func (s *IncomeStatement) GetTaxes() int64 {
	if taxes, ok := s.categories[TAXES_PAID]; ok {
		return int64(taxes)
	}

	taxableIncome := s.GetTaxableIncome()
	taxes := int64(float64(taxableIncome) * TAX_RATE)

	if taxes > 0 {
		taxes -= s.GetDeferredTaxes()
	}

	return int64(taxes)
}

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetIncomeStatement(ctx context.Context, start, end time.Time, companyId int64) (*IncomeStatement, error) {
	transactions, err := s.repository.GetIncomeTransactions(ctx, start, end, companyId)
	if err != nil {
		return nil, err
	}
	return NewIncomeStatement(transactions), nil
}

func (s *service) PayTaxes(ctx context.Context, start, end time.Time, companyId int64) (int64, error) {
	statement, err := s.GetIncomeStatement(ctx, start, end, companyId)
	if err != nil {
		return -1, err
	}

	taxes := statement.GetTaxes()
	s.repository.SaveTaxes(ctx, taxes, companyId)

	return taxes, nil
}
