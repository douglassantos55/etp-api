package notification_test

import (
	"api/notification"
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

type socket struct {
	mutex  sync.Mutex
	buffer *notification.Notification
}

func (s *socket) Flush() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.buffer == nil {
		return "empty"
	}

	message := s.buffer.Message
	s.buffer = nil

	return message
}

func (s *socket) Write(p []byte) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

	s1, s2, s3 := &socket{}, &socket{}, &socket{}
	notifier.Connect(1, s1)
	notifier.Connect(2, s2)

	t.Run("notify", func(t *testing.T) {
		notifier.Notify(context.TODO(), "hi there", 1)
		notifier.Notify(context.TODO(), "hello there", 2)

		time.Sleep(50 * time.Millisecond)

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

		time.Sleep(50 * time.Millisecond)

		if s1.Flush() != "empty" {
			t.Errorf("expected message %s, got %s", "empty", s1.Flush())
		}
	})

	t.Run("broadcast", func(t *testing.T) {
		notifier.Connect(1, s1)
		notifier.Connect(3, s3)

		notifier.Broadcast(context.TODO(), "broadcasting")

		time.Sleep(50 * time.Millisecond)

		if s1.Flush() != "broadcasting" {
			t.Errorf("expected message %s, got %s", "broadcasting", s1.Flush())
		}

		if s2.Flush() != "broadcasting" {
			t.Errorf("expected message %s, got %s", "broadcasting", s2.Flush())
		}

		if s3.Flush() != "broadcasting" {
			t.Errorf("expected message %s, got %s", "broadcasting", s3.Flush())
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		go notifier.Disconnect(1)
		go notifier.Disconnect(2)
		go notifier.Disconnect(3)

		go notifier.Connect(1, s1)
		go notifier.Connect(2, s2)
		go notifier.Connect(3, s3)

		go notifier.Notify(context.TODO(), "uga buga", 1)
		go notifier.Notify(context.TODO(), "do something!", 2)
		go notifier.Notify(context.TODO(), "are you there?", 1)
		go notifier.Notify(context.TODO(), "just do it", 2)
		go notifier.Broadcast(context.TODO(), "hello everyone")
	})
}
