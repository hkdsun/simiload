package platform

import (
	"math/rand"
)

type Scope struct {
	ShopId int
}

type Regulator struct {
	Scope Scope
	Rate  float32
}

func (r *Regulator) Allow() bool {
	return rand.Float32() > r.Rate
}
