package financing

import (
	"context"
	"time"
)

type fakeRepository struct {
	prices map[time.Time]map[int64]int64
}

func NewFakeRepository() Repository {
	october, _ := time.Parse("2006-01-02 15:04:05", "2023-10-01 00:00:00")
	november, _ := time.Parse("2006-01-02 15:04:05", "2023-11-01 00:00:00")
	december, _ := time.Parse("2006-01-02 15:04:05", "2023-12-01 00:00:00")

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

func (r *fakeRepository) GetAveragePrices(ctx context.Context, start, end time.Time) (map[int64]int64, error) {
	return r.prices[start], nil
}
