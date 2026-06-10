package biguint

import (
	"math/big"
	"zkvdf/utils"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

func init() {
	solver.RegisterHint(hintBorrowSub, hintSub)
}

func (api bigUintAPI) SubOneWithBorrow(in BigUint) (BigUint, frontend.Variable) {
	borrow := frontend.Variable(1)
	out := New(len(in.Limbs))

	for i := range out.Limbs {
		x := in.GetLimb(i)
		newborrow := api.core.Mul(api.core.IsZero(x), borrow)
		out.Limbs[i] = api.core.Select(newborrow, utils.MaxValue(api.setup.LimbBits), api.core.Sub(x, borrow))
		borrow = newborrow
	}

	return out, borrow
}

func (api bigUintAPI) Sub(x, y BigUint) BigUint {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)
	hintInputs = append(hintInputs, y.Limbs...)

	subHint, err := api.core.Compiler().NewHint(hintSub, max(len(x.Limbs), len(y.Limbs)), hintInputs...)
	if err != nil {
		panic(err)
	}

	z := FromHint(subHint)
	api.AssertRangeCheck(z)
	api.assertSub(z, x, y)
	return z
}

func (api bigUintAPI) SubWithBorrow(x, y BigUint) (BigUint, frontend.Variable) {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)
	hintInputs = append(hintInputs, y.Limbs...)

	subHint, err := api.core.Compiler().NewHint(hintSub, max(len(x.Limbs), len(y.Limbs)), hintInputs...)
	if err != nil {
		panic(err)
	}

	z := FromHint(subHint)
	api.AssertRangeCheck(z)
	borrow := api.assertSubWithBorrow(z, x, y)
	return z, borrow
}

func (api bigUintAPI) assertSub(z, x, y BigUint) {
	borrow := api.assertSubWithBorrow(z, x, y)
	api.core.AssertIsEqual(borrow, 0)
}

func (api bigUintAPI) assertSubWithBorrow(z, x, y BigUint) frontend.Variable {
	M := frontend.Variable(utils.Modulus(api.setup.LimbBits))

	if len(z.Limbs) < max(len(x.Limbs), len(y.Limbs)) {
		panic("biguint.AssertSub: output limbs is not enough")
	}

	borrowIn := frontend.Variable(0)

	for i := range z.Limbs {
		xi := x.GetLimb(i)
		yi := y.GetLimb(i)
		zi := z.GetLimb(i)

		borrowOut, err := api.core.Compiler().NewHint(hintBorrowSub, 1, xi, yi, borrowIn)
		if err != nil {
			panic(err)
		}

		api.core.AssertIsBoolean(borrowOut[0])

		api.core.AssertIsEqual(
			api.core.Add(xi, api.core.Mul(borrowOut[0], M)),
			api.core.Add(yi, borrowIn, zi),
		)

		borrowIn = borrowOut[0]
	}

	return borrowIn
}

func hintBorrowSub(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	xi := inputs[0]
	yi := inputs[1]
	borrowIn := inputs[2]

	t := new(big.Int).Add(yi, borrowIn)

	if xi.Cmp(t) < 0 {
		outputs[0].SetUint64(1)
	} else {
		outputs[0].SetUint64(0)
	}
	return nil
}

func hintSub(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())
	yNumLimbs := len(inputs) - xNumLimbs - prefixInputs
	zNumLimbs := max(xNumLimbs, yNumLimbs)

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)
	y := FromHintInput(inputs[xNumLimbs+prefixInputs:], limbbits)

	z := new(big.Int).Sub(x, y)
	modulus := new(big.Int).Lsh(big.NewInt(1), uint(zNumLimbs*limbbits))
	if z.Sign() < 0 {
		z.Add(z, modulus)
	}

	ToHintOutput(outputs, z, zNumLimbs, limbbits)
	return nil
}
