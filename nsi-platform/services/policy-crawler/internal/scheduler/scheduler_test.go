package scheduler

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler()
	if s == nil {
		t.Fatal("NewScheduler() returned nil")
	}
}

func TestSchedulerAddSource(t *testing.T) {
	s := NewScheduler()
	s.AddSource("test-1", LevelHigh, 0)
	s.AddSource("test-2", LevelMedium, 0)
	s.AddSource("test-3", LevelLow, 0)

	sources := s.GetScheduledSources()
	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d", len(sources))
	}
}

func TestSchedulerIntervalForLevel(t *testing.T) {
	if got := IntervalForLevel(LevelHigh); got != 24*time.Hour {
		t.Errorf("HIGH should be 24h, got %v", got)
	}
	if got := IntervalForLevel(LevelMedium); got != 168*time.Hour {
		t.Errorf("MEDIUM should be 168h, got %v", got)
	}
	if got := IntervalForLevel(LevelLow); got != 0 {
		t.Errorf("LOW should be 0, got %v", got)
	}
}

func TestSchedulerLOWSourceNotAutoCrawled(t *testing.T) {
	var callCount int32
	s := NewScheduler()
	s.SetTask(func(sourceID string) {
		atomic.AddInt32(&callCount, 1)
	})

	s.AddSource("low-src", LevelLow, 0)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	count := atomic.LoadInt32(&callCount)
	if count != 0 {
		t.Errorf("LOW source should not be auto-crawled, got %d calls", count)
	}
}

func TestSchedulerManualTrigger(t *testing.T) {
	var callCount int32
	s := NewScheduler()
	s.SetTask(func(sourceID string) {
		atomic.AddInt32(&callCount, 1)
	})

	s.AddSource("test-src", LevelLow, 0)
	s.TriggerManual("test-src")
	time.Sleep(20 * time.Millisecond)

	count := atomic.LoadInt32(&callCount)
	if count != 1 {
		t.Errorf("manual trigger should invoke task once, got %d", count)
	}
}

func TestSchedulerStartStop(t *testing.T) {
	var callCount int32
	s := NewScheduler()
	s.SetTask(func(sourceID string) {
		atomic.AddInt32(&callCount, 1)
	})

	s.AddSource("test-src", LevelHigh, 10*time.Millisecond)
	s.Start()
	time.Sleep(25 * time.Millisecond)
	s.Stop()

	count := atomic.LoadInt32(&callCount)
	if count < 2 {
		t.Errorf("expected at least 2 calls (25ms / 10ms interval), got %d", count)
	}
}

func TestSchedulerStopBeforeStart(t *testing.T) {
	s := NewScheduler()
	s.Stop() // should not panic
}

func TestSchedulerDoubleStart(t *testing.T) {
	var callCount int32
	s := NewScheduler()
	s.SetTask(func(sourceID string) {
		atomic.AddInt32(&callCount, 1)
	})

	s.AddSource("test-src", LevelHigh, 10*time.Millisecond)
	s.Start()
	s.Start() // second start should not create duplicate goroutine
	time.Sleep(20 * time.Millisecond)
	s.Stop()
}

func TestSchedulerRemoveSource(t *testing.T) {
	s := NewScheduler()
	s.AddSource("remove-me", LevelHigh, 10*time.Millisecond)
	s.RemoveSource("remove-me")

	sources := s.GetScheduledSources()
	if len(sources) != 0 {
		t.Errorf("expected 0 sources after removal, got %d", len(sources))
	}
}
