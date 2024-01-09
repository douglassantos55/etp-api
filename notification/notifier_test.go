package notification_test

import (
	"api/notification"
	"context"
	"encoding/json"
	"testing"
)

type socket struct {
	buffer *notification.Notification
}

func (s *socket) Flush() string {
	if s.buffer == nil {
		return "empty"
	}
	message := s.buffer.Message
	s.buffer = nil
	return message
}

func (s *socket) Write(p []byte) (int, error) {
	if err := json.Unmarshal(p, &s.buffer); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *socket) Close() error {
	return nil
}

func TestNotifier(t *testing.T) {
	notifier := notification.NewNotifier(notification.NewFakeRepository())

	s1, s2 := &socket{}, &socket{}
	notifier.Connect(1, s1)
	notifier.Connect(2, s2)

	t.Run("notify", func(t *testing.T) {
		notifier.Notify(context.TODO(), "hi there", 1)
		notifier.Notify(context.TODO(), "hello there", 2)

		if s1.Flush() != "hi there" {
			t.Errorf("expected message %s, got %s", "hi there", s1.Flush())
		}

		if s2.Flush() != "hello there" {
			t.Errorf("expected message %s, got %s", "hello there", s2.Flush())
		}
	})

	t.Run("disconnect", func(t *testing.T) {
		notifier.Disconnect(1)
		notifier.Notify(context.TODO(), "are you there?", 1)

		if s1.Flush() != "empty" {
			t.Errorf("expected message %s, got %s", "empty", s1.Flush())
		}
	})
}
