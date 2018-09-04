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
	// -max-worker-rps [int]
	// -num-workers [int]
	// -feedback-delay [duration]
	// -load-control [bool]
	// -worker-response-time [duration]

	enableLoadControl := true

	var loadRegulator platform.LoadRegulator = &platform.DummyRegulator{}
	if enableLoadControl {
		loadRegulator = &platform.OverloadRegulator{
			ActiveRegulators: make(map[platform.Scope]*platform.Regulator),
		}

		loadController := platform.OverloadController{
			OverloadQueueingTimeThreshold: 100 * time.Millisecond,
			CircuitTimeout:                5 * time.Second,
			Regulator:                     loadRegulator,
		}

		loadController.Init()

		loadRegulator.AddAnalyzer(loadController.AnalyzeRequest)
	}

	workerGroup := &platform.WorkerGroup{
		NumWorkers: 10,
		Handler:    platform.DelayedResponder{100 * time.Millisecond},
		MaxRPS:     10,
	}
	workerGroupWg := workerGroup.Run()
	defer workerGroupWg.Wait()

	sim := &platform.Simulation{
		WorkerGroup:   workerGroup,
		Port:          8080,
		LoggingDelay:  10 * time.Second,
		LoadRegulator: loadRegulator,
	}

	configureMetrics()
	sim.Run()
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
