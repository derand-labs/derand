package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type AbsTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
}

func (circuit *AbsTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B)

	B := bigintapi.Abs(circuit.A)
	bigintapi.AssertIsEqual(B, circuit.B)

	return nil
}

func TestBigIntAbsCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).Abs(a)
			},
		},
		5,
	)
}

func TestBigIntAbsCircuitSmallPositive(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AbsTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(12), numlimbs),
		},
	)
}

func TestBigIntAbsCircuitSmallNegative(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AbsTestCircuit{
			A: setup.From(big.NewInt(-12), numlimbs),
			B: setup.From(big.NewInt(12), numlimbs),
		},
	)
}

func TestBigIntAbsCircuitSmallIncorrectNegativeOutput(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&AbsTestCircuit{
			A: setup.From(big.NewInt(-12), numlimbs),
			B: setup.From(big.NewInt(-12), numlimbs),
		},
	)
}

func TestBigIntAbsCircuitSmallIncorrectNegativeInput(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&AbsTestCircuit{
			A: setup.From(big.NewInt(-12), numlimbs),
			B: setup.From(big.NewInt(13), numlimbs),
		},
	)
}

func TestBigIntAbsCircuitSmallIncorrectPositiveInput(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&AbsTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(13), numlimbs),
		},
	)
}

func TestBigIntAbsCircuitBigPositive(t *testing.T) {
	A := commontest.BigInt1()

	commontest.TestCircuitValid(
		t,
		&AbsTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(A, numlimbs),
		},
	)
}

func TestBigIntAbsCircuitBigNegative(t *testing.T) {
	A := commontest.BigInt1()
	negA := new(big.Int).Neg(A)

	commontest.TestCircuitValid(
		t,
		&AbsTestCircuit{
			A: setup.From(negA, numlimbs),
			B: setup.From(A, numlimbs),
		},
	)
}
