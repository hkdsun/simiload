package platform

import (
	"net/http"
	"time"
)

type DelayedResponder struct {
	ResponseTime time.Duration
}

func (s DelayedResponder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(s.ResponseTime)
}
