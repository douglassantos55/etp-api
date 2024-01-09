package notification

import (
	"context"
	"encoding/json"
	"io"
	"sync"
)

type (
	Notifier interface {
		Disconnect(identifier int64)
		Connect(identifier int64, client io.WriteCloser)
		Notify(ctx context.Context, message string, indentifier int64) error
	}

	Notification struct {
		Id        int64  `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Message   string `db:"message" json:"message"`
		CompanyId int64  `db:"company_id" json:"-"`
	}

	notifier struct {
		clients    *sync.Map
		repository Repository
	}
)

func NewNotifier(repository Repository) Notifier {
	notifier := &notifier{
		clients:    &sync.Map{},
		repository: repository,
	}

	return notifier
}

func (n *notifier) Connect(identifier int64, client io.WriteCloser) {
	n.clients.Store(identifier, client)
}

func (n *notifier) Disconnect(identifier int64) {
	n.clients.Delete(identifier)
}

func (n *notifier) Notify(ctx context.Context, message string, companyId int64) error {
	notification, err := n.repository.SaveNotification(ctx, &Notification{
		Message:   message,
		CompanyId: companyId,
	})

	if err != nil {
		return err
	}

	if client, ok := n.clients.Load(companyId); ok {
		stream, err := json.Marshal(notification)
		if err != nil {
			return err
		}
		if _, err := client.(io.Writer).Write(stream); err != nil {
			return err
		}
	}

	return nil
}
