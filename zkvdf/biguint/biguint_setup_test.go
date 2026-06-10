package biguint_test

import "zkvdf/biguint"

const (
	targetbits = 1024
	limbbits   = 64
	numlimbs   = (targetbits-1)/limbbits + 1
)

var setup = biguint.NewSetup(limbbits)
