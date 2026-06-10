package biguint

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/cmp"
)

func NewLimbComparator(api frontend.API, limbbits int) *cmp.BoundedComparator {
	bound := new(big.Int).Lsh(big.NewInt(1), uint(limbbits))
	bound.Sub(bound, big.NewInt(1))
	return cmp.NewBoundedComparator(api, bound, false)
}
