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
