package bigint

import (
	"zkvdf/biguint"

	"github.com/consensys/gnark/frontend"
)

type BigInt struct {
	Sign frontend.Variable
	Mag  biguint.BigUint
}
