package platform

import (
	"time"

	metrics "github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"
)

type OverloadController struct {
	OverloadQueueingTimeThreshold time.Duration
	CircuitTimeout                time.Duration
	Regulator                     LoadRegulator

	unhealthy       bool
	unhealthyTime   time.Time
	scopeUsage      map[Scope]time.Duration
	queueingTimeAvg time.Duration
}

func (c *OverloadController) Init() {
	c.scopeUsage = make(map[Scope]time.Duration)
}

func (c *OverloadController) AnalyzeRequest(req *HttpRequest) {
	c.evaluateScopeUsage(req)
	c.evaluatePlatformHealth(req)
}

func (c *OverloadController) evaluateScopeUsage(req *HttpRequest) {
	for _, scope := range RequestScopes(req) {
		c.scopeUsage[scope] += req.ProcessingTime
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

	for scope, usage := range c.scopeUsage {
		if usage > maxUsage {
			maxScope = scope
			maxUsage = usage
		}
	}

	return maxScope
}
