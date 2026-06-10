package bigint

import (
	"zkvdf/biguint"

	"github.com/consensys/gnark/frontend"
)

type API interface {
	BigUintAPI() biguint.API

	AssertRangeCheck(ins ...BigInt)
	Select(selector frontend.Variable, a, b BigInt) BigInt

	Cast(a BigInt, nlimbs int) BigInt

	Abs(a BigInt) BigInt
	Neg(a BigInt) BigInt

	Add(a, b BigInt) BigInt
	AddOne(a BigInt) BigInt

	Sub(a, b BigInt) BigInt
	SubOne(a BigInt) BigInt

	Mul(a, b BigInt) BigInt

	DivExact(a, b BigInt) BigInt
	Div2Exact(a BigInt) BigInt

	FloorDivQuoRem(a, b BigInt) (BigInt, BigInt)
	FloorDiv2QuoRem(a BigInt) (BigInt, frontend.Variable)
	FloorDiv4QuoRem(a BigInt) (BigInt, frontend.Variable)

	GCD(a, b BigInt) (g, ga, gb, x, y BigInt)

	InvMod(a, b BigInt) BigInt

	IsSameSign(x, y BigInt) frontend.Variable
	IsSameSignOrZero(x, y BigInt) frontend.Variable
	IsEqual(x, y BigInt) frontend.Variable
	IsEqualUI(x BigInt, ui uint64) frontend.Variable
	IsNonEqual(x, y BigInt) frontend.Variable
	IsZero(x BigInt) frontend.Variable
	IsNonZero(x BigInt) frontend.Variable
	IsNegative(x BigInt) frontend.Variable
	IsNonNegative(x BigInt) frontend.Variable
	IsPositive(x BigInt) frontend.Variable
	IsNonPositive(x BigInt) frontend.Variable

	AssertIsEqual(x, y BigInt)
	AssertIsEqualUI(x BigInt, y uint64)
	AssertIsNonEqual(x, y BigInt)
	AssertIsZero(x BigInt)
	AssertIsNonZero(x BigInt)
	AssertIsNegative(x BigInt)
	AssertIsNonNegative(x BigInt)
	AssertIsPositive(x BigInt)
	AssertIsNonPositive(x BigInt)
}

type bigIntAPI struct {
	core       frontend.API
	setup      *Setup
	biguintAPI biguint.API
}

func NewAPI(api frontend.API, setup *Setup) API {
	return bigIntAPI{
		core:       api,
		setup:      setup,
		biguintAPI: biguint.NewAPI(api, setup.BigUint),
	}
}

func (api bigIntAPI) BigUintAPI() biguint.API {
	return api.biguintAPI
}
