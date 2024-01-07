package accounting

import (
	"api/database"
	"context"
	"errors"
	"sync"
	"time"
)

type fakeRepository struct {
	fails        int
	mutex        sync.Mutex
	transactions map[int64][]*Transaction
}

func NewFakeRepository() Repository {
	transactions := map[int64][]*Transaction{
		1: {
			{Classification: MARKET_SALE, Value: 100_000_00},
			{Classification: MARKET_FEE, Value: -37_000_00},
			{Classification: MARKET_PURCHASE, Value: -57_000_00},
			{Classification: TRANSPORT_FEE, Value: -7_000_00},
			{Classification: WAGES, Value: -15_000_00},
		},
		2: {
			{Classification: MARKET_SALE, Value: 527_000_00},
			{Classification: MARKET_FEE, Value: -137_000_00},
			{Classification: MARKET_PURCHASE, Value: -257_000_00},
			{Classification: TRANSPORT_FEE, Value: -37_000_00},
			{Classification: WAGES, Value: -75_000_00},
		},
		3: {
			{Classification: MARKET_SALE, Value: 527_000_00},
			{Classification: MARKET_FEE, Value: -137_000_00},
			{Classification: MARKET_PURCHASE, Value: -257_000_00},
			{Classification: TRANSPORT_FEE, Value: -37_000_00},
			{Classification: WAGES, Value: -75_000_00},
			{Classification: TAXES_DEFERRED, Value: 1_150_00},
		},
	}

	return &fakeRepository{fails: 0, transactions: transactions}
}

func (r *fakeRepository) GetPeriodResults(ctx context.Context, start, end time.Time) ([]*IncomeResult, error) {
	results := make([]*IncomeResult, 0)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for company, transactions := range r.transactions {
		var deferredTaxes int64
		var taxableIncome int64

		for _, transaction := range transactions {
			for _, classification := range INCOME_STATEMENT_CLASSIFICATIONS {
				if transaction.Classification == uint64(classification) {
					taxableIncome += int64(transaction.Value)
				}
			}
			if transaction.Classification == TAXES_DEFERRED {
				deferredTaxes += int64(transaction.Value)
			}
		}

		results = append(results, &IncomeResult{
			CompanyId:     company,
			DeferredTaxes: deferredTaxes,
			TaxableIncome: taxableIncome,
		})
	}

	return results, nil
}

func (r *fakeRepository) SaveTaxes(ctx context.Context, taxes, companyId int64) error {
	if companyId == 3 && r.fails == 0 {
		r.fails++
		return errors.New("bip bop bup")
	}

	r.mutex.Lock()

	if _, ok := r.transactions[companyId]; ok {
		description := "Taxes"
		classification := TAXES_PAID

		if taxes < 0 {
			description = "Deferred taxes"
			classification = TAXES_DEFERRED
		}

		r.mutex.Unlock()

		r.RegisterTransaction(nil, Transaction{
			Classification: uint64(classification),
			Description:    description,
			Value:          -int(taxes),
		}, uint64(companyId))
	} else {
		r.mutex.Unlock()
	}

	return nil
}

func (r *fakeRepository) RegisterTransaction(tx *database.DB, transaction Transaction, companyId uint64) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if transactions, ok := r.transactions[int64(companyId)]; ok {
		r.transactions[int64(companyId)] = append(transactions, &transaction)
	}

	return 0, nil
}

func (r *fakeRepository) GetIncomeTransactions(ctx context.Context, start, end time.Time, companyId int64) ([]*Transaction, error) {
	incomeTransactions := make([]*Transaction, 0)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if transactions, ok := r.transactions[companyId]; ok {
		for _, transaction := range transactions {
			for _, classification := range INCOME_STATEMENT_CLASSIFICATIONS {
				if transaction.Classification == uint64(classification) || transaction.Classification == TAXES_DEFERRED {
					incomeTransactions = append(incomeTransactions, transaction)
					break
				}
			}
		}
	}

	return incomeTransactions, nil
}
