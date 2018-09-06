package platform

import (
	"context"
	"net/http"
	"sync"
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

	w.workQueue = make(WorkQueue, 1000)

	for id := 0; id < w.NumWorkers; id++ {
		go w.consumeWorkQueue(w.workQueue)
	}

	return wg
}

func (w *WorkerGroup) consumeWorkQueue(queue WorkQueue) {
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

		req := work.Request

		start := time.Now()
		w.Handler.ServeHTTP(req.httpResp, req.httpReq)
		req.ProcessingTime = time.Now().Sub(start)

		work.doneChan <- true
	}
}
