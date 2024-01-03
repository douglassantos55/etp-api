package financing

import (
	"context"
	"time"
)

type (
	Rates struct {
		Inflation float64 `db:"inflation" json:"inflation"`
		Interest  float64 `db:"interest" json:"interest"`
	}

	Service interface {
		GetEffectiveRates(ctx context.Context) (*Rates, error)
		CalculateRates(ctx context.Context) (float64, float64, error)
		GetInflationPeriod(ctx context.Context, start, end time.Time) (float64, error)
		GetInterestPeriod(ctx context.Context, start, end time.Time, inflation float64) (float64, error)
	}

	service struct {
		repository Repository
	}
)

const Day = 24 * time.Hour

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetEffectiveRates(ctx context.Context) (*Rates, error) {
	return s.repository.GetEffectiveRates(ctx)
}

func (s *service) CalculateRates(ctx context.Context) (float64, float64, error) {
	start, end := s.getCurrentPeriod()

	inflation, err := s.GetInflationPeriod(ctx, start, end)
	if err != nil {
		return -1, -1, err
	}

	interest, err := s.GetInterestPeriod(ctx, start, end, inflation)
	if err != nil {
		return -1, -1, err
	}

	rates := &Rates{Inflation: inflation, Interest: interest}
	if err := s.repository.SaveRates(ctx, end, rates); err != nil {
		return -1, -1, err
	}

	return inflation, interest, err
}

func (s *service) getCurrentPeriod() (time.Time, time.Time) {
	now := time.Now().UTC()
	year, month, day := now.Date()

	start := time.Date(year, month, day-int(now.Weekday())-7, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month, day-int(now.Weekday())-1, 23, 59, 59, 0, time.UTC)

	return start, end
}

func (s *service) GetInflationPeriod(ctx context.Context, start, end time.Time) (float64, error) {
	currentPrices, err := s.repository.GetAveragePrices(ctx, start, end)
	if err != nil {
		return -1, err
	}

	previousPrices, err := s.repository.GetAveragePrices(ctx, start.Add(-7*Day), end.Add(-7*Day))
	if err != nil {
		return -1, err
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
		return 0.0, nil
	}

	return inflation / float64(len(categories)), nil
}

func (s *service) GetInterestPeriod(ctx context.Context, start, end time.Time, inflation float64) (float64, error) {
	averageRate, err := s.repository.GetAverageInterestRate(ctx, start, end)
	if err != nil {
		return -1, err
	}

	return (averageRate + inflation) / (1 + inflation), nil
}
