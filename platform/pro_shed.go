package platform

type ProShed struct {
	SoftLimit int
	HardLimit int
	Regulator LoadRegulator
}

func (p *ProShed) AnalyzeRequest(req *HttpRequest) {

}
