package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type MulTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
	B biguint.BigUint `gnark:",public"`
	C biguint.BigUint `gnark:",public"`
}

func (circuit *MulTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := biguintapi.Mul(circuit.A, circuit.B)
	biguintapi.AssertIsEqual(C, circuit.C)

	return nil
}

func TestBigUintMulCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).Mul(a, b)
			},
		},
		4447,
	)
}

func TestBigUintMulCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&MulTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(120), numlimbs*2),
		},
	)
}

func TestBigUintMulCircuitSmallCorrect2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&MulTestCircuit{
			A: setup.From(big.NewInt(12), 1),
			B: setup.From(big.NewInt(10), 1),
			C: setup.From(big.NewInt(120), 2),
		},
	)
}

func TestBigUintMulCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&MulTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(121), numlimbs*2),
		},
	)
}

func TestBigUintMulCircuitBigCorrect(t *testing.T) {
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

func TestBigUintMulCircuitBigIncorrect(t *testing.T) {
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
