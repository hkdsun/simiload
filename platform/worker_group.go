package platform

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	metrics "github.com/armon/go-metrics"
	"golang.org/x/time/rate"
)

type ReqQueue chan *HttpRequest
type WorkQueue chan Work

type Work struct {
	Request  *HttpRequest
	doneChan chan bool
}

// Simulates a limited capacity pool of workers
type WorkerGroup struct {
	NumWorkers int
	NumWorking uint32
	Handler    http.Handler
	MaxRPS     int

	workQueue WorkQueue
}

func (w *WorkerGroup) Serve(req *HttpRequest) {
	startQueueing := time.Now()

	<-w.serveReq(req)

	req.TotalTime = time.Now().Sub(startQueueing)
	req.QueueingTime = req.TotalTime - req.ProcessingTime
	req.QueueLength = len(w.workQueue)
	req.NumWorking = atomic.LoadUint32(&w.NumWorking)
}

func (w *WorkerGroup) serveReq(req *HttpRequest) chan bool {
	doneChan := make(chan bool)
	work := Work{req, doneChan}
	w.workQueue <- work
	return doneChan
}

func (w *WorkerGroup) Run() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(w.NumWorkers)

	w.NumWorking = 0
	w.workQueue = make(WorkQueue, 1000)

	for id := 0; id < w.NumWorkers; id++ {
		go w.consumeWorkQueue(w.workQueue, id)
	}

	go func() {
		for {
			<-time.After(1 * time.Second)
			metrics.SetGauge([]string{"workers.online"}, float32(w.NumWorkers))
			metrics.SetGauge([]string{"workers.utilized"}, float32(atomic.LoadUint32(&w.NumWorking)))
		}
	}()

	return wg
}

func (w *WorkerGroup) consumeWorkQueue(queue WorkQueue, id int) {
	limiter := rate.NewLimiter(rate.Limit(w.MaxRPS), 1)

	for {
		err := limiter.Wait(context.TODO())
		if err != nil {
			panic(err)
		}

		metrics.IncrCounter([]string{"worker.pass"}, 1)

		work, ok := <-queue
		if !ok {
			break
		}

		atomic.AddUint32(&w.NumWorking, 1)
		metrics.SetGaugeWithLabels([]string{"workers.working"}, 1, []metrics.Label{{"id", fmt.Sprintf("%d", id)}})

		req := work.Request

		start := time.Now()
		w.Handler.ServeHTTP(req.httpResp, req.httpReq)
		req.ProcessingTime = time.Now().Sub(start)

		atomic.AddUint32(&w.NumWorking, ^uint32(0))
		metrics.SetGaugeWithLabels([]string{"workers.working"}, 0, []metrics.Label{{"id", fmt.Sprintf("%d", id)}})

		work.doneChan <- true
	}
}
