package platform

import (
	"math"
	"time"
)

// TODO: LRU eviction
type Bucket struct {
	frequencies map[Scope]float64
	maxSize     int
}

func NewBucket(maxSize int) Bucket {
	return Bucket{
		frequencies: make(map[Scope]float64),
		maxSize:     maxSize,
	}
}

type SlidingWindowCounter struct {
	size        time.Duration
	granularity time.Duration
	numBuckets  int

	pos        int
	lastUpdate time.Time
	buckets    []Bucket
	summary    Bucket // TODO: keep sorted?
}

func NewSlidingWindowCounter(size, granularity time.Duration) *SlidingWindowCounter {
	c := &SlidingWindowCounter{
		size:        size,
		granularity: granularity,
		numBuckets:  int(size.Nanoseconds() / granularity.Nanoseconds()),
	}
	c.Clear()
	return c
}

func (s *SlidingWindowCounter) Add(scope Scope, value float64) {
	s.tick()
	s.addValue(scope, value)
}

// TODO: lulwat
func (s *SlidingWindowCounter) Max(n int) []Scope {
	var maxScope Scope
	var maxValue float64

	for scope, value := range s.summary.frequencies {
		if value > maxValue {
			maxScope = scope
			maxValue = value
		}
	}

	return []Scope{maxScope}
}

func (s *SlidingWindowCounter) Clear() {
	s.pos = 0
	s.buckets = make([]Bucket, s.numBuckets)
	for b := 0; b < s.numBuckets; b++ {
		s.buckets[b] = NewBucket(100)
	}
	s.summary = NewBucket(100)
}

func (s *SlidingWindowCounter) addValue(scope Scope, value float64) {
	// fmt.Printf("s.buckets[s.pos].frequencies[scope] = %+v\n", s.buckets[s.pos].frequencies)
	// fmt.Printf("s.summary.frequencies[scope] = %+v\n", s.summary.frequencies)
	// fmt.Printf("s.pos = %+v\n", s.pos)
	s.buckets[s.pos].frequencies[scope] += value
	s.summary.frequencies[scope] += value
}

func (s *SlidingWindowCounter) tick() {
	now := time.Now()

	elapsedTicks := int(now.Sub(s.lastUpdate).Nanoseconds() / s.granularity.Nanoseconds())
	elapsedTicks = int(math.Floor(float64(elapsedTicks)))
	if elapsedTicks < 1 {
		return
	}

	// fmt.Println("ticking forward")
	// fmt.Printf("s = %+v\n", s)

	s.lastUpdate = now

	if elapsedTicks >= s.numBuckets {
		s.Clear()
		return
	}

	for pos := 0; pos < elapsedTicks; pos++ {
		if s.pos += 1; s.pos >= s.numBuckets {
			s.pos = 0
		}
		s.replaceBucket(s.pos, NewBucket(100))
	}
}

func (s *SlidingWindowCounter) replaceBucket(pos int, bucket Bucket) {
	s.subtractFromSummary(s.buckets[pos])
	s.buckets[pos] = bucket
}

func (s *SlidingWindowCounter) subtractFromSummary(bucket Bucket) {
	for scope, value := range bucket.frequencies {
		s.summary.frequencies[scope] -= value
	}
}
