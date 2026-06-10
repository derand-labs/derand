package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type DivQuoRemTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
	B biguint.BigUint `gnark:",public"`
	Q biguint.BigUint `gnark:",public"`
	R biguint.BigUint `gnark:",public"`
}

func (circuit *DivQuoRemTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.Q)

	Q, R := biguintapi.DivQuoRem(circuit.A, circuit.B)
	biguintapi.AssertIsEqual(Q, circuit.Q)
	biguintapi.AssertIsEqual(R, circuit.R)

	return nil
}

type Div2QuoRemTestCircuit struct {
	A biguint.BigUint   `gnark:",public"`
	Q biguint.BigUint   `gnark:",public"`
	R frontend.Variable `gnark:",public"`
}

func (circuit *Div2QuoRemTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.Q)
	api.AssertIsBoolean(circuit.R)

	Q, R := biguintapi.Div2QuoRem(circuit.A)
	biguintapi.AssertIsEqual(Q, circuit.Q)
	api.AssertIsEqual(R, circuit.R)

	return nil
}

type Div4QuoRemTestCircuit struct {
	A biguint.BigUint   `gnark:",public"`
	Q biguint.BigUint   `gnark:",public"`
	R frontend.Variable `gnark:",public"`
}

func (circuit *Div4QuoRemTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.Q)

	Q, R := biguintapi.Div4QuoRem(circuit.A)
	biguintapi.AssertIsEqual(Q, circuit.Q)
	api.AssertIsEqual(R, circuit.R)

	return nil
}

func TestBigUintDivQuoRemCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs * 2),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).DivQuoRem(a, b)
			},
		},
		12154,
	)
}

func TestBigUintDiv2QuoRemCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs * 2),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).Div2QuoRem(a)
			},
		},
		4084,
	)
}

func TestBigUintDiv4QuoRemCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs * 2),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).Div4QuoRem(a)
			},
		},
		6745,
	)
}

func TestBigUintDivQuoRemCircuitSmallDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs*2),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), numlimbs),
			R: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigUintDivQuoRemCircuitSmallDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs*2),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), numlimbs),
			R: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigUintDivQuoRemCircuitSmallNonDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs*2),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), 1),
			R: setup.From(big.NewInt(1), 1),
		},
	)
}

func TestBigUintDivQuoRemCircuitSmallNonDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs*2),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), 1),
			R: setup.From(big.NewInt(0), 1),
		},
	)
}

func TestBigUintDivQuoRemCircuitBigDivisibleCorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)

	commontest.TestCircuitValid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(0), 1),
		},
	)
}

func TestBigUintDivQuoRemCircuitBigDivisibleIncorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)

	commontest.TestCircuitInvalid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(B, numlimbs),
		},
	)
}

func TestBigUintDivQuoRemCircuitBigNonDivisibleCorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)
	A.Add(A, big.NewInt(1222))

	commontest.TestCircuitValid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(1222), 1),
		},
	)
}

func TestBigUintDivQuoRemCircuitBigNonDivisibleIncorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)
	A.Add(A, big.NewInt(1222))

	commontest.TestCircuitInvalid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(1221), 1),
		},
	)
}

func TestBigUintDivQuoRemCircuitTruthTableTruthTableZero(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(big.NewInt(1), numlimbs*2),
			B: setup.From(big.NewInt(2), numlimbs),
			Q: setup.From(big.NewInt(0), numlimbs),
			R: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigUintDivQuoRemCircuitTruthTableTruthTableDivisible(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&DivQuoRemTestCircuit{
			A: setup.From(big.NewInt(8), numlimbs*2),
			B: setup.From(big.NewInt(2), numlimbs),
			Q: setup.From(big.NewInt(4), numlimbs),
			R: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigUintDiv2QuoRemCircuitTruthTableTruthTableDivisible(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&Div2QuoRemTestCircuit{
			A: setup.From(big.NewInt(8), numlimbs*2),
			Q: setup.From(big.NewInt(4), numlimbs),
			R: 0,
		},
	)
}

func TestBigUintDiv2QuoRemCircuitTruthTableTruthTableNonDivisible(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&Div2QuoRemTestCircuit{
			A: setup.From(big.NewInt(9), numlimbs*2),
			Q: setup.From(big.NewInt(4), numlimbs),
			R: 1,
		},
	)
}

func TestBigUintDiv4QuoRemCircuitTruthTableTruthTable0(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&Div4QuoRemTestCircuit{
			A: setup.From(big.NewInt(8), numlimbs*2),
			Q: setup.From(big.NewInt(2), numlimbs),
			R: 0,
		},
	)
}

func TestBigUintDiv4QuoRemCircuitTruthTableTruthTable1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&Div4QuoRemTestCircuit{
			A: setup.From(big.NewInt(9), numlimbs*2),
			Q: setup.From(big.NewInt(2), numlimbs),
			R: 1,
		},
	)
}

func TestBigUintDiv4QuoRemCircuitTruthTableTruthTable2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&Div4QuoRemTestCircuit{
			A: setup.From(big.NewInt(10), numlimbs*2),
			Q: setup.From(big.NewInt(2), numlimbs),
			R: 2,
		},
	)
}

func TestBigUintDiv4QuoRemCircuitTruthTableTruthTable3(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&Div4QuoRemTestCircuit{
			A: setup.From(big.NewInt(11), numlimbs*2),
			Q: setup.From(big.NewInt(2), numlimbs),
			R: 3,
		},
	)
}
