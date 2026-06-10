package bigint

import (
	"github.com/consensys/gnark/frontend"
)

func (api bigIntAPI) Select(selector frontend.Variable, a, b BigInt) BigInt {
	var out BigInt
	out.Sign = api.core.Select(selector, a.Sign, b.Sign)
	out.Mag = api.BigUintAPI().Select(selector, a.Mag, b.Mag)
	return out
}
