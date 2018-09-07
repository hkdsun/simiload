package main

import (
	"net/http"
	"sync"
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

	// evaluationWindow := 10 * time.Second

	// loadControlStrategy := "none"
	loadControlStrategy := "pro_queueing"
	// loadControlStrategy := "pro_num_workers"
	// loadControlStrategy := "p1"

	var accessController platform.AccessController

	if loadControlStrategy == "none" {
		accessController = &platform.DummyController{}
	} else if loadControlStrategy == "p1" {
		controller := &platform.ActiveController{}
		analyzer := &platform.P1Controller{
			QueueingTimeThreshold: 50 * time.Millisecond,
			CircuitTimeout:        30 * time.Second,
			AccessController:      accessController,
			StatsEvaluator:        platform.NewSlidingWindowRequestCounter(60 * time.Second),
			ActiveThrottlers:      make(map[platform.Scope]*platform.Throttler),
			ThrottleStrategy:      "global",
			// ThrottleStrategy:      "top_hitter",
		}
		controller.Analyzer = analyzer
		accessController = controller
	} else if loadControlStrategy == "pro_queueing" {
		controller := &platform.ActiveController{}
		analyzer := &platform.ProShed{
			SoftLimit:        10, // queueing time
			HardLimit:        50,
			Steps:            10,
			AccessController: accessController,
			LoadMut:          &sync.Mutex{},
			LoadStrategy:     "queueing",
		}
		controller.Analyzer = analyzer
		accessController = controller
	} else if loadControlStrategy == "pro_num_workers" {
		controller := &platform.ActiveController{}
		analyzer := &platform.ProShed{
			SoftLimit:        90, // worker utilization
			HardLimit:        100,
			Steps:            10,
			AccessController: accessController,
			LoadMut:          &sync.Mutex{},
			LoadStrategy:     "num_working",
		}
		controller.Analyzer = analyzer
		accessController = controller
	}

	workerGroup := &platform.WorkerGroup{
		NumWorkers: 100,
		Handler:    platform.DelayedResponder{100 * time.Millisecond},
		MaxRPS:     20,
	}
	workerGroupWg := workerGroup.Run()
	defer workerGroupWg.Wait()

	sim := &platform.Simulation{
		WorkerGroup:          workerGroup,
		Port:                 8080,
		RequestSamplingDelay: 0 * time.Millisecond,
		AccessController:     accessController,
	}

	configureMetrics()
	sim.Run()
}

func configureMetrics() {
	promSink, err := prom.NewPrometheusSink()
	if err != nil {
		panic(err)
	}

	config := metrics.DefaultConfig("sim")
	config.EnableHostname = false
	metrics.NewGlobal(config, promSink)

	log.Info("Starting prometheus handler on port 8081")
	go http.ListenAndServe(":8081", prometheus.Handler())
}
