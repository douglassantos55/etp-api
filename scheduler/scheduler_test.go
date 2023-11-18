package scheduler_test

import (
	"api/scheduler"
	"errors"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		counter := make(chan int)
		scheduler := scheduler.NewScheduler()

		scheduler.Add(1, time.Second, func() error {
			counter <- 1
			return nil
		})

		count := <-counter
		if count != 1 {
			t.Errorf("expected count %d, got %d", 1, count)
		}
	})

	t.Run("Retry", func(t *testing.T) {
		counter := make(chan int)
		ready := make(chan bool, 2)
		scheduler := scheduler.NewScheduler()

		scheduler.Add(1, time.Second, func() error {
			select {
			case <-ready:
				counter <- 2
				return nil
			case <-time.After(time.Millisecond):
				ready <- true
				return errors.New("nope")
			}
		})

		count := <-counter
		if count != 2 {
			t.Errorf("expected count %d, got %d", 2, count)
		}
	})

	t.Run("Remove", func(t *testing.T) {
		counter := make(chan int)
		scheduler := scheduler.NewScheduler()

		scheduler.Add(1, time.Second, func() error {
			counter <- 1
			return nil
		})

		scheduler.Remove(1)

		select {
		case <-time.After(70 * time.Millisecond):
		case c := <-counter:
			t.Errorf("should not receive on channel, got: %d", c)
		}
	})
}
