package financing

import (
	"context"
	"time"
)

type fakeRepository struct {
	prices map[time.Time]map[int64]int64
}

func NewFakeRepository() Repository {
	october, _ := time.Parse(time.DateTime, "2023-10-01 00:00:00")
	november, _ := time.Parse(time.DateTime, "2023-11-26 00:00:00")
	december, _ := time.Parse(time.DateTime, "2023-12-03 00:00:00")

	prices := map[time.Time]map[int64]int64{
		october: {
			1: 1153,
			2: 3397,
		},
		november: {
			1: 1000,
		},
		december: {
			1: 1250,
			2: 3350,
		},
	}

	return &fakeRepository{prices}
}

func (r *fakeRepository) GetEffectiveRates(ctx context.Context) (*Rates, error) {
	return &Rates{Inflation: 0.125, Interest: 0.136}, nil
}

func (r *fakeRepository) GetAveragePrices(ctx context.Context, start, end time.Time) (map[int64]int64, error) {
	return r.prices[start], nil
}

func (r *fakeRepository) GetAverageInterestRate(ctx context.Context, start, end time.Time) (float64, error) {
	return 0.0165, nil
}

func (r *fakeRepository) SaveRates(ctx context.Context, period time.Time, rates *Rates) error {
	return nil
}
