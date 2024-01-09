package notification

import (
	"context"
)

type (
	Service interface {
		GetNotifications(ctx context.Context, companyId int64) ([]*Notification, error)
	}

	service struct {
		repository Repository
	}
)

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetNotifications(ctx context.Context, companyId int64) ([]*Notification, error) {
	return s.repository.GetNotifications(ctx, companyId)
}
