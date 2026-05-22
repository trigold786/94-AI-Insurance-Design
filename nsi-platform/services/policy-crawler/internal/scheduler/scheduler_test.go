package scheduler

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler(1 * time.Hour)
	if s == nil {
		t.Fatal("NewScheduler() returned nil")
	}
	if s.Interval != 1*time.Hour {
		t.Errorf("expected 1h interval, got %v", s.Interval)
	}
}

func TestSchedulerStartStop(t *testing.T) {
	var callCount int32
	s := NewScheduler(10 * time.Millisecond)
	s.Task = func() {
		atomic.AddInt32(&callCount, 1)
	}

	s.Start()
	time.Sleep(25 * time.Millisecond)
	s.Stop()

	count := atomic.LoadInt32(&callCount)
	if count < 2 {
		t.Errorf("expected at least 2 calls (25ms / 10ms), got %d", count)
	}
}

func TestSchedulerStopBeforeStart(t *testing.T) {
	s := NewScheduler(1 * time.Hour)
	s.Stop() // should not panic
}

func TestSchedulerDoubleStart(t *testing.T) {
	s := NewScheduler(10 * time.Millisecond)
	s.Task = func() {}

	s.Start()
	s.Start() // second start should not create duplicate goroutine
	time.Sleep(20 * time.Millisecond)
	s.Stop()
}

func TestSchedulerInterval(t *testing.T) {
	s := NewScheduler(50 * time.Millisecond)

	var count int32
	s.Task = func() {
		atomic.AddInt32(&count, 1)
	}

	s.Start()
	time.Sleep(110 * time.Millisecond)
	s.Stop()

	c := atomic.LoadInt32(&count)
	if c < 2 {
		t.Errorf("expected at least 2 invocations in 110ms (50ms interval), got %d", c)
	}
}

func TestSchedulerZeroInterval(t *testing.T) {
	s := NewScheduler(0)
	if s == nil {
		t.Fatal("NewScheduler() returned nil")
	}
}
