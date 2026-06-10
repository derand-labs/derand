package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type DivExactTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
	C bigint.BigInt `gnark:",public"`
}

func (circuit *DivExactTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := bigintapi.DivExact(circuit.A, circuit.B)
	bigintapi.AssertIsEqual(C, circuit.C)

	return nil
}

func TestBigIntDivExactCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs * 2),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).DivExact(a, b)
			},
		},
		5776,
	)
}

func TestBigIntDiv2ExactCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs * 2),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).Div2Exact(a)
			},
		},
		2271,
	)
}

func TestBigIntDivExactCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&DivExactTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(12), 8),
		},
	)
}

func TestBigIntDivExactCircuitSmallCorrect2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&DivExactTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			B: setup.From(big.NewInt(10), 8),
			C: setup.From(big.NewInt(12), 10),
		},
	)
}

func TestBigIntDivExactCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&DivExactTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			C: setup.From(big.NewInt(12), numlimbs),
		},
	)
}

func TestBigIntDivExactCircuitBigCorrect(t *testing.T) {
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

func TestBigIntDivExactCircuitBigNegative(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	B.Neg(B)
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

func TestBigIntDivExactCircuitNegativeBoth(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()

	B.Neg(B)
	C.Neg(C)
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

func TestBigIntDivExactCircuitBigInvalid(t *testing.T) {
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
