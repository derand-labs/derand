package bigint

import (
	"zkvdf/biguint"
)

type Setup struct {
	BigUint *biguint.Setup
}

func NewSetup(limbbits int) *Setup {
	return &Setup{BigUint: biguint.NewSetup(limbbits)}
}
