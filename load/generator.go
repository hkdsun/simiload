package load

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hkdsun/simiload/platform"
	"github.com/rakyll/hey/requester"
)

type Load struct {
	Scope      platform.Scope `json:"scope"`
	StartAfter time.Duration  `json:"start_after"`
	Duration   time.Duration  `json:"duration"`
}

type Generator struct {
	ServerURL string
	Loads     []*Load

	runningWork []*requester.Work
}

func (g *Generator) Run() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		g.Stop()
	}()

	wg := &sync.WaitGroup{}

	for _, l := range g.Loads {
		var load Load = *l
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.ExecuteLoadAfter(load, load.StartAfter)
		}()
	}

	wg.Wait()
}

func (g *Generator) Stop() {
	for _, work := range g.runningWork {
		work.Stop()
	}
}

func (g *Generator) ExecuteLoadAfter(load Load, wait time.Duration) {
	fmt.Printf("load = %+v\n", load)
	<-time.After(wait)
	g.ExecuteLoad(load)
}

func (g *Generator) ExecuteLoad(load Load) {
	fmt.Printf("load = %+v\n", load)
	// w := &requester.Work{
	// 	Request:            req,
	// 	RequestBody:        bodyAll,
	// 	C:                  conc,
	// 	QPS:                q,
	// 	Timeout:            *t,
	// 	DisableCompression: *disableCompression,
	// 	DisableKeepAlives:  *disableKeepAlives,
	// 	DisableRedirects:   *disableRedirects,
	// 	H2:                 *h2,
	// 	ProxyAddr:          proxyURL,
	// 	Output:             *output,
	// }
	// w.Init()
	// if dur > 0 {
	// 	go func() {
	// 		time.Sleep(dur)
	// 		w.Stop()
	// 	}()
	// }
	// w.Run()

}
