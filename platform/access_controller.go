package platform

import (
	"math/rand"
	"net/http"
)

type LoadAnalyzer interface {
	AnalyzeRequest(req *HttpRequest)
	AllowAccess(req *HttpRequest) bool
}

type AccessController interface {
	AllowAccess(req *HttpRequest) bool
	LogAccess(req *HttpRequest)
}

type DummyController struct{}

func (d *DummyController) AllowAccess(req *HttpRequest) bool {
	return true
}

func (d *DummyController) LogAccess(req *HttpRequest) {
	if rand.Float64() < 1.0 {
		// fmt.Printf("req = %+v\n", req.RequestStats)
	}
}

type ActiveController struct {
	Analyzer LoadAnalyzer
}

func (d *ActiveController) AllowAccess(req *HttpRequest) bool {
	return d.Analyzer.AllowAccess(req)
}

func (d *ActiveController) LogAccess(req *HttpRequest) {
	if d.Analyzer != nil && req.HttpStatus != http.StatusTooManyRequests {
		d.Analyzer.AnalyzeRequest(req)
	}
}
