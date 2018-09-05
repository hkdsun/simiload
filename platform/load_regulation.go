package platform

import (
	"net/http"
	"sync"
)

type LoadRegulator interface {
	AllowAccess(req *HttpRequest) bool
	LogAccess(req *HttpRequest)
	ActivateRegulator(regulator *Regulator)
	ClearRegulators()
	AddAnalyzer(func(req *HttpRequest))
}

type DummyRegulator struct{}

func (d *DummyRegulator) AllowAccess(req *HttpRequest) bool {
	return true
}

func (d *DummyRegulator) LogAccess(req *HttpRequest) {
	// fmt.Printf("req = %+v\n", req)
}

func (d *DummyRegulator) ClearRegulators() {
}

func (d *DummyRegulator) ActivateRegulator(regulator *Regulator) {
}

func (d *DummyRegulator) AddAnalyzer(f func(req *HttpRequest)) {
}

type OverloadRegulator struct {
	ActiveRegulators map[Scope]*Regulator
	AnalyzerFunc     func(req *HttpRequest)

	regulatorsMut sync.RWMutex
}

func (d *OverloadRegulator) AllowAccess(req *HttpRequest) bool {
	d.regulatorsMut.RLock()
	defer d.regulatorsMut.RUnlock()

	if len(d.ActiveRegulators) > 0 {
		for _, scope := range RequestScopes(req) {
			regulator, ok := d.ActiveRegulators[scope]
			if !ok {
				continue
			}

			if !regulator.Allow() {
				return false
			}
		}
	}

	return true
}

func (d *OverloadRegulator) LogAccess(req *HttpRequest) {
	if d.AnalyzerFunc != nil && req.HttpStatus != http.StatusTooManyRequests {
		d.AnalyzerFunc(req)
	}
}

func (d *OverloadRegulator) ActivateRegulator(regulator *Regulator) {
	d.regulatorsMut.Lock()
	defer d.regulatorsMut.Unlock()

	d.ActiveRegulators[regulator.Scope] = regulator
}

func (d *OverloadRegulator) ClearRegulators() {
	d.regulatorsMut.Lock()
	defer d.regulatorsMut.Unlock()

	d.ActiveRegulators = make(map[Scope]*Regulator)
}

func RequestScopes(req *HttpRequest) []Scope {
	return []Scope{Scope{ShopId: req.ShopId}}
}

func (d *OverloadRegulator) AddAnalyzer(f func(req *HttpRequest)) {
	d.AnalyzerFunc = f
}
