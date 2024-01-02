package financing

import (
	"context"
	"time"
)

type (
	Service interface {
		GetInterestRate(ctx context.Context) (float64, error)
		GetInflation(ctx context.Context) (float64, map[int64]float64, error)

		GetInterestPeriod(ctx context.Context, start, end time.Time) (float64, error)
		GetInflationPeriod(ctx context.Context, start, end time.Time) (float64, map[int64]float64, error)
	}

	service struct {
		repository Repository
	}
)

const Day = 24 * time.Hour

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetCurrentPeriod() (time.Time, time.Time) {
	now := time.Now().UTC()
	year, month, day := now.Date()

	start := time.Date(year, month, day-int(now.Weekday())-7, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month, day-int(now.Weekday())-1, 23, 59, 59, 0, time.UTC)

	return start, end
}

func (s *service) GetInterestRate(ctx context.Context) (float64, error) {
	start, end := s.GetCurrentPeriod()
	return s.GetInterestPeriod(ctx, start, end)
}

func (s *service) GetInflation(ctx context.Context) (float64, map[int64]float64, error) {
	start, end := s.GetCurrentPeriod()
	return s.GetInflationPeriod(ctx, start, end)
}

func (s *service) GetInterestPeriod(ctx context.Context, start, end time.Time) (float64, error) {
	averageRate, err := s.repository.GetAverageInterestRate(ctx, start, end)
	if err != nil {
		return -1, err
	}

	inflation, _, err := s.GetInflationPeriod(ctx, start, end)
	if err != nil {
		return -1, err
	}

	return (averageRate + inflation) / (1 + inflation), nil
}

func (s *service) GetInflationPeriod(ctx context.Context, start, end time.Time) (float64, map[int64]float64, error) {
	currentPrices, err := s.repository.GetAveragePrices(ctx, start, end)
	if err != nil {
		return -1, nil, err
	}

	previousPrices, err := s.repository.GetAveragePrices(ctx, start.Add(-7*Day), end.Add(-7*Day))
	if err != nil {
		return -1, nil, err
	}

	var inflation float64
	categories := make(map[int64]float64)
	for category, price := range currentPrices {
		if previousPrice, ok := previousPrices[category]; ok {
			categories[category] = float64(price)/float64(previousPrice) - 1
		} else {
			categories[category] = 0
		}
		inflation += categories[category]
	}

	if len(categories) == 0 {
		return 0.0, categories, nil
	}

	return inflation / float64(len(categories)), categories, nil
}
