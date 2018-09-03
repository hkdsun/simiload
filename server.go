package main

import (
	"time"

	"github.com/hkdsun/simiload/platform"
)

func main() {
	// args:
	// -num-workers [int]
	// -port [int]
	// -load-control [bool]
	// -feedback-delay [duration]

	enableLoadControl := false

	var loadController platform.LoadController = &platform.DummyController{}
	if enableLoadControl {
		loadController = &platform.OverloadController{}
	}

	workerGroup := &platform.WorkerGroup{
		NumWorkers:           8,
		Handler:              platform.DelayedResponder{500 * time.Millisecond},
		MaxRequestsPerSecond: 0, // TODO: fancy capacity number
	}
	workerGroupWg := workerGroup.Run()
	defer workerGroupWg.Wait()

	lb := &platform.LB{
		WorkerGroup:    workerGroup,
		Port:           8080,
		LoggingDelay:   1 * time.Second,
		LoadController: loadController,
	}

	lb.Run()
}
