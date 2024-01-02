package financing

import (
	"api/accounting"
	"api/database"
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type (
	Repository interface {
		GetAveragePrices(ctx context.Context, start, end time.Time) (map[int64]int64, error)
		GetAverageInterestRate(ctx context.Context, start, end time.Time) (float64, error)
	}

	goquRepository struct {
		builder *goqu.Database
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) GetAverageInterestRate(ctx context.Context, start, end time.Time) (float64, error) {
	var interestRate float64

	_, err := r.builder.
		Select(goqu.COALESCE(goqu.AVG(goqu.I("interest_rate")), 0.01)).
		From(goqu.T("loans")).
		Where(goqu.I("created_at").Between(exp.NewRangeVal(start, end))).
		ScanValContext(ctx, &interestRate)

	if err != nil {
		return -1, err
	}

	return interestRate, nil
}

func (r *goquRepository) GetAveragePrices(ctx context.Context, start, end time.Time) (map[int64]int64, error) {
	var averagePrices []struct {
		CategoryId   int64 `db:"category_id"`
		AveragePrice int64 `db:"average_price"`
	}

	err := r.builder.
		Select(
			goqu.I("r.category_id"),
			goqu.L("? / ?",
				goqu.SUM(goqu.I("t.value")),
				goqu.SUM(goqu.I("ot.quantity")),
			).As("average_price"),
		).
		From(goqu.T("transactions").As("t")).
		InnerJoin(
			goqu.T("orders_transactions").As("ot"),
			goqu.On(goqu.I("ot.transaction_id").Eq(goqu.I("t.id"))),
		).
		InnerJoin(
			goqu.T("orders").As("o"),
			goqu.On(goqu.I("ot.order_id").Eq(goqu.I("o.id"))),
		).
		InnerJoin(
			goqu.T("resources").As("r"),
			goqu.On(goqu.I("o.resource_id").Eq(goqu.I("r.id"))),
		).
		Where(goqu.And(
			goqu.I("t.classification_id").Eq(accounting.MARKET_SALE),
			goqu.I("t.created_at").Between(exp.NewRangeVal(start.UTC(), end.UTC())),
		)).
		GroupBy(goqu.I("r.category_id")).
		ScanStructsContext(ctx, &averagePrices)

	if err != nil {
		return nil, err
	}

	prices := make(map[int64]int64)
	for _, averagePrice := range averagePrices {
		prices[averagePrice.CategoryId] = averagePrice.AveragePrice
	}

	return prices, err
}
