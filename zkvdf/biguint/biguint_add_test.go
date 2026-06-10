package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type AddTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
	B biguint.BigUint `gnark:",public"`
	C biguint.BigUint `gnark:",public"`
}

func (circuit *AddTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)

	biguintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := biguintapi.Add(circuit.A, circuit.B)
	biguintapi.AssertIsEqual(C, circuit.C)

	return nil
}

func TestBigUintMinimalAddCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).Add(a, b)
			},
		},
		1341,
	)

	commontest.AssertCircuitConstraints(t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).AddOne(a)
			},
		},
		102,
	)
}

func TestBigUintAddCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AddTestCircuit{
			A: setup.From(big.NewInt(12), 1),
			B: setup.From(big.NewInt(24), 1),
			C: setup.From(big.NewInt(36), 1),
		},
	)
}

func TestBigUintAddCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&AddTestCircuit{
			A: setup.From(big.NewInt(12), 1),
			B: setup.From(big.NewInt(24), 1),
			C: setup.From(big.NewInt(37), 1),
		},
	)
}

func TestBigUintAddCircuitBigCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	C := new(big.Int).Add(A, B)

	commontest.TestCircuitValid(
		t,
		&AddTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}

func TestBigUintAddCircuitBigIncorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	C := new(big.Int).Add(A, B)
	C = C.Add(C, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&AddTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}

func TestBigUintAddCircuitValidOverflow(t *testing.T) {
	A := utils.MaxValue(numlimbs * setup.LimbBits)
	B := utils.MaxValue(numlimbs * setup.LimbBits)
	C := new(big.Int).Add(A, B)

	commontest.TestCircuitValid(
		t,
		&AddTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, 17),
		},
	)
}

func TestBigUintAddCircuitInvalidOverflow(t *testing.T) {
	A := utils.MaxValue(numlimbs * setup.LimbBits)
	B := utils.MaxValue(numlimbs * setup.LimbBits)
	C := new(big.Int).Add(A, B)

	commontest.TestCircuitInvalid(
		t,
		&AddTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.FromUnsafe(C, numlimbs),
		},
	)
}
