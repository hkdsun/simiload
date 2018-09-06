package platform

import (
	"time"

	metrics "github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"
)

type P1Controller struct {
	QueueingTimeThreshold time.Duration
	CircuitTimeout        time.Duration
	AccessController      AccessController
	StatsEvaluator        Tracker

	unhealthy       bool
	unhealthyTime   time.Time
	queueingTimeAvg time.Duration
}

func (c *P1Controller) AnalyzeRequest(req *HttpRequest) {
	c.evaluateScopeUsage(req)
	c.evaluatePlatformHealth(req)
}

func (c *P1Controller) evaluateScopeUsage(req *HttpRequest) {
	for _, scope := range RequestScopes(req) {
		c.StatsEvaluator.Add(scope, req.ProcessingTime)
	}
}

func (c *P1Controller) evaluatePlatformHealth(req *HttpRequest) {
	c.queueingTimeAvg -= c.queueingTimeAvg / 100
	c.queueingTimeAvg += req.QueueingTime / 30

	metrics.SetGauge([]string{"overload.queueing_time_avg"}, float32(c.queueingTimeAvg.Seconds()))

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
	c.AccessController.ClearRegulators()
	log.Info("Recovered from high load")
}

func (c *P1Controller) triggerUnhealthy() {
	// TODO: use events?
	if c.unhealthy {
		return
	}

	c.unhealthy = true
	c.unhealthyTime = time.Now()

	maxScope := c.StatsEvaluator.Max(1)[0]
	log.WithField("scope", maxScope).Warn("Banning scope due to high load")

	regulator := &Regulator{
		Scope: maxScope,
		Rate:  1.0,
	}

	c.AccessController.ActivateRegulator(regulator)
}
