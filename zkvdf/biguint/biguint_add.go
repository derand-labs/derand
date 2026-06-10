package biguint

import (
	"math/big"
	"zkvdf/utils"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

func init() {
	solver.RegisterHint(hintCarryAdd, hintAdd)
}

func (api bigUintAPI) Add(x, y BigUint) BigUint {
	znlimbs := max(len(x.Limbs), len(y.Limbs)) + 1

	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)
	hintInputs = append(hintInputs, y.Limbs...)

	addHint, err := api.core.Compiler().NewHint(hintAdd, znlimbs, hintInputs...)
	if err != nil {
		panic(err)
	}

	z := FromHint(addHint)
	api.AssertRangeCheck(z)
	api.assertAdd(z, x, y)
	return z
}

func (api bigUintAPI) AddOne(in BigUint) BigUint {
	M := utils.Modulus(api.setup.LimbBits)

	carry := frontend.Variable(1)
	out := New(len(in.Limbs) + 1)

	for i := range out.Limbs {
		x := in.GetLimb(i)
		tmp := api.core.Add(x, carry)
		carry = api.core.IsZero(api.core.Sub(M, tmp))
		out.Limbs[i] = api.core.Select(carry, 0, tmp)
	}

	api.core.AssertIsEqual(carry, 0)

	return out
}

func (api bigUintAPI) assertAdd(z, x, y BigUint) {
	M := frontend.Variable(utils.Modulus(api.setup.LimbBits))

	carry := frontend.Variable(0)
	for i := range z.Limbs {
		xi := x.GetLimb(i)
		yi := y.GetLimb(i)

		sum := api.core.Add(xi, yi, carry)
		carryHint, err := api.core.Compiler().NewHint(hintCarryAdd, 1, api.setup.LimbBits, sum)
		if err != nil {
			panic(err)
		}

		api.core.AssertIsBoolean(carryHint[0])
		api.core.AssertIsEqual(sum, api.core.Add(z.GetLimb(i), api.core.Mul(carryHint[0], M)))

		carry = carryHint[0]
	}

	api.core.AssertIsEqual(carry, 0)
}

func hintCarryAdd(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	outputs[0].Rsh(inputs[1], uint(inputs[0].Int64()))
	return nil
}

func hintAdd(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())
	yNumLimbs := len(inputs) - xNumLimbs - prefixInputs
	zNumLimbs := max(xNumLimbs, yNumLimbs) + 1

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)
	y := FromHintInput(inputs[xNumLimbs+prefixInputs:], limbbits)

	z := new(big.Int).Add(x, y)

	ToHintOutput(outputs, z, zNumLimbs, limbbits)
	return nil
}
