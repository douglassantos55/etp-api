package notification

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"
)

type EventType string

const (
	OrderPlaced    EventType = "order_placed"
	OrderPurchased           = "order_purchased"
	OrderCanceled            = "order_canceled"

	FinancingRatesUpdated = "financing_rates_updated"
)

type (
	Event struct {
		Type    EventType `json:"type"`
		Payload any       `json:"payload"`
	}

	Notifier interface {
		Disconnect(identifier int64)
		Connect(identifier int64, client io.WriteCloser)

		Broadcast(ctx context.Context, message any) error
		Notify(ctx context.Context, message string, indentifier int64) error
	}

	notifier struct {
		clients       *sync.Map
		repository    Repository
		notifications chan *Notification
		broadcasts    chan any
	}

	noOpNotifier struct {
	}
)

func NewNotifier(repository Repository) Notifier {
	notifier := &notifier{
		clients:       &sync.Map{},
		repository:    repository,
		notifications: make(chan *Notification),
		broadcasts:    make(chan any),
	}

	go notifier.handleMessages()

	return notifier
}

func (n *notifier) handleMessages() {
	for {
		select {
		case notification := <-n.notifications:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := n.doNotify(ctx, notification); err != nil {
				println(err.Error())
			}
		case message := <-n.broadcasts:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := n.doBroadcast(ctx, message); err != nil {
				println(err.Error())
			}
		}
	}
}

func (n *notifier) Connect(identifier int64, client io.WriteCloser) {
	n.clients.Store(identifier, client)
}

func (n *notifier) Disconnect(identifier int64) {
	n.clients.Delete(identifier)
}

func (n *notifier) Notify(ctx context.Context, message string, identifier int64) error {
	n.notifications <- &Notification{
		Message:   message,
		CompanyId: &identifier,
	}
	return nil
}

func (n *notifier) Broadcast(ctx context.Context, message any) error {
	n.broadcasts <- message
	return nil
}

func (n *notifier) doBroadcast(ctx context.Context, message any) error {
	stream, err := json.Marshal(message)
	if err != nil {
		return err
	}

	n.clients.Range(func(_, client any) bool {
		if _, err := client.(io.WriteCloser).Write(stream); err != nil {
			println(err.Error())
		}
		return true
	})

	return nil
}

func (n *notifier) doNotify(ctx context.Context, notification *Notification) error {
	notification, err := n.repository.SaveNotification(ctx, notification)

	if err != nil {
		return err
	}

	if client, ok := n.clients.Load(*notification.CompanyId); ok {
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

func NoOpNotifier() Notifier {
	return new(noOpNotifier)
}

func (n *noOpNotifier) Connect(identifier int64, client io.WriteCloser) {
}

func (n *noOpNotifier) Disconnect(identifier int64) {
}

func (n *noOpNotifier) Broadcast(ctx context.Context, message any) error {
	return nil
}

func (n *noOpNotifier) Notify(ctx context.Context, message string, identifier int64) error {
	return nil
}
