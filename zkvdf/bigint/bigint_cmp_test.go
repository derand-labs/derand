package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type CompareUnaryTestCircuit struct {
	A bigint.BigInt `gnark:",public"`

	isZero bool
	isPos  bool
	isNeg  bool
}

func (circuit *CompareUnaryTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A)

	resZero := bigintapi.IsZero(circuit.A)
	resNonZero := bigintapi.IsNonZero(circuit.A)
	if circuit.isZero {
		api.AssertIsEqual(resZero, frontend.Variable(1))
		api.AssertIsEqual(resNonZero, frontend.Variable(0))
		bigintapi.AssertIsZero(circuit.A)
	} else {
		api.AssertIsEqual(resZero, frontend.Variable(0))
		api.AssertIsEqual(resNonZero, frontend.Variable(1))
		bigintapi.AssertIsNonZero(circuit.A)
	}

	resPos := bigintapi.IsPositive(circuit.A)
	resNonPos := bigintapi.IsNonPositive(circuit.A)
	if circuit.isPos {
		api.AssertIsEqual(resPos, frontend.Variable(1))
		api.AssertIsEqual(resNonPos, frontend.Variable(0))
		bigintapi.AssertIsPositive(circuit.A)
	} else {
		api.AssertIsEqual(resPos, frontend.Variable(0))
		api.AssertIsEqual(resNonPos, frontend.Variable(1))
		bigintapi.AssertIsNonPositive(circuit.A)
	}

	resNeg := bigintapi.IsNegative(circuit.A)
	resNonNeg := bigintapi.IsNonNegative(circuit.A)
	if circuit.isNeg {
		api.AssertIsEqual(resNeg, frontend.Variable(1))
		api.AssertIsEqual(resNonNeg, frontend.Variable(0))
		bigintapi.AssertIsNegative(circuit.A)
	} else {
		api.AssertIsEqual(resNeg, frontend.Variable(0))
		api.AssertIsEqual(resNonNeg, frontend.Variable(1))
		bigintapi.AssertIsNonNegative(circuit.A)
	}

	return nil
}

type CompareBinaryTestCircuit struct {
	A       bigint.BigInt `gnark:",public"`
	B       bigint.BigInt `gnark:",public"`
	isEqual bool
}

func (circuit *CompareBinaryTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B)

	resEqual := bigintapi.IsEqual(circuit.A, circuit.B)
	resNonEqual := bigintapi.IsNonEqual(circuit.A, circuit.B)
	if circuit.isEqual {
		api.AssertIsEqual(resEqual, frontend.Variable(1))
		api.AssertIsEqual(resNonEqual, frontend.Variable(0))
		bigintapi.AssertIsEqual(circuit.A, circuit.B)
	} else {
		api.AssertIsEqual(resEqual, frontend.Variable(0))
		api.AssertIsEqual(resNonEqual, frontend.Variable(1))
		bigintapi.AssertIsNonEqual(circuit.A, circuit.B)
	}

	return nil
}

func TestBigIntCompareUnaryCircuitConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).IsZero(a)
			},
		},
		2,
	)
}

func TestBigIntCompareBinaryCircuitConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).IsEqual(a, b)
			},
		},
		53,
	)
}

func TestBigIntCompareUnaryZero(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareUnaryTestCircuit{
		A:      setup.From(big.NewInt(0), numlimbs),
		isZero: true,
		isPos:  false,
		isNeg:  false,
	})
}

func TestBigIntCompareUnaryPosSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareUnaryTestCircuit{
		A:      setup.From(big.NewInt(10), numlimbs),
		isZero: false,
		isPos:  true,
		isNeg:  false,
	})
}

func TestBigIntCompareUnaryNegSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareUnaryTestCircuit{
		A:      setup.From(big.NewInt(-10), numlimbs),
		isZero: false,
		isPos:  false,
		isNeg:  true,
	})
}

func TestBigIntCompareBinaryEqualSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(big.NewInt(1), numlimbs),
		B:       setup.From(big.NewInt(1), numlimbs),
		isEqual: true,
	})
}

func TestBigIntCompareBinaryEqualPositiveBig(t *testing.T) {
	A := commontest.BigInt1()
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(A, numlimbs),
		B:       setup.From(A, numlimbs),
		isEqual: true,
	})
}

func TestBigIntCompareBinaryEqualNegativeBig(t *testing.T) {
	A := commontest.BigInt1()
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(new(big.Int).Neg(A), numlimbs),
		B:       setup.From(new(big.Int).Neg(A), numlimbs),
		isEqual: true,
	})
}

func TestBigIntCompareBinaryNonEqualPositveBig(t *testing.T) {
	A := commontest.BigInt1()
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(A, numlimbs),
		B:       setup.From(new(big.Int).Neg(A), numlimbs),
		isEqual: false,
	})
}

func TestBigIntCompareBinaryNonEqualNegativeBig(t *testing.T) {
	A := commontest.BigInt1()
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(new(big.Int).Neg(A), numlimbs),
		B:       setup.From(A, numlimbs),
		isEqual: false,
	})
}

func TestBigIntCompareBinaryNonEqualSmallZeroFirst(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(big.NewInt(0), numlimbs),
		B:       setup.From(big.NewInt(5), numlimbs),
		isEqual: false,
	})
}

func TestBigIntCompareBinaryNonEqualSmallZeroSecond(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(big.NewInt(5), numlimbs),
		B:       setup.From(big.NewInt(0), numlimbs),
		isEqual: false,
	})
}

func TestBigIntCompareBinaryEqualNegativeSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(big.NewInt(-1), numlimbs),
		B:       setup.From(big.NewInt(-1), numlimbs),
		isEqual: true,
	})
}

func TestBigIntCompareBinaryEqualZero(t *testing.T) {
	commontest.TestCircuitValid(t, &CompareBinaryTestCircuit{
		A:       setup.From(big.NewInt(0), numlimbs),
		B:       setup.From(big.NewInt(0), numlimbs),
		isEqual: true,
	})
}
