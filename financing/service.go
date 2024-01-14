package financing

import (
	"api/accounting"
	"api/notification"
	"context"
	"log"
	"time"
)

type (
	Rates struct {
		Period    time.Time `db:"period" json:"period,omitempty"`
		Inflation float64   `db:"inflation" json:"inflation"`
		Interest  float64   `db:"interest" json:"interest"`
	}

	Service interface {
		CalculateRates(ctx context.Context) (*Rates, error)
		GetEffectiveRates(ctx context.Context) (*Rates, error)
		GetInflationPeriod(ctx context.Context, start, end time.Time) (float64, error)
		GetInterestPeriod(ctx context.Context, start, end time.Time, inflation float64) (float64, error)
	}

	service struct {
		repository Repository
		notifier   notification.Notifier
		logger     *log.Logger
	}
)

const Day = 24 * time.Hour

func NewService(repository Repository, notifier notification.Notifier, logger *log.Logger) Service {
	return &service{repository, notifier, logger}
}

func (s *service) GetEffectiveRates(ctx context.Context) (*Rates, error) {
	return s.repository.GetEffectiveRates(ctx)
}

func (s *service) CalculateRates(ctx context.Context) (*Rates, error) {
	start, end := accounting.GetCurrentPeriod()

	inflation, err := s.GetInflationPeriod(ctx, start, end)
	if err != nil {
		return nil, err
	}

	interest, err := s.GetInterestPeriod(ctx, start, end, inflation)
	if err != nil {
		return nil, err
	}

	rates := &Rates{
		Period:    end,
		Inflation: inflation,
		Interest:  interest,
	}

	if err := s.repository.SaveRates(ctx, end, rates); err != nil {
		return nil, err
	}

	event := &notification.Event{
		Type:    notification.FinancingRatesUpdated,
		Payload: rates,
	}

	if err := s.notifier.Broadcast(ctx, event); err != nil {
		s.logger.Printf("error broadcasting rates updated event: %s", err)
	}

	return rates, err
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
