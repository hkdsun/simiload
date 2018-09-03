package platform

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type ReqQueue chan *HttpRequest
type WorkQueue chan Work

type Work struct {
	Request  *HttpRequest
	doneChan chan bool
}

// Simulates a limited capacity pool of workers
type WorkerGroup struct {
	NumWorkers           int
	Handler              http.Handler
	MaxRequestsPerSecond int

	workerQueues []WorkQueue
}

func (w *WorkerGroup) Serve(req *HttpRequest) {
	startQueueing := time.Now()

	<-w.serveReq(req)

	req.TotalTime = time.Now().Sub(startQueueing)
	req.QueueingTime = req.TotalTime - req.ProcessingTime
}

func (w *WorkerGroup) serveReq(req *HttpRequest) chan bool {
	doneChan := make(chan bool)
	work := Work{req, doneChan}

	workerId := rand.Intn(w.NumWorkers)
	w.workerQueues[workerId] <- work

	return doneChan
}

func (w *WorkerGroup) Run() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(w.NumWorkers)

	w.workerQueues = make([]WorkQueue, w.NumWorkers)

	for id := 0; id < w.NumWorkers; id++ {
		queue := make(WorkQueue, 16)
		go w.consumeWorkQueue(queue)

		w.workerQueues[id] = queue
	}

	return wg
}

func (w *WorkerGroup) consumeWorkQueue(queue WorkQueue) {
	for {
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
