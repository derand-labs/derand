package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type NegTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
}

func (circuit *NegTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B)

	B := bigintapi.Neg(circuit.A)
	bigintapi.AssertIsEqual(B, circuit.B)

	return nil
}

func TestBigIntNegCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).Neg(a)
			},
		},
		0,
	)
}

func TestBigIntNegCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&NegTestCircuit{
			A: setup.From(big.NewInt(24), numlimbs),
			B: setup.From(big.NewInt(-24), numlimbs),
		},
	)
}

func TestBigIntNegCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&NegTestCircuit{
			A: setup.From(big.NewInt(24), numlimbs),
			B: setup.From(big.NewInt(24), numlimbs),
		},
	)
}
