package platform

import (
	"math/rand"
)

type Scope struct {
	ShopId int
}

type Throttler struct {
	Scope Scope
	Rate  float32
}

func (r *Throttler) Allow() bool {
	return rand.Float32() > r.Rate
}

type ProThrottler struct {
	Steps      int
	reqModulus int
}

func (t *ProThrottler) Allow(soft, hard, load float64) bool {
	divisor := (hard - soft) / float64(t.Steps)
	threshold := int((float64(hard) - load) / divisor)

	t.reqModulus = (t.reqModulus + 1) % t.Steps
	return threshold > t.reqModulus
}
