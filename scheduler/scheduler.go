package scheduler

import (
	"log"
	"sync"
	"time"
)

type (
	Scheduler struct {
		retries *sync.Map
		timers  *sync.Map
		tickers *sync.Map
	}
)

func NewScheduler() *Scheduler {
	return &Scheduler{
		retries: &sync.Map{},
		timers:  &sync.Map{},
		tickers: &sync.Map{},
	}
}

func (s *Scheduler) Add(id any, duration time.Duration, callback func() error) {
	s.timers.Store(id, time.AfterFunc(duration, func() {
		s.timers.Delete(id)

		if err := callback(); err != nil {
			log.Println("error, retrying")

			s.retries.Store(id, time.AfterFunc(time.Second, func() {
				s.retries.Delete(id)

				if err := callback(); err != nil {
					log.Printf("could not run callback: %s", id)
				}
			}))
		}
	}))
}

func (s *Scheduler) Remove(id any) {
	if retry, found := s.retries.LoadAndDelete(id); found {
		timer := retry.(*time.Timer)
		if !timer.Stop() {
			<-timer.C
		}
	}

	if timer, found := s.timers.LoadAndDelete(id); found {
		if !timer.(*time.Timer).Stop() {
			<-timer.(*time.Timer).C
		}
	}

	if ticker, found := s.tickers.LoadAndDelete(id); found {
		ticker.(*time.Ticker).Stop()
	}
}

func (s *Scheduler) Repeat(id any, duration time.Duration, callback func() error) {
	ticker := time.NewTicker(duration)
	s.tickers.Store(id, ticker)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := callback(); err != nil {
					s.Remove(id)
				}
			}
		}
	}()
}
