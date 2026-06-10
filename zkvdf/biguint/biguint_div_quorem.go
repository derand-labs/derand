package biguint

import (
	"math/big"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/rangecheck"
)

func init() {
	solver.RegisterHint(hintDivQuorem, hintDiv2Quorem, hintDiv4Quorem)
}

func (api bigUintAPI) DivQuoRem(x, y BigUint) (BigUint, BigUint) {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)
	hintInputs = append(hintInputs, y.Limbs...)

	divQuoRemHint, err := api.core.Compiler().NewHint(hintDivQuorem, len(x.Limbs)+len(y.Limbs), hintInputs...)
	if err != nil {
		panic(err)
	}

	quo := FromHint(divQuoRemHint[:len(x.Limbs)])
	rem := FromHint(divQuoRemHint[len(x.Limbs):])
	api.AssertRangeCheck(quo, rem)
	api.assertDivQuoRem(quo, rem, x, y)
	return quo, rem
}

func (api bigUintAPI) Div2QuoRem(x BigUint) (BigUint, frontend.Variable) {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)

	divQuoRemHint, err := api.core.Compiler().NewHint(hintDiv2Quorem, len(x.Limbs)+1, hintInputs...)
	if err != nil {
		panic(err)
	}

	quo := FromHint(divQuoRemHint[:len(x.Limbs)])
	rem := divQuoRemHint[len(x.Limbs)]
	api.AssertRangeCheck(quo)
	api.core.AssertIsBoolean(rem)

	api.assertDiv2QuoRem(quo, rem, x)
	return quo, rem
}

func (api bigUintAPI) Div4QuoRem(x BigUint) (BigUint, frontend.Variable) {
	hintInputs := []frontend.Variable{api.setup.LimbBits, len(x.Limbs)}
	hintInputs = append(hintInputs, x.Limbs...)

	divQuoRemHint, err := api.core.Compiler().NewHint(hintDiv4Quorem, len(x.Limbs)+1, hintInputs...)
	if err != nil {
		panic(err)
	}

	rc := rangecheck.New(api.core)

	quo := FromHint(divQuoRemHint[:len(x.Limbs)])
	rem := divQuoRemHint[len(x.Limbs)]
	api.AssertRangeCheck(quo)
	rc.Check(rem, 2)

	api.assertDiv4QuoRem(quo, rem, x)
	return quo, rem
}

func (api bigUintAPI) assertDivQuoRem(q, r, x, y BigUint) {
	yq := api.Mul(y, q)
	yq = api.Cast(yq, len(x.Limbs))
	rhs := api.Add(r, yq)

	api.AssertIsNonZero(y)
	api.AssertIsEqual(x, rhs)
	api.core.AssertIsEqual(api.IsLess(r, y), 1)
}

func (api bigUintAPI) assertDiv2QuoRem(q BigUint, r frontend.Variable, x BigUint) {
	q2 := api.Add(q, q)
	q2 = api.Cast(q2, len(x.Limbs))
	rhs := api.Select(r, api.AddOne(q2), q2)

	api.AssertIsEqual(x, rhs)
}

func (api bigUintAPI) assertDiv4QuoRem(q BigUint, r frontend.Variable, x BigUint) {
	q2 := api.Add(q, q)
	q4 := api.Add(q2, q2)
	q4 = api.Cast(q4, len(x.Limbs))
	rhs := api.Add(q4, BigUint{Limbs: []frontend.Variable{r}})

	api.AssertIsEqual(x, rhs)
}

func hintDivQuorem(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)
	y := FromHintInput(inputs[xNumLimbs+prefixInputs:], limbbits)

	quo, rem := new(big.Int).QuoRem(x, y, new(big.Int))
	ToHintOutput(outputs[:xNumLimbs], quo, xNumLimbs, limbbits)
	ToHintOutput(outputs[xNumLimbs:], rem, len(outputs)-xNumLimbs, limbbits)
	return nil
}

func hintDiv2Quorem(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)

	quo, rem := new(big.Int).QuoRem(x, big.NewInt(2), new(big.Int))
	ToHintOutput(outputs[:len(outputs)-1], quo, len(outputs)-1, limbbits)
	outputs[len(outputs)-1] = rem

	return nil
}

func hintDiv4Quorem(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	xNumLimbs := int(inputs[1].Int64())

	x := FromHintInput(inputs[prefixInputs:xNumLimbs+prefixInputs], limbbits)

	quo, rem := new(big.Int).QuoRem(x, big.NewInt(4), new(big.Int))
	ToHintOutput(outputs[:len(outputs)-1], quo, len(outputs)-1, limbbits)
	outputs[len(outputs)-1] = rem

	return nil
}
