package platform

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	metrics "github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Simulates the edge of the platform.  It can accept tons of requests but
// ultimately its response throughput is bottlenecked by its WorkerGroup
// throughput
// TODO: rename to SimulationServer
type LB struct {
	WorkerGroup    *WorkerGroup
	Port           uint
	LoadRegulator LoadRegulator
	LoggingDelay   time.Duration

	logQueue ReqQueue
}

func (lb *LB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := &HttpRequest{
		httpReq:  r,
		httpResp: w,
	}

	if strings.HasPrefix(r.URL.Path, "/shop") {
		split := strings.Split(r.URL.Path, "/")
		shopId, err := strconv.Atoi(split[len(split)-1])
		if err != nil {
			log.WithError(err).Error("unable to parse shopid")
			w.WriteHeader(500)
			return
		}

		request.ShopId = shopId
	}

	if !lb.LoadRegulator.AllowAccess(request) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	lb.WorkerGroup.Serve(request)

	lb.emitRequestMetrics(request)

	go func() {
		time.Sleep(lb.LoggingDelay)
		lb.logQueue <- request
	}()
}

func (lb *LB) Run() {
	lb.logQueue = make(ReqQueue, 1000)

	loggerWg := lb.startRequestLogger(lb.logQueue)
	defer loggerWg.Wait()

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", lb.Port),
		Handler: lb,
	}

	log.Infof("Starting HTTP server on port %d", lb.Port)
	log.Fatal(s.ListenAndServe())
}

func (lb *LB) startRequestLogger(logQueue ReqQueue) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for {
			request, ok := <-logQueue
			if !ok {
				break
			}
			lb.LoadRegulator.LogAccess(request)
		}
	}()

	return wg
}

func (lb *LB) emitRequestMetrics(req *HttpRequest) {
	metrics.AddSample([]string{"request.processing_time"}, float32(req.ProcessingTime.Seconds()))
	metrics.AddSample([]string{"request.queueing_time"}, float32(req.QueueingTime.Seconds()))
	metrics.IncrCounter([]string{"request.count"}, 1)
}
