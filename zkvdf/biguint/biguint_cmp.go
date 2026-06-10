package biguint

import (
	"github.com/consensys/gnark/frontend"
)

func (api bigUintAPI) IsEqual(x, y BigUint) frontend.Variable {
	maxLimbs := max(len(x.Limbs), len(y.Limbs))
	diffSum := frontend.Variable(0)

	for i := range maxLimbs {
		diff := api.core.Sub(x.GetLimb(i), y.GetLimb(i))
		diffSum = api.core.Add(diffSum, api.core.Mul(diff, diff))
	}

	return api.core.IsZero(diffSum)
}

func (api bigUintAPI) IsEqualUI(x BigUint, ui uint64) frontend.Variable {
	firstLimbEqualUI := api.core.IsZero(api.core.Sub(x.Limbs[0], ui))
	otherLimbsZero := api.IsZero(BigUint{Limbs: x.Limbs[1:]})

	return api.core.Mul(firstLimbEqualUI, otherLimbsZero)
}

func (api bigUintAPI) IsNonEqual(x, y BigUint) frontend.Variable {
	return api.core.Sub(1, api.IsEqual(x, y))
}

func (api bigUintAPI) IsZero(x BigUint) frontend.Variable {
	sum := api.SumLimbs(x)
	return api.core.IsZero(sum)
}

func (api bigUintAPI) IsNonZero(x BigUint) frontend.Variable {
	return api.core.Sub(1, api.IsZero(x))
}

func (api bigUintAPI) IsGreater(x, y BigUint) frontend.Variable {
	maxLimbs := max(len(x.Limbs), len(y.Limbs))
	comparator := NewLimbComparator(api.core, api.setup.LimbBits)

	result := frontend.Variable(0)
	stillEq := frontend.Variable(1)

	for i := maxLimbs - 1; i >= 0; i-- {
		xi := x.GetLimb(i)
		yi := y.GetLimb(i)

		isGt := comparator.IsLessEq(yi, xi)
		isEq := api.core.IsZero(api.core.Sub(xi, yi))

		result = api.core.Select(stillEq, isGt, result)
		stillEq = api.core.Mul(stillEq, isEq)
	}

	return api.core.And(result, api.core.Sub(1, stillEq))
}

func (api bigUintAPI) IsGreaterEq(x, y BigUint) frontend.Variable {
	maxLimbs := max(len(x.Limbs), len(y.Limbs))
	comparator := NewLimbComparator(api.core, api.setup.LimbBits)

	result := frontend.Variable(0)
	stillEq := frontend.Variable(1)

	for i := maxLimbs - 1; i >= 0; i-- {
		xi := x.GetLimb(i)
		yi := y.GetLimb(i)

		isEq := api.core.IsZero(api.core.Sub(xi, yi))
		isGt := comparator.IsLessEq(yi, xi)

		result = api.core.Select(stillEq, isGt, result)
		stillEq = api.core.Mul(stillEq, isEq)
	}

	return api.core.Or(result, stillEq)
}

func (api bigUintAPI) IsLess(x, y BigUint) frontend.Variable {
	return api.IsGreater(y, x)
}

func (api bigUintAPI) IsLessEq(x, y BigUint) frontend.Variable {
	return api.IsGreaterEq(y, x)
}

func (api bigUintAPI) AssertIsEqual(x, y BigUint) {
	minLimbs := min(len(x.Limbs), len(y.Limbs))
	for i := range minLimbs {
		api.core.AssertIsEqual(x.GetLimb(i), y.GetLimb(i))
	}

	if len(x.Limbs) > minLimbs {
		api.AssertIsZero(BigUint{Limbs: x.Limbs[minLimbs:]})
	}

	if len(y.Limbs) > minLimbs {
		api.AssertIsZero(BigUint{Limbs: y.Limbs[minLimbs:]})
	}
}

func (api bigUintAPI) AssertIsEqualUI(x BigUint, y uint64) {
	api.core.AssertIsEqual(x.Limbs[0], y)
	api.AssertIsZero(BigUint{Limbs: x.Limbs[1:]})
}

func (api bigUintAPI) AssertIsNonEqual(x, y BigUint) {
	diffSum := frontend.Variable(0)
	maxLimbs := max(len(x.Limbs), len(y.Limbs))

	for i := range maxLimbs {
		diff := api.core.Sub(x.GetLimb(i), y.GetLimb(i))
		diffSum = api.core.Add(diffSum, api.core.Mul(diff, diff))
	}

	api.core.Inverse(diffSum)
}

func (api bigUintAPI) AssertIsZero(x BigUint) {
	api.core.AssertIsEqual(api.SumLimbs(x), 0)
}

func (api bigUintAPI) AssertIsNonZero(x BigUint) {
	api.core.Inverse(api.SumLimbs(x))
}
