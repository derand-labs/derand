package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type DivExactTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
	B biguint.BigUint `gnark:",public"`
	C biguint.BigUint `gnark:",public"`
}

func (circuit *DivExactTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := biguintapi.DivExact(circuit.A, circuit.B)
	biguintapi.AssertIsEqual(C, circuit.C)

	return nil
}

func TestBigUintMinimalDivExactCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs * 2),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).DivExact(a, b)
			},
		},
		5775,
	)
}

func TestBigUintDivExactCircuitSmallDivisible(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&DivExactTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs*2),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(12), numlimbs),
		},
	)
}

func TestBigUintDivExactCircuitSmallNonDivisible(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&DivExactTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs*2),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(12), 1),
		},
	)
}

func TestBigUintDivExactCircuitBigDivisible(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)

	commontest.TestCircuitValid(
		t,
		&DivExactTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}

func TestBigUintDivExactCircuitBigNonDivisible(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)
	A.Add(A, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&DivExactTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}
