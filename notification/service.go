package notification

import (
	"context"
)

type (
	Notification struct {
		Id        int64  `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Message   string `db:"message" json:"message"`
		CompanyId *int64 `db:"company_id" json:"-"`
	}

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
