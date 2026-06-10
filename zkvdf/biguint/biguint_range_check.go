package biguint

import (
	"github.com/consensys/gnark/std/rangecheck"
)

func (api bigUintAPI) AssertRangeCheck(ins ...BigUint) {
	rc := rangecheck.New(api.core)
	for _, in := range ins {
		for i := range in.Limbs {
			rc.Check(in.Limbs[i], api.setup.LimbBits)
		}
	}
}
