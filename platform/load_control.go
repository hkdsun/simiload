package platform

import "fmt"

type LoadController interface {
	AllowAccess(req *HttpRequest) bool
	LogAccess(req *HttpRequest)
}

type DummyController struct{}

func (d *DummyController) AllowAccess(req *HttpRequest) bool {
	return true
}

func (d *DummyController) LogAccess(req *HttpRequest) {
	fmt.Printf("req = %+v\n", req)
}

type OverloadController struct{}

func (d *OverloadController) AllowAccess(req *HttpRequest) bool {
	return false
}

func (d *OverloadController) LogAccess(req *HttpRequest) {
}
