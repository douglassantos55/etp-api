package accounting

import (
	"api/database"
	"context"
	"time"
)

type fakeRepository struct {
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

	return &fakeRepository{transactions}
}

func (r *fakeRepository) GetTransactions(ctx context.Context, companyId int64) ([]*Transaction, error) {
	return nil, nil
}

func (r *fakeRepository) SaveTaxes(ctx context.Context, taxes, companyId int64) error {
	if transactions, ok := r.transactions[companyId]; ok {
		description := "Taxes"
		classification := TAXES_PAID
		if taxes < 0 {
			description = "Deferred taxes"
			classification = TAXES_DEFERRED
		}
		transactions = append(transactions, &Transaction{
			Classification: uint64(classification),
			Description:    description,
			Value:          -int(taxes),
		})
	}
	return nil
}

func (r *fakeRepository) RegisterTransaction(tx *database.DB, transaction Transaction, companyId uint64) (int64, error) {
	if transactions, ok := r.transactions[int64(companyId)]; ok {
		transactions = append(transactions, &transaction)
	}
	return 0, nil
}

func (r *fakeRepository) GetIncomeTransactions(ctx context.Context, start, end time.Time, companyId int64) ([]*Transaction, error) {
	incomeTransactions := make([]*Transaction, 0)
	if transactions, ok := r.transactions[companyId]; ok {
		for _, transaction := range transactions {
			for _, classification := range INCOME_STATEMENT_CLASSIFICATIONS {
				if transaction.Classification == uint64(classification) {
					incomeTransactions = append(incomeTransactions, transaction)
				}
			}
		}
	}
	return incomeTransactions, nil
}
