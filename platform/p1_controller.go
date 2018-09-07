package platform

import (
	"fmt"
	"sync"
	"time"

	metrics "github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"
)

type P1Controller struct {
	QueueingTimeThreshold time.Duration
	CircuitTimeout        time.Duration
	AccessController      AccessController
	StatsEvaluator        Tracker
	ThrottleStrategy      string

	unhealthy       bool
	unhealthyTime   time.Time
	queueingTimeAvg time.Duration

	ActiveThrottlers map[Scope]*Throttler
	GlobalThrottler  *Throttler

	throttlersMut sync.RWMutex
}

func (c *P1Controller) AnalyzeRequest(req *HttpRequest) {
	// TODO: instrument load
	c.evaluateScopeUsage(req)
	c.evaluatePlatformHealth(req)
}
func (c *P1Controller) AllowAccess(req *HttpRequest) bool {
	// TODO: instrument access
	c.throttlersMut.RLock()
	defer c.throttlersMut.RUnlock()

	if c.GlobalThrottler != nil {
		return c.GlobalThrottler.Allow()
	}

	if len(c.ActiveThrottlers) > 0 {
		for _, scope := range RequestScopes(req) {
			throttler, ok := c.ActiveThrottlers[scope]
			if !ok {
				continue
			}

			if !throttler.Allow() {
				return false
			}
		}
	}

	return true
}

func (c *P1Controller) activateThrottler(throttler *Throttler) {
	c.throttlersMut.Lock()
	defer c.throttlersMut.Unlock()

	c.ActiveThrottlers[throttler.Scope] = throttler
}

func (c *P1Controller) clearThrottlers() {
	c.throttlersMut.Lock()
	defer c.throttlersMut.Unlock()

	c.ActiveThrottlers = make(map[Scope]*Throttler)
	c.GlobalThrottler = nil
}

func RequestScopes(req *HttpRequest) []Scope {
	return []Scope{Scope{ShopId: req.ShopId}}
}

func (c *P1Controller) evaluateScopeUsage(req *HttpRequest) {
	for _, scope := range RequestScopes(req) {
		c.StatsEvaluator.Add(scope, req.ProcessingTime)
	}
}

func (c *P1Controller) evaluatePlatformHealth(req *HttpRequest) {
	c.queueingTimeAvg -= c.queueingTimeAvg / 100
	c.queueingTimeAvg += req.QueueingTime / 100

	metrics.SetGauge([]string{"measured_load"}, float32(c.queueingTimeAvg.Seconds()))

	if c.queueingTimeAvg > c.QueueingTimeThreshold {
		c.triggerUnhealthy()
	} else {
		if c.unhealthy && time.Since(c.unhealthyTime) > c.CircuitTimeout {
			c.triggerHealthy()
		}
	}
}

func (c *P1Controller) triggerHealthy() {
	c.unhealthy = false
	c.unhealthyTime = time.Time{}
	c.clearThrottlers()
	log.Info("Recovered from high load")
}

func (c *P1Controller) triggerUnhealthy() {
	// TODO: use events?
	if c.unhealthy {
		return
	}

	c.unhealthy = true
	c.unhealthyTime = time.Now()

	switch c.ThrottleStrategy {
	case "global":
		c.GlobalThrottler = &Throttler{
			Rate: 0.5,
		}
	case "top_hitter":
		c.activateMaxScopeThrottler()
	default:
		panic(fmt.Sprintf("throttler %s not recognized", c.ThrottleStrategy))
	}
}

func (c *P1Controller) activateMaxScopeThrottler() {
	maxScope := c.StatsEvaluator.Max(1)[0]
	log.WithField("scope", maxScope).Warn("Banning scope due to high load")

	throttler := &Throttler{
		Scope: maxScope,
		Rate:  1.0,
	}

	c.activateThrottler(throttler)
}
