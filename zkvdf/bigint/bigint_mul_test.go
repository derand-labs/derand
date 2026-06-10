package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type MulTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
	C bigint.BigInt `gnark:",public"`
}

func (circuit *MulTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := bigintapi.Mul(circuit.A, circuit.B)
	bigintapi.AssertIsEqual(C, circuit.C)

	return nil
}

func TestBigIntMulCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).Mul(a, b)
			},
		},
		4448,
	)
}

func TestBigIntMulCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&MulTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(120), numlimbs*2),
		},
	)
}

func TestBigIntMulCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&MulTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(121), numlimbs*2),
		},
	)
}

func TestBigIntMulCircuitBigCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	C := new(big.Int).Mul(A, B)

	commontest.TestCircuitValid(
		t,
		&MulTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs*2),
		},
	)
}

func TestBigIntMulCircuitBigNegativeCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	A.Neg(A)
	C := new(big.Int).Mul(A, B)

	commontest.TestCircuitValid(
		t,
		&MulTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs*2),
		},
	)
}

func TestBigIntMulCircuitNegativeBothCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()

	A.Neg(A)
	B.Neg(B)
	C := new(big.Int).Mul(A, B)

	commontest.TestCircuitValid(
		t,
		&MulTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs*2),
		},
	)
}

func TestBigIntMulCircuitBigInvalid(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	C := new(big.Int).Mul(A, B)
	C.Add(C, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&MulTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs*2),
		},
	)
}
