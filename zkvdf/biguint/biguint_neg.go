package biguint

import (
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

func (api bigUintAPI) Neg(in BigUint) BigUint {
	M := utils.Modulus(api.setup.LimbBits)

	carry := frontend.Variable(1)
	out := New(len(in.Limbs))

	for i := range in.Limbs {
		x := in.Limbs[i]
		notx := api.core.Sub(M, 1, x)

		tmp := api.core.Add(notx, carry)

		carry = api.core.IsZero(api.core.Sub(M, tmp))

		out.Limbs[i] = api.core.Select(carry, 0, tmp)
	}

	return out
}
