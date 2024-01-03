package financing

import (
	"api/accounting"
	"api/database"
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

const ACCUMULATED_PERIOD_WEEKS = 4

type (
	Repository interface {
		GetEffectiveRates(ctx context.Context) (*Rates, error)
		SaveRates(ctx context.Context, period time.Time, rates *Rates) error
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

func (r *goquRepository) GetEffectiveRates(ctx context.Context) (*Rates, error) {
	rates := new(Rates)

	_, err := r.builder.
		Select(
			goqu.COALESCE(goqu.SUM(goqu.I("inflation")), 0).As("inflation"),
			goqu.COALESCE(goqu.SUM(goqu.I("interest")), 0).As("interest"),
		).
		From(
			goqu.
				Select(goqu.I("inflation"), goqu.I("interest")).
				From(goqu.T("rates_history")).
				Order(goqu.I("period").Desc()).
				Limit(ACCUMULATED_PERIOD_WEEKS),
		).
		ScanStructContext(ctx, rates)

	if err != nil {
		return nil, err
	}

	return rates, nil
}

func (r *goquRepository) SaveRates(ctx context.Context, period time.Time, rates *Rates) error {
	_, err := r.builder.
		Insert(goqu.T("rates_history")).
		Rows(goqu.Record{
			"inflation": rates.Inflation,
			"interest":  rates.Interest,
			"period":    period,
		}).
		Executor().
		ExecContext(ctx)

	return err
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
