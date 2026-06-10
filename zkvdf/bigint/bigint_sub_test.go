package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type SubTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
	C bigint.BigInt `gnark:",public"`
}

func (circuit *SubTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := bigintapi.Sub(circuit.A, circuit.B)
	bigintapi.AssertIsEqual(C, circuit.C)
	return nil
}

type SubOneTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
}

func (circuit *SubOneTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B)

	B := bigintapi.SubOne(circuit.A)
	bigintapi.AssertIsEqual(B, circuit.B)
	return nil
}

func TestBigIntSubCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).Sub(a, b)
			},
		},
		2641,
	)
}

func TestBigIntSubCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubTestCircuit{
			A: setup.From(big.NewInt(36), numlimbs),
			B: setup.From(big.NewInt(24), numlimbs),
			C: setup.From(big.NewInt(12), numlimbs),
		},
	)
}

func TestBigIntSubCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(big.NewInt(36), numlimbs),
			B: setup.From(big.NewInt(24), numlimbs),
			C: setup.From(big.NewInt(11), numlimbs),
		},
	)
}

func TestBigIntSubCircuitBigBothPositiveCorrect(t *testing.T) {
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

func TestBigIntSubCircuitBigPositiveNegativeCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	B = B.Neg(B)
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

func TestBigIntSubCircuitBigBothPositiveIncorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	C := new(big.Int).Sub(A, B)
	C = C.Add(C, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}

func TestBigIntSubCircuitBigPositiveNegativeIncorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	B = B.Neg(B)
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

func TestBigIntSubCircuitOverflow(t *testing.T) {
	A := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	B := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	B.Neg(B)

	C := new(big.Int).Sub(A, B)
	signedC := new(big.Int).Sub(C, utils.Modulus(numlimbs*setup.BigUint.LimbBits))

	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(signedC, numlimbs),
		},
	)
}

func TestBigIntSubCircuitUnderflow(t *testing.T) {
	A := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	A.Neg(A)
	B := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	C := new(big.Int).Sub(A, B)
	signedC := new(big.Int).Add(C, utils.Modulus(numlimbs*setup.BigUint.LimbBits))

	commontest.TestCircuitInvalid(
		t,
		&SubTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(signedC, numlimbs),
		},
	)
}

func TestBigIntSubOneeCircuit1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A: setup.From(big.NewInt(1), numlimbs),
			B: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntSubOneeCircuit2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A: setup.From(big.NewInt(5), numlimbs),
			B: setup.From(big.NewInt(4), numlimbs),
		},
	)
}

func TestBigIntSubOneeCircuit3(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A: setup.From(big.NewInt(0), numlimbs),
			B: setup.From(big.NewInt(-1), numlimbs),
		},
	)
}

func TestBigIntSubOneeCircuit4(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&SubOneTestCircuit{
			A: setup.From(big.NewInt(-1), numlimbs),
			B: setup.From(big.NewInt(-2), numlimbs),
		},
	)
}
