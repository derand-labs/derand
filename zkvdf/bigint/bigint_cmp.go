package bigint

import (
	"github.com/consensys/gnark/frontend"
)

func (api bigIntAPI) IsSameSign(x, y BigInt) frontend.Variable {
	return api.core.IsZero(api.core.Sub(x.Sign, y.Sign))
}

func (api bigIntAPI) IsSameSignOrZero(x, y BigInt) frontend.Variable {
	sameSign := api.core.IsZero(api.core.Sub(x.Sign, y.Sign))
	anyZeroSign := api.core.IsZero(api.core.Mul(x.Sign, y.Sign))
	return api.core.Or(sameSign, anyZeroSign)
}

func (api bigIntAPI) IsEqual(x, y BigInt) frontend.Variable {
	isSameSign := api.core.IsZero(api.core.Sub(x.Sign, y.Sign))
	isSameMag := api.BigUintAPI().IsEqual(x.Mag, y.Mag)

	return api.core.And(isSameSign, isSameMag)
}

func (api bigIntAPI) IsEqualUI(x BigInt, ui uint64) frontend.Variable {
	nonNegative := api.core.IsZero(api.core.Add(x.Sign, 1))
	return api.core.Mul(nonNegative, api.BigUintAPI().IsEqualUI(x.Mag, ui))
}

func (api bigIntAPI) IsNonEqual(x, y BigInt) frontend.Variable {
	return api.core.Sub(1, api.IsEqual(x, y))
}

func (api bigIntAPI) IsZero(x BigInt) frontend.Variable {
	return api.core.IsZero(x.Sign)
}

func (api bigIntAPI) IsNonZero(x BigInt) frontend.Variable {
	return api.core.Sub(1, api.IsZero(x))
}

func (api bigIntAPI) IsNegative(x BigInt) frontend.Variable {
	return api.core.IsZero(api.core.Add(x.Sign, 1))
}

func (api bigIntAPI) IsNonNegative(x BigInt) frontend.Variable {
	return api.core.Sub(1, api.IsNegative(x))
}

func (api bigIntAPI) IsPositive(x BigInt) frontend.Variable {
	return api.core.IsZero(api.core.Sub(x.Sign, 1))
}

func (api bigIntAPI) IsNonPositive(x BigInt) frontend.Variable {
	return api.core.Sub(1, api.IsPositive(x))
}

func (api bigIntAPI) AssertIsEqual(x, y BigInt) {
	api.core.AssertIsEqual(x.Sign, y.Sign)
	api.BigUintAPI().AssertIsEqual(x.Mag, y.Mag)
}

func (api bigIntAPI) AssertIsEqualUI(x BigInt, y uint64) {
	api.AssertIsNonNegative(x)
	api.BigUintAPI().AssertIsEqualUI(x.Mag, y)
}

func (api bigIntAPI) AssertIsNonEqual(x, y BigInt) {
	isDiffSign := api.core.Sub(1, api.core.IsZero(api.core.Sub(x.Sign, y.Sign)))
	isDiffMag := api.BigUintAPI().IsNonEqual(x.Mag, y.Mag)

	api.core.AssertIsEqual(api.core.Or(isDiffSign, isDiffMag), 1)
}

func (api bigIntAPI) AssertIsZero(x BigInt) {
	api.core.AssertIsEqual(x.Sign, 0)
}

func (api bigIntAPI) AssertIsNonZero(x BigInt) {
	api.core.AssertIsDifferent(x.Sign, 0)
}

func (api bigIntAPI) AssertIsNegative(x BigInt) {
	api.core.AssertIsEqual(x.Sign, -1)
}

func (api bigIntAPI) AssertIsNonNegative(x BigInt) {
	api.core.AssertIsDifferent(x.Sign, -1)
}

func (api bigIntAPI) AssertIsPositive(x BigInt) {
	api.core.AssertIsEqual(x.Sign, 1)
}

func (api bigIntAPI) AssertIsNonPositive(x BigInt) {
	api.core.AssertIsDifferent(x.Sign, 1)
}
