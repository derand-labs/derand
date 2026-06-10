package biguint

import (
	"math/big"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

func init() {
	solver.RegisterHint(hintDivExact, hintDiv2Exact)
}

func (api bigUintAPI) DivExact(x, y BigUint) BigUint {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)
	hintInputs = append(hintInputs, y.Limbs...)

	divExactHint, err := api.core.Compiler().NewHint(hintDivExact, len(x.Limbs), hintInputs...)
	if err != nil {
		panic(err)
	}

	z := FromHint(divExactHint)
	api.AssertRangeCheck(z)
	api.assertDivExact(z, x, y)
	return z
}

func (api bigUintAPI) Div2Exact(x BigUint) BigUint {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)

	divExactHint, err := api.core.Compiler().NewHint(hintDiv2Exact, len(x.Limbs), hintInputs...)
	if err != nil {
		panic(err)
	}

	z := FromHint(divExactHint)
	api.AssertRangeCheck(z)
	api.assertAdd(x, z, z)
	return z
}

func (api bigUintAPI) assertDivExact(z, x, y BigUint) {
	api.AssertIsNonZero(y)
	api.assertMul(api.Cast(x, len(y.Limbs)+len(z.Limbs)), y, z)
}

func hintDivExact(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)
	y := FromHintInput(inputs[xNumLimbs+prefixInputs:], limbbits)

	z := new(big.Int).Div(x, y)
	ToHintOutput(outputs, z, len(outputs), limbbits)
	return nil
}

func hintDiv2Exact(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)
	z := new(big.Int).Div(x, big.NewInt(2))
	ToHintOutput(outputs, z, len(outputs), limbbits)
	return nil
}
