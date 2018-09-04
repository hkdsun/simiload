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
type Simulation struct {
	WorkerGroup   *WorkerGroup
	Port          uint
	LoadRegulator LoadRegulator
	LoggingDelay  time.Duration

	logQueue ReqQueue
}

func (s *Simulation) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := &HttpRequest{
		httpReq:  r,
		httpResp: w,
	}

	defer func() {
		go func() {
			time.Sleep(s.LoggingDelay)
			s.logQueue <- request
		}()
	}()

	if strings.HasPrefix(r.URL.Path, "/shop") {
		split := strings.Split(r.URL.Path, "/")
		shopId, err := strconv.Atoi(split[len(split)-1])
		if err != nil {
			log.WithError(err).Error("unable to parse shopid")
			w.WriteHeader(500)
			request.HttpStatus = 500
			return
		}

		request.ShopId = shopId
	}

	if !s.LoadRegulator.AllowAccess(request) {
		w.WriteHeader(http.StatusTooManyRequests)
		request.HttpStatus = http.StatusTooManyRequests
		metrics.IncrCounterWithLabels([]string{"request.edge.dropped"}, 1, []metrics.Label{{"shop_id", fmt.Sprintf("%d", request.ShopId)}})
		return
	} else {
		metrics.IncrCounterWithLabels([]string{"request.edge.passed"}, 1, []metrics.Label{{"shop_id", fmt.Sprintf("%d", request.ShopId)}})
	}

	s.WorkerGroup.Serve(request)

	request.HttpStatus = 200

	s.emitRequestMetrics(request)
}

func (s *Simulation) Run() {
	s.logQueue = make(ReqQueue, 1000)

	loggerWg := s.startRequestLogger(s.logQueue)
	defer loggerWg.Wait()

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s,
	}

	log.Infof("Starting HTTP server on port %d", s.Port)
	log.Fatal(s.ListenAndServe())
}

func (s *Simulation) startRequestLogger(logQueue ReqQueue) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for {
			request, ok := <-logQueue
			if !ok {
				break
			}
			s.LoadRegulator.LogAccess(request)
		}
	}()

	return wg
}

func (s *Simulation) emitRequestMetrics(req *HttpRequest) {
	metrics.AddSample([]string{"request.processing_time"}, float32(req.ProcessingTime.Seconds()))
	metrics.AddSample([]string{"request.queueing_time"}, float32(req.QueueingTime.Seconds()))
	metrics.IncrCounterWithLabels([]string{"request.count"}, 1, []metrics.Label{{"status", string(req.HttpStatus)}})
}
