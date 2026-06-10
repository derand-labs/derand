package biguint

import (
	"github.com/consensys/gnark/frontend"
)

type API interface {
	AssertRangeCheck(ins ...BigUint)
	Select(selector frontend.Variable, a, b BigUint) BigUint
	Swap(selector frontend.Variable, a, b BigUint) (BigUint, BigUint)

	Cast(a BigUint, nlimbs int) BigUint

	Neg(a BigUint) BigUint

	Add(a, b BigUint) BigUint
	AddOne(a BigUint) BigUint

	Sub(a, b BigUint) BigUint
	SubWithBorrow(a, b BigUint) (BigUint, frontend.Variable)
	SubOneWithBorrow(a BigUint) (BigUint, frontend.Variable)

	Mul(a, b BigUint) BigUint

	DivExact(a, b BigUint) BigUint
	Div2Exact(a BigUint) BigUint

	DivQuoRem(a, b BigUint) (BigUint, BigUint)
	Div2QuoRem(a BigUint) (BigUint, frontend.Variable)
	Div4QuoRem(a BigUint) (BigUint, frontend.Variable)

	IsEqual(x, y BigUint) frontend.Variable
	IsEqualUI(x BigUint, ui uint64) frontend.Variable
	IsNonEqual(x, y BigUint) frontend.Variable
	IsZero(x BigUint) frontend.Variable
	IsNonZero(x BigUint) frontend.Variable
	IsGreater(x, y BigUint) frontend.Variable
	IsGreaterEq(x, y BigUint) frontend.Variable
	IsLess(x, y BigUint) frontend.Variable
	IsLessEq(x, y BigUint) frontend.Variable

	AssertIsEqual(x, y BigUint)
	AssertIsEqualUI(x BigUint, y uint64)
	AssertIsNonEqual(x, y BigUint)
	AssertIsZero(x BigUint)
	AssertIsNonZero(x BigUint)

	ToBinary(b BigUint) []frontend.Variable
}

type bigUintAPI struct {
	core  frontend.API
	setup *Setup
}

func NewAPI(api frontend.API, setup *Setup) API {
	return bigUintAPI{
		core:  api,
		setup: setup,
	}
}
