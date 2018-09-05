package platform

import (
	"container/ring"
	"time"
)

// Stats evaluation interface
type Tracker interface {
	Add(Scope, time.Duration)
	Max(k int) []Scope
}

// Simple sum tracker:
// Ring buffer with size <limit> keeps track of processing time
// of each scope
type ProcessingTimeSumTracker struct {
	trackers map[Scope]*ring.Ring
	limit    int
}

func NewProcessingTimeSumTracker(limit int) *ProcessingTimeSumTracker {
	return &ProcessingTimeSumTracker{
		trackers: make(map[Scope]*ring.Ring),
		limit:    limit,
	}
}

func (u *ProcessingTimeSumTracker) Add(scope Scope, dur time.Duration) {
	_, ok := u.trackers[scope]
	if !ok {
		u.trackers[scope] = ring.New(u.limit)
	}

	u.trackers[scope].Value = dur
	u.trackers[scope] = u.trackers[scope].Next()
}

func (u *ProcessingTimeSumTracker) Max(n int) []Scope {
	var maxScope Scope
	var maxUsage time.Duration

	for scope, tracker := range u.trackers {
		usage := u.sum(tracker)
		if usage > maxUsage {
			maxScope = scope
			maxUsage = usage
		}
	}

	return []Scope{maxScope}
}

func (u *ProcessingTimeSumTracker) sum(tracker *ring.Ring) time.Duration {
	var sum time.Duration = 0
	tracker.Do(func(dur interface{}) {
		switch d := dur.(type) {
		case time.Duration:
			sum += d
		default:
		}
	})

	return sum
}

type SlidingWindowRequestCounter struct {
	*SlidingWindowCounter
}

func NewSlidingWindowRequestCounter(size time.Duration) *SlidingWindowRequestCounter {
	counter := NewSlidingWindowCounter(size, 1*time.Second)

	return &SlidingWindowRequestCounter{
		SlidingWindowCounter: counter,
	}
}

func (s *SlidingWindowRequestCounter) Add(scope Scope, dur time.Duration) {
	s.SlidingWindowCounter.Add(scope, 1) // Simple count of requests
}
