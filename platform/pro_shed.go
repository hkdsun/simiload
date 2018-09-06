package platform

type ProShed struct {
	SoftLimit        int
	HardLimit        int
	AccessController AccessController
}

func (p *ProShed) AnalyzeRequest(req *HttpRequest) {

}
