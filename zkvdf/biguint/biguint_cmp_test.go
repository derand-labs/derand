package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type CompareUnaryTestCircuit struct {
	A      biguint.BigUint `gnark:",public"`
	isZero bool
}

func (circuit *CompareUnaryTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)

	biguintapi.AssertRangeCheck(circuit.A)

	resZero := biguintapi.IsZero(circuit.A)
	resNonZero := biguintapi.IsNonZero(circuit.A)
	if circuit.isZero {
		api.AssertIsEqual(resZero, frontend.Variable(1))
		api.AssertIsEqual(resNonZero, frontend.Variable(0))
		biguintapi.AssertIsZero(circuit.A)
	} else {
		api.AssertIsEqual(resZero, frontend.Variable(0))
		api.AssertIsEqual(resNonZero, frontend.Variable(1))
		biguintapi.AssertIsNonZero(circuit.A)
	}

	return nil
}

type CompareBinaryTestCircuit struct {
	A         biguint.BigUint `gnark:",public"`
	B         biguint.BigUint `gnark:",public"`
	isEqual   bool
	isGreater bool
	isLess    bool
}

func (circuit *CompareBinaryTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)

	biguintapi.AssertRangeCheck(circuit.A, circuit.B)

	resEqual := biguintapi.IsEqual(circuit.A, circuit.B)
	resNonEqual := biguintapi.IsNonEqual(circuit.A, circuit.B)
	resGreater := biguintapi.IsGreater(circuit.A, circuit.B)
	resGreaterOrEqual := biguintapi.IsGreaterEq(circuit.A, circuit.B)
	resLess := biguintapi.IsLess(circuit.A, circuit.B)
	resLessOrEqual := biguintapi.IsLessEq(circuit.A, circuit.B)

	if circuit.isEqual {
		api.AssertIsEqual(resEqual, frontend.Variable(1))
		api.AssertIsEqual(resNonEqual, frontend.Variable(0))
		biguintapi.AssertIsEqual(circuit.A, circuit.B)
	} else {
		api.AssertIsEqual(resEqual, frontend.Variable(0))
		api.AssertIsEqual(resNonEqual, frontend.Variable(1))
		biguintapi.AssertIsNonEqual(circuit.A, circuit.B)
	}

	if circuit.isGreater {
		api.AssertIsEqual(resGreater, frontend.Variable(1))
		api.AssertIsEqual(resLessOrEqual, frontend.Variable(0))
	} else {
		api.AssertIsEqual(resGreater, frontend.Variable(0))
		api.AssertIsEqual(resLessOrEqual, frontend.Variable(1))
	}

	if circuit.isLess {
		api.AssertIsEqual(resLess, frontend.Variable(1))
		api.AssertIsEqual(resGreaterOrEqual, frontend.Variable(0))
	} else {
		api.AssertIsEqual(resLess, frontend.Variable(0))
		api.AssertIsEqual(resGreaterOrEqual, frontend.Variable(1))
	}

	return nil
}

func TestBigUintCompareIsZeroCircuitConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).IsZero(a)
			},
		},
		17,
	)
}

func TestBigUintCompareAssertZeroCircuitConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).AssertIsZero(a)
			},
		},
		16,
	)
}

func TestBigUintCompareIsEqualCircuitConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).IsEqual(a, b)
			},
		},
		49,
	)
}

func TestBigUintCompareAssertEqualCircuitConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).AssertIsEqual(a, b)
			},
		},
		numlimbs,
	)
}

func TestBigUintCompareIsGreaterCircuitConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).IsGreater(a, b)
			},
		},
		2286,
	)
}

func TestBigUintCompareUnarySmallZero(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareUnaryTestCircuit{
		A:      setup.From(big.NewInt(0), numlimbs),
		isZero: true,
	})
}

func TestBigUintCompareUnarySmallNonZero(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareUnaryTestCircuit{
		A:      setup.From(big.NewInt(10), numlimbs),
		isZero: false,
	})
}

func TestBigUintCompareUnaryBigNonZero(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareUnaryTestCircuit{
		A:      setup.From(commontest.BigInt1(), numlimbs),
		isZero: false,
	})
}

func TestBigUintCompareBinaryEqualZero(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(big.NewInt(0), numlimbs),
		B:       setup.From(big.NewInt(0), numlimbs),
		isEqual: true,
	})
}

func TestBigUintCompareBinaryEqualSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(big.NewInt(1), numlimbs),
		B:       setup.From(big.NewInt(1), numlimbs),
		isEqual: true,
	})
}

func TestBigUintCompareBinaryGreaterSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:         setup.From(big.NewInt(5), numlimbs),
		B:         setup.From(big.NewInt(0), numlimbs),
		isGreater: true,
	})
}

func TestBigUintCompareBinaryLessSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:      setup.From(big.NewInt(0), numlimbs),
		B:      setup.From(big.NewInt(5), numlimbs),
		isLess: true,
	})
}

func TestBigUintCompareBinaryEqualBig(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(commontest.BigInt1(), numlimbs),
		B:       setup.From(commontest.BigInt1(), numlimbs),
		isEqual: true,
	})
}

func TestBigUintCompareBinaryGreaterBig(t *testing.T) {
	bigA := commontest.BigInt1()

	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:         setup.From(bigA, numlimbs),
		B:         setup.From(new(big.Int).Sub(bigA, big.NewInt(1)), numlimbs),
		isGreater: true,
	})
}

func TestBigUintCompareBinaryLessBig(t *testing.T) {
	bigA := commontest.BigInt1()

	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:      setup.From(bigA, numlimbs),
		B:      setup.From(new(big.Int).Add(bigA, big.NewInt(1)), numlimbs),
		isLess: true,
	})
}
