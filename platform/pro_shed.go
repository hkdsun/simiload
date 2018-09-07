package platform

import (
	"sync"
	"time"

	metrics "github.com/armon/go-metrics"
)

// def addRequest(self, r):
//
//     HARD_QUOTA = 45
//     SOFT_QUOTA = 25
//     STEPS = 10
//
//     divisor = (HARD_QUOTA - SOFT_QUOTA) / STEPS
//
//     self.received += 1
//     self.req_modulus = (self.req_modulus + 1) % STEPS
//
//     # Are we overloaded?
//     load = self.getLoad()
//
//     # Become progressively more likely to reject requests
//     # once load > soft quota; reject everything once load
//     # hits hard limit.
//
//     threshold = int((HARD_QUOTA - load) / divisor)
//
//
//     if self.req_modulus < threshold:
//       # We're not too loaded
//       self.active_requests.append(r)
//       self.accepted += 1
//     else:
//       self.rejected += 1

type ProShed struct {
	SoftLimit        float64
	HardLimit        float64
	Steps            int
	AccessController AccessController
	LoadStrategy     string

	lastUpdate     time.Time
	queueingLoad   float64 // in milliseconds
	numWorkingLoad float64
	LoadMut        *sync.Mutex
	reqModulus     int
}

func (p *ProShed) AnalyzeRequest(req *HttpRequest) {
	p.updateLoad(req.QueueingTime.Seconds()*1000, req.NumWorking)
}

func (p *ProShed) AllowAccess(req *HttpRequest) bool {
	if time.Now().Sub(p.lastUpdate) >= 1*time.Second {
		return true
	}

	divisor := (p.HardLimit - p.SoftLimit) / float64(p.Steps)

	p.reqModulus = (p.reqModulus + 1) % p.Steps

	load := p.getLoad()
	threshold := int((float64(p.HardLimit) - load) / divisor)

	if p.reqModulus >= threshold {
		return false
	}

	return true
}

func (p *ProShed) updateLoad(queueingTime float64, numWorking uint32) {
	if time.Now().Sub(p.lastUpdate) <= 100*time.Millisecond {
		return
	}

	p.LoadMut.Lock()
	defer p.LoadMut.Unlock()

	p.queueingLoad -= p.queueingLoad / 30
	p.queueingLoad += float64(queueingTime) / 30

	p.numWorkingLoad -= p.numWorkingLoad / 30
	p.numWorkingLoad += float64(numWorking) / 30

	p.lastUpdate = time.Now()
	switch p.LoadStrategy {
	case "queueing":
		metrics.SetGauge([]string{"measured_load"}, float32(p.queueingLoad))
	case "num_working":
		metrics.SetGauge([]string{"measured_load"}, float32(p.numWorkingLoad))
	default:
		panic("no such laod strategy")
	}
}

func (p *ProShed) getLoad() float64 {
	p.LoadMut.Lock()
	defer p.LoadMut.Unlock()

	switch p.LoadStrategy {
	case "queueing":
		return p.queueingLoad
	case "num_working":
		return p.numWorkingLoad
	default:
		panic("no such laod strategy")
	}
}
