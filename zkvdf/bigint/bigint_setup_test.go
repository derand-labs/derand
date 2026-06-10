package bigint_test

import "zkvdf/bigint"

const (
	targetbit = 1024
	limbbits  = 64
	numlimbs  = (targetbit-1)/limbbits + 1
)

var setup = bigint.NewSetup(limbbits)
