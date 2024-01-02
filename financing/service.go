package financing

import (
	"context"
	"time"
)

type (
	Service interface {
		GetInterestRate(ctx context.Context, start, end time.Time) (float64, error)
		GetInflation(ctx context.Context, start, end time.Time) (float64, map[int64]float64, error)
	}

	service struct {
		repository Repository
	}
)

const Day = 24 * time.Hour

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetInterestRate(ctx context.Context, start, end time.Time) (float64, error) {
	averageRate, err := s.repository.GetAverageInterestRate(ctx, start, end)
	if err != nil {
		return -1, err
	}

	inflation, _, err := s.GetInflation(ctx, start, end)
	if err != nil {
		return -1, err
	}

	return (averageRate + inflation) / (1 + inflation), nil
}

func (s *service) GetInflation(ctx context.Context, start, end time.Time) (float64, map[int64]float64, error) {
	currentPrices, err := s.repository.GetAveragePrices(ctx, start, end)
	if err != nil {
		return -1, nil, err
	}

	previousPrices, err := s.repository.GetAveragePrices(ctx, start.Add(-30*Day), end.Add(-30*Day))
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

	return inflation / float64(len(categories)), categories, nil
}
