package platform

import (
	"net/http"
	"sync"
)

type LoadAnalyzer interface {
	AnalyzeRequest(req *HttpRequest)
}

type AccessController interface {
	AllowAccess(req *HttpRequest) bool
	LogAccess(req *HttpRequest)
	ActivateThrottler(throttler *Throttler)
	ClearThrottlers()
}

type DummyController struct{}

func (d *DummyController) AllowAccess(req *HttpRequest) bool {
	return true
}

func (d *DummyController) LogAccess(req *HttpRequest) {
	// fmt.Printf("req = %+v\n", req)
}

func (d *DummyController) ClearThrottlers() {
}

func (d *DummyController) ActivateThrottler(throttler *Throttler) {
}

type ActiveController struct {
	ActiveThrottlers map[Scope]*Throttler
	Analyzer         LoadAnalyzer

	throttlersMut sync.RWMutex
}

func (d *ActiveController) AllowAccess(req *HttpRequest) bool {
	d.throttlersMut.RLock()
	defer d.throttlersMut.RUnlock()

	if len(d.ActiveThrottlers) > 0 {
		for _, scope := range RequestScopes(req) {
			throttler, ok := d.ActiveThrottlers[scope]
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

func (d *ActiveController) LogAccess(req *HttpRequest) {
	if d.Analyzer != nil && req.HttpStatus != http.StatusTooManyRequests {
		d.Analyzer.AnalyzeRequest(req)
	}
}

func (d *ActiveController) ActivateThrottler(throttler *Throttler) {
	d.throttlersMut.Lock()
	defer d.throttlersMut.Unlock()

	d.ActiveThrottlers[throttler.Scope] = throttler
}

func (d *ActiveController) ClearThrottlers() {
	d.throttlersMut.Lock()
	defer d.throttlersMut.Unlock()

	d.ActiveThrottlers = make(map[Scope]*Throttler)
}

func RequestScopes(req *HttpRequest) []Scope {
	return []Scope{Scope{ShopId: req.ShopId}}
}
