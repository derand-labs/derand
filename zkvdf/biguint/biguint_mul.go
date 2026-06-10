package biguint

import (
	"math/big"
	"math/bits"
	"zkvdf/utils"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/rangecheck"
)

func init() {
	solver.RegisterHint(hintCarryMul, hintMul)
}

func (api bigUintAPI) Mul(x, y BigUint) BigUint {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)
	hintInputs = append(hintInputs, y.Limbs...)

	mulHint, err := api.core.Compiler().NewHint(hintMul, len(x.Limbs)+len(y.Limbs), hintInputs...)
	if err != nil {
		panic(err)
	}

	z := FromHint(mulHint)
	api.AssertRangeCheck(z)
	api.assertMul(z, x, y)
	return z
}

func (api bigUintAPI) assertMul(z, x, y BigUint) {
	minNumLimbs := min(len(x.Limbs), len(y.Limbs))
	if 2*api.setup.LimbBits+bits.Len(uint(minNumLimbs-1)) >= 254 {
		panic("biguint.assertMul: number of limbs is too large")
	}

	B := frontend.Variable(utils.Modulus(api.setup.LimbBits))

	rc := rangecheck.New(api.core)

	carry := frontend.Variable(0)
	for k := range z.Limbs {
		// sk = sum(xi * yj) + carry_in
		var sk frontend.Variable = carry

		for i := range x.Limbs {
			j := k - i
			if j >= 0 && j < len(y.Limbs) {
				sk = api.core.Add(sk, api.core.Mul(x.Limbs[i], y.Limbs[j]))
			}
		}

		carryHint, err := api.core.Compiler().NewHint(hintCarryMul, 1, api.setup.LimbBits, sk)
		if err != nil {
			panic(err)
		}

		// Range check carry to prevent field wrap-around.
		carryBits := api.setup.LimbBits + bits.Len(uint(minNumLimbs-1))
		rc.Check(carryHint[0], carryBits)

		// sk = z[k] + carry_out * B
		api.core.AssertIsEqual(sk, api.core.Add(z.Limbs[k], api.core.Mul(carryHint[0], B)))

		carry = carryHint[0]
	}

	// Final carry must be zero for full product
	api.core.AssertIsEqual(carry, 0)
}

func hintCarryMul(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	outputs[0].Rsh(inputs[1], uint(inputs[0].Int64()))
	return nil
}

func hintMul(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)
	y := FromHintInput(inputs[xNumLimbs+prefixInputs:], limbbits)

	z := new(big.Int).Mul(x, y)
	ToHintOutput(outputs, z, len(outputs), limbbits)
	return nil
}
