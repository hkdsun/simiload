package platform

import (
	"fmt"
	"sync"
)

type LoadRegulator interface {
	AllowAccess(req *HttpRequest) bool
	LogAccess(req *HttpRequest)
	ActivateRegulator(regulator *Regulator)
}

type DummyRegulator struct{}

func (d *DummyRegulator) AllowAccess(req *HttpRequest) bool {
	return true
}

func (d *DummyRegulator) LogAccess(req *HttpRequest) {
	fmt.Printf("req = %+v\n", req)
}

type OverloadRegulator struct {
	ActiveRegulators map[Scope]*Regulator

	regulatorsMut *sync.RWMutex
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
}

func (d *OverloadRegulator) ActivateRegulator(regulator *Regulator) {
	d.regulatorsMut.Lock()
	defer d.regulatorsMut.Unlock()

	d.ActiveRegulators[regulator.Scope] = regulator
}

func RequestScopes(req *HttpRequest) []Scope {
	return []Scope{Scope{ShopId: req.ShopId}}
}
