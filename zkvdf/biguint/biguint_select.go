package biguint

import (
	"github.com/consensys/gnark/frontend"
)

func (api bigUintAPI) Select(selector frontend.Variable, a, b BigUint) BigUint {
	out := New(max(len(a.Limbs), len(b.Limbs)))
	for i := range out.Limbs {
		out.Limbs[i] = api.core.Select(selector, a.GetLimb(i), b.GetLimb(i))
	}
	return out
}

func (api bigUintAPI) Swap(selector frontend.Variable, a, b BigUint) (BigUint, BigUint) {
	x := api.Select(selector, a, b)
	y := api.Select(selector, b, a)
	return x, y
}
