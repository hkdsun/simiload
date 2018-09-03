package platform

import "time"

type Controller struct {
	ScopeUsage                    map[Scope]time.Duration
	RunningScopes                 map[Scope]struct{}
	OverloadQueueingTimeThreshold time.Duration
	LoadControl                   LoadRegulator

	QueueingTimeAvg time.Duration
}

func (c *Controller) AnalyzeRequest(req *HttpRequest) {
	c.evaluateScopeUsage(req)
	c.evaluatePlatformHealth(req)
}

func (c *Controller) evaluateScopeUsage(req *HttpRequest) {
	for _, scope := range RequestScopes(req) {
		_, ok := c.RunningScopes[scope]
		if !ok {
			c.RunningScopes[scope] = struct{}{}
		}

		c.ScopeUsage[scope] += req.ProcessingTime
	}
}

func (c *Controller) evaluatePlatformHealth(req *HttpRequest) {
	c.QueueingTimeAvg -= c.QueueingTimeAvg / 100
	c.QueueingTimeAvg += req.QueueingTime / 30

	if c.QueueingTimeAvg > c.OverloadQueueingTimeThreshold {
		c.triggerUnhealthy()
	}
}

func (c *Controller) triggerUnhealthy() {
}
