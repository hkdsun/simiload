package platform

import (
	"fmt"
	"sync"
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
	SoftLimit        int
	HardLimit        int
	Steps            int
	AccessController AccessController

	queueLength int
	QueueMut    *sync.Mutex
	reqModulus  int
}

func (p *ProShed) AnalyzeRequest(req *HttpRequest) {
	p.updateQueueLength(req.QueueLength)
}

func (p *ProShed) AllowAccess(req *HttpRequest) bool {
	p.QueueMut.Lock()
	defer p.QueueMut.Unlock()

	divisor := (float64(p.HardLimit) - float64(p.SoftLimit)) / float64(p.Steps)

	p.reqModulus = (p.reqModulus + 1) % p.Steps

	threshold := int((float64(p.HardLimit) - float64(p.queueLength)) / divisor)

	fmt.Printf("p.queueLength = %+v\n", p.queueLength)
	fmt.Printf("threshold = %+v\n", threshold)

	if p.reqModulus >= threshold {
		return false
	}

	return true
}

func (p *ProShed) updateQueueLength(queueLength int) {
	p.QueueMut.Lock()
	defer p.QueueMut.Unlock()
	p.queueLength = queueLength
}
