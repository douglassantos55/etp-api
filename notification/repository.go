package notification

import (
	"api/database"
	"context"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		GetNotifications(ctx context.Context, companyId int64) ([]*Notification, error)
		SaveNotification(ctx context.Context, notification *Notification) (*Notification, error)
	}

	goquRepository struct {
		builder *goqu.Database
	}

	inMemoryRepository struct {
		notifications map[int64][]*Notification
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder}
}

func (r *goquRepository) GetNotifications(ctx context.Context, companyId int64) ([]*Notification, error) {
	notifications := make([]*Notification, 0)

	err := r.builder.
		Select().
		From(goqu.T("notifications")).
		Where(goqu.I("company_id").Eq(companyId)).
		Limit(20).
		Order(goqu.I("created_at").Desc()).
		ScanStructsContext(ctx, &notifications)

	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *goquRepository) SaveNotification(ctx context.Context, notification *Notification) (*Notification, error) {
	result, err := r.builder.
		Insert(goqu.T("notifications")).
		Rows(notification).
		Executor().
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	notification.Id = id
	return notification, nil
}

func NewFakeRepository() Repository {
	return &inMemoryRepository{
		notifications: make(map[int64][]*Notification),
	}
}

func (r *inMemoryRepository) GetNotifications(ctx context.Context, companyId int64) ([]*Notification, error) {
	return r.notifications[companyId], nil
}

func (r *inMemoryRepository) SaveNotification(ctx context.Context, notification *Notification) (*Notification, error) {
	r.notifications[notification.CompanyId] = append(r.notifications[notification.CompanyId], notification)
	return notification, nil
}
