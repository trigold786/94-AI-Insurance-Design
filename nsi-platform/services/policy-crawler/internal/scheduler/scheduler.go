package scheduler

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type SourceLevel string

const (
	LevelHigh   SourceLevel = "HIGH"
	LevelMedium SourceLevel = "MEDIUM"
	LevelLow    SourceLevel = "LOW"
)

func IntervalForLevel(level SourceLevel) time.Duration {
	switch level {
	case LevelHigh:
		return 24 * time.Hour
	case LevelMedium:
		return 168 * time.Hour // 7 days
	case LevelLow:
		return 0 // on-demand only
	default:
		return 24 * time.Hour
	}
}

type ScheduledSource struct {
	SourceID      string        `json:"source_id"`
	SourceLevel   SourceLevel   `json:"source_level"`
	Interval      time.Duration `json:"interval"`
	LastCrawlTime time.Time     `json:"last_crawl_time"`
}

type Scheduler struct {
	sources map[string]*ScheduledSource
	taskFn  func(sourceID string)

	stopped atomic.Bool
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		sources: make(map[string]*ScheduledSource),
		stopCh:  make(chan struct{}),
	}
}

func (s *Scheduler) SetTask(fn func(sourceID string)) {
	s.taskFn = fn
}

func (s *Scheduler) AddSource(sourceID string, level SourceLevel, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if interval <= 0 {
		interval = IntervalForLevel(level)
	}

	s.sources[sourceID] = &ScheduledSource{
		SourceID:    sourceID,
		SourceLevel: level,
		Interval:    interval,
	}
	log.Printf("[scheduler] added source %s (level=%s, interval=%v)", sourceID, level, interval)
}

func (s *Scheduler) RemoveSource(sourceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sources, sourceID)
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
	log.Printf("[scheduler] started with %d sources", len(s.sources))
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

func (s *Scheduler) TriggerManual(sourceID string) {
	s.mu.Lock()
	src, ok := s.sources[sourceID]
	s.mu.Unlock()

	if !ok {
		log.Printf("[scheduler] unknown source %s for manual trigger", sourceID)
		return
	}

	now := time.Now()
	src.LastCrawlTime = now
	if s.taskFn != nil {
		go s.taskFn(sourceID)
	}
	log.Printf("[scheduler] manual trigger for %s", sourceID)
}

func (s *Scheduler) GetScheduledSources() []*ScheduledSource {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*ScheduledSource, 0, len(s.sources))
	for _, src := range s.sources {
		result = append(result, src)
	}
	return result
}

func (s *Scheduler) loop() {
	for {
		s.mu.Lock()
		minWait := 24 * time.Hour
		now := time.Now()
		for _, src := range s.sources {
			if src.Interval <= 0 {
				continue
			}
			next := src.LastCrawlTime.Add(src.Interval)
			if src.LastCrawlTime.IsZero() || now.After(next) || now.Equal(next) {
				minWait = 0
				break
			}
			wait := next.Sub(now)
			if wait < minWait {
				minWait = wait
			}
		}
		s.mu.Unlock()

		timer := time.NewTimer(minWait)
		select {
		case <-timer.C:
			s.crawlDue()
		case <-s.stopCh:
			timer.Stop()
			return
		}
	}
}

func (s *Scheduler) crawlDue() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, src := range s.sources {
		if src.Interval <= 0 {
			continue
		}
		next := src.LastCrawlTime.Add(src.Interval)
		if src.LastCrawlTime.IsZero() || now.After(next) || now.Equal(next) {
			src.LastCrawlTime = now
			if s.taskFn != nil {
				go s.taskFn(src.SourceID)
			}
		}
	}
}
