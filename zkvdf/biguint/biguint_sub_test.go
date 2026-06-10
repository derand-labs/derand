package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type SubTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
	B biguint.BigUint `gnark:",public"`
	C biguint.BigUint `gnark:",public"`
}

func (circuit *SubTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := biguintapi.Sub(circuit.A, circuit.B)
	biguintapi.AssertIsEqual(C, circuit.C)

	return nil
}

type SubOneTestCircuit struct {
	A      biguint.BigUint   `gnark:",public"`
	B      biguint.BigUint   `gnark:",public"`
	Borrow frontend.Variable `gnark:",public"`
}

func (circuit *SubOneTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.B)

	B, b := biguintapi.SubOneWithBorrow(circuit.A)
	biguintapi.AssertIsEqual(B, circuit.B)
	api.AssertIsEqual(b, circuit.Borrow)

	return nil
}

func TestBigUintSubCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[biguint.BigUint]{
			A: biguint.New(numlimbs),
			B: biguint.New(numlimbs),
			F: func(api frontend.API, a, b biguint.BigUint) {
				biguint.NewAPI(api, setup).Sub(a, b)
			},
		},
		1279,
	)
}

func TestBigUintSubOneCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).SubOneWithBorrow(a)
			},
		},
		126,
	)
}

func TestBigUintSubCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubTestCircuit{
			A: setup.From(big.NewInt(36), 1),
			B: setup.From(big.NewInt(24), 1),
			C: setup.From(big.NewInt(12), 1),
		},
	)
}

func TestBigUintSubCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(big.NewInt(37), 1),
			B: setup.From(big.NewInt(24), 1),
			C: setup.From(big.NewInt(12), 1),
		},
	)
}

func TestBigUintSubCircuitBigCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	C := new(big.Int).Sub(A, B)

	commontest.TestCircuitValid(
		t,
		&SubTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}

func TestBigUintSubCircuitBigIncorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	C := new(big.Int).Sub(A, B)
	C = C.Sub(C, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}

func TestBigUintSubCircuitInvalidNegative(t *testing.T) {
	A := new(big.Int).Sub(utils.MaxValue(numlimbs*setup.LimbBits), big.NewInt(1))
	B := utils.MaxValue(numlimbs * setup.LimbBits)

	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigUintSubCircuitSmallUnderflow(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(big.NewInt(5), numlimbs),
			B: setup.From(big.NewInt(6), numlimbs),
			C: setup.From(utils.MaxValue(numlimbs*setup.LimbBits), numlimbs),
		},
	)
}

func TestBigUintSubOneCircuitSmall(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A:      setup.From(big.NewInt(6), numlimbs),
			B:      setup.From(big.NewInt(5), numlimbs),
			Borrow: 0,
		},
	)
}

func TestBigUintSubOneCircuitSmall2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A:      setup.From(big.NewInt(1), numlimbs),
			B:      setup.From(big.NewInt(0), numlimbs),
			Borrow: 0,
		},
	)
}

func TestBigUintSubOneCircuitBig(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A:      setup.From(utils.MaxValue(2*setup.LimbBits), numlimbs),
			B:      setup.From(new(big.Int).Sub(utils.MaxValue(2*setup.LimbBits), big.NewInt(1)), numlimbs),
			Borrow: 0,
		},
	)
}

func TestBigUintSubOneCircuitZero(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A:      setup.From(big.NewInt(0), numlimbs),
			B:      setup.From(utils.MaxValue(numlimbs*setup.LimbBits), numlimbs),
			Borrow: 1,
		},
	)
}
