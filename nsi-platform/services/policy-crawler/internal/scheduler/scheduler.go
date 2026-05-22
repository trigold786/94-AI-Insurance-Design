package scheduler

import (
	"sync"
	"sync/atomic"
	"time"
)

type Scheduler struct {
	Interval time.Duration
	Task     func()

	stopped atomic.Bool
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{
		Interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true
	s.stopped.Store(false)
	s.stopCh = make(chan struct{})

	go s.loop()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.stopped.Store(true)
	close(s.stopCh)
}

func (s *Scheduler) loop() {
	if s.Interval <= 0 {
		return
	}

	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.Task != nil {
				s.Task()
			}
		case <-s.stopCh:
			return
		}
	}
}
