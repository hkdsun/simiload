package platform

import (
	"container/ring"
	"time"

	metrics "github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"
)

type UsageTracker struct {
	samples *ring.Ring
}

func NewUsageTracker(limit int) *UsageTracker {
	return &UsageTracker{
		samples: ring.New(limit),
	}
}

func (u *UsageTracker) Add(dur time.Duration) {
	u.samples.Value = dur
	u.samples = u.samples.Next()
}

func (u *UsageTracker) Sum() time.Duration {
	var sum time.Duration = 0
	u.samples.Do(func(dur interface{}) {
		switch d := dur.(type) {
		case time.Duration:
			sum += d
		default:
		}
	})

	return sum
}

type OverloadController struct {
	OverloadQueueingTimeThreshold time.Duration
	CircuitTimeout                time.Duration
	Regulator                     LoadRegulator

	unhealthy       bool
	unhealthyTime   time.Time
	scopeUsage      map[Scope]*UsageTracker
	queueingTimeAvg time.Duration
}

func (c *OverloadController) Init() {
	c.scopeUsage = make(map[Scope]*UsageTracker)
}

func (c *OverloadController) AnalyzeRequest(req *HttpRequest) {
	c.evaluateScopeUsage(req)
	c.evaluatePlatformHealth(req)
}

func (c *OverloadController) evaluateScopeUsage(req *HttpRequest) {
	for _, scope := range RequestScopes(req) {
		_, ok := c.scopeUsage[scope]
		if !ok {
			c.scopeUsage[scope] = NewUsageTracker(1000)
		}

		c.scopeUsage[scope].Add(req.ProcessingTime)
	}
}

func (c *OverloadController) evaluatePlatformHealth(req *HttpRequest) {
	c.queueingTimeAvg -= c.queueingTimeAvg / 100
	c.queueingTimeAvg += req.QueueingTime / 30

	metrics.SetGauge([]string{"overload.queueing_time_avg"}, float32(c.queueingTimeAvg.Seconds()))

	if c.queueingTimeAvg > c.OverloadQueueingTimeThreshold {
		c.triggerUnhealthy()
	} else {
		if c.unhealthy && time.Since(c.unhealthyTime) > c.CircuitTimeout {
			c.triggerHealthy()
		}
	}
}

func (c *OverloadController) triggerHealthy() {
	c.unhealthy = false
	c.unhealthyTime = time.Time{}
	c.Regulator.ClearRegulators()
	log.Info("Recovered from high load")
}

func (c *OverloadController) triggerUnhealthy() {
	// TODO: use events?
	if c.unhealthy {
		return
	}

	c.unhealthy = true
	c.unhealthyTime = time.Now()

	maxScope := c.scopeWithMaxUsage()
	log.WithField("scope", maxScope).Warn("Banning scope due to high load")

	regulator := &Regulator{
		Scope: maxScope,
		Rate:  1.0,
	}

	c.Regulator.ActivateRegulator(regulator)
}

func (c *OverloadController) scopeWithMaxUsage() Scope {
	var maxScope Scope
	var maxUsage time.Duration

	for scope, tracker := range c.scopeUsage {
		usage := tracker.Sum()
		if usage > maxUsage {
			maxScope = scope
			maxUsage = usage
		}
	}

	return maxScope
}
