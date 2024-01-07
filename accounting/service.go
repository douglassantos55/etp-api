package accounting

import (
	"api/scheduler"
	"context"
	"fmt"
	"time"
)

const TAX_RATE = 0.15

type (
	IncomeStatement struct {
		categories map[uint64]int
	}

	IncomeResult struct {
		CompanyId     int64 `db:"company_id"`
		TaxableIncome int64 `db:"taxable_income"`
		DeferredTaxes int64 `db:"deferred_taxes"`
	}

	Transaction struct {
		Value          int    `db:"value"`
		Description    string `db:"description"`
		Classification uint64 `db:"classification_id"`
	}

	Service interface {
		PayTaxes(ctx context.Context, start, end time.Time) error
	}

	service struct {
		repository Repository
		timer      *scheduler.Scheduler
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

func NewService(repository Repository, timer *scheduler.Scheduler) Service {
	return &service{repository, timer}
}

func GetCurrentPeriod() (start, end time.Time) {
	now := time.Now().UTC()
	year, month, day := now.Date()

	start = time.Date(year, month, day-int(now.Weekday())-7, 0, 0, 0, 0, time.UTC)
	end = time.Date(year, month, day-int(now.Weekday())-1, 23, 59, 59, 0, time.UTC)

	return start, end
}

func (s *service) GetIncomeStatement(ctx context.Context, start, end time.Time, companyId int64) (*IncomeStatement, error) {
	transactions, err := s.repository.GetIncomeTransactions(ctx, start, end, companyId)
	if err != nil {
		return nil, err
	}
	return NewIncomeStatement(transactions), nil
}

func (s *service) PayTaxes(ctx context.Context, start, end time.Time) error {
	results, err := s.repository.GetPeriodResults(ctx, start, end)
	if err != nil {
		return err
	}

	for _, result := range results {
		companyId := result.CompanyId
		taxes := int64(float64(result.TaxableIncome)*TAX_RATE) - result.DeferredTaxes

		if err := s.repository.SaveTaxes(ctx, taxes, result.CompanyId); err != nil {
			s.timer.Add(fmt.Sprintf("TAXES_%d", result.CompanyId), 3*time.Second, func() error {
				return s.repository.SaveTaxes(ctx, taxes, companyId)
			})
		}
	}

	return nil
}
