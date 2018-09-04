package load

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/containous/traefik/log"
	"github.com/hkdsun/simiload/platform"
	"github.com/rakyll/hey/requester"
)

type Load struct {
	Scope       platform.Scope
	StartAfter  time.Duration
	Duration    time.Duration
	Concurrency int
	QPS         float64
}

type Generator struct {
	ServerURL string
	Loads     []*Load

	workMut     *sync.Mutex
	runningWork []*requester.Work
}

func (g *Generator) Run() {
	g.workMut = &sync.Mutex{}
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
	<-time.After(wait)
	g.ExecuteLoad(load)
}

func (g *Generator) ExecuteLoad(load Load) {
	log.WithField("scope", load.Scope).WithField("qps", load.QPS).WithField("concurrency", load.Concurrency).Infof("Starting load")

	path := fmt.Sprintf("%s/shop/%d", g.ServerURL, load.Scope.ShopId)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		panic(err)
	}

	var body []byte
	var proxyAddr *url.URL

	work := &requester.Work{
		Request:            req,
		RequestBody:        body,
		C:                  load.Concurrency, // num workers
		QPS:                load.QPS,
		N:                  math.MaxInt32,
		Timeout:            25,
		DisableCompression: false,
		DisableKeepAlives:  false,
		DisableRedirects:   false,
		H2:                 false,
		ProxyAddr:          proxyAddr,
		Output:             "",
	}
	work.Init()

	g.registerWork(work)

	if load.Duration > 0 {
		go func() {
			time.Sleep(load.Duration)
			work.Stop()
		}()
	}

	work.Run()
	log.WithField("scope", load.Scope).WithField("qps", load.QPS).WithField("concurrency", load.Concurrency).Infof("Finished load")
}

func (g *Generator) registerWork(work *requester.Work) {
	g.workMut.Lock()
	defer g.workMut.Unlock()

	g.runningWork = append(g.runningWork, work)
}
