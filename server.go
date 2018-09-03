package main

import (
	"net/http"
	"time"

	metrics "github.com/armon/go-metrics"
	prom "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/hkdsun/simiload/platform"
)

func main() {
	// args:
	// -num-workers [int]
	// -port [int]
	// -load-control [bool]
	// -feedback-delay [duration]

	enableLoadControl := true

	var loadRegulator platform.LoadRegulator = &platform.DummyRegulator{}
	if enableLoadControl {
		loadRegulator = &platform.OverloadRegulator{}
	}

	workerGroup := &platform.WorkerGroup{
		NumWorkers:           16,
		Handler:              platform.DelayedResponder{100 * time.Millisecond},
		MaxRequestsPerSecond: 0, // TODO: fancy capacity number
	}
	workerGroupWg := workerGroup.Run()
	defer workerGroupWg.Wait()

	lb := &platform.LB{
		WorkerGroup:   workerGroup,
		Port:          8080,
		LoggingDelay:  1 * time.Second,
		LoadRegulator: loadRegulator,
	}

	configureMetrics()
	lb.Run()
}

func configureMetrics() {
	promSink, err := prom.NewPrometheusSink()
	if err != nil {
		panic(err)
	}

	metrics.NewGlobal(metrics.DefaultConfig("sim"), promSink)

	log.Info("Starting prometheus handler on port 8081")
	go http.ListenAndServe(":8081", prometheus.Handler())
}
