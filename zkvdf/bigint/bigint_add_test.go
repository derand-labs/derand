package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type AddTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
	C bigint.BigInt `gnark:",public"`
}

func (circuit *AddTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.C)

	C := bigintapi.Add(circuit.A, circuit.B)
	bigintapi.AssertIsEqual(C, circuit.C)

	return nil
}

type AddOneTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
}

func (circuit *AddOneTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B)

	B := bigintapi.AddOne(circuit.A)
	bigintapi.AssertIsEqual(B, circuit.B)

	return nil
}

func TestBigIntAddCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).Add(a, b)
			},
		},
		2641,
	)
}

func TestBigIntAddOneCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).AddOne(a)
			},
		},
		308,
	)
}

func TestBigIntAddCircuitSmallCorrect1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AddTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(24), numlimbs),
			C: setup.From(big.NewInt(36), numlimbs),
		},
	)
}

func TestBigIntAddCircuitSmallCorrect2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AddTestCircuit{
			A: setup.From(big.NewInt(6), numlimbs),
			B: setup.From(big.NewInt(-5), numlimbs),
			C: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntAddCircuitSmallIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&AddTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(24), numlimbs),
			C: setup.From(big.NewInt(37), numlimbs),
		},
	)
}

func TestBigIntAddCircuitBigBothPositiveCorrect(t *testing.T) {
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

func TestBigIntAddCircuitBigPositiveNegativeCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	B = B.Neg(B)
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

func TestBigIntAddCircuitBigNegativePositiveCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	A = A.Neg(A)
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

func TestBigIntAddCircuitBigBothNegativeCorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	A = A.Neg(A)
	B = B.Neg(B)
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

func TestBigIntAddCircuitBigBothPositiveIncorrect(t *testing.T) {
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

func TestBigIntAddCircuitBigPositiveNegativeIncorrect(t *testing.T) {
	A := commontest.BigInt1()
	B := commontest.BigInt2()
	B = B.Neg(B)
	C := new(big.Int).Add(A, B)
	C = C.Sub(C, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&AddTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(C, numlimbs),
		},
	)
}

func TestBigIntAddCircuitOverflow(t *testing.T) {
	A := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	B := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)

	C := new(big.Int).Add(A, B)
	signedC := new(big.Int).Sub(C, utils.Modulus(numlimbs*setup.BigUint.LimbBits))

	commontest.TestCircuitInvalid(
		t,
		&AddTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(signedC, numlimbs),
		},
	)
}

func TestBigIntAddCircuitUnderflow(t *testing.T) {
	A := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	A.Neg(A)

	B := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	B.Neg(B)
	B.Add(B, big.NewInt(100))

	C := new(big.Int).Add(A, B)
	signedC := new(big.Int).Add(C, utils.Modulus(numlimbs*setup.BigUint.LimbBits))

	commontest.TestCircuitInvalid(
		t,
		&AddTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
			C: setup.From(signedC, numlimbs),
		},
	)
}

func TestBigIntAddOneCircuit1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AddOneTestCircuit{
			A: setup.From(big.NewInt(0), numlimbs),
			B: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntAddOneCircuit2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AddOneTestCircuit{
			A: setup.From(big.NewInt(1), numlimbs),
			B: setup.From(big.NewInt(2), numlimbs),
		},
	)
}

func TestBigIntAddOneCircuit3(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AddOneTestCircuit{
			A: setup.From(big.NewInt(-1), numlimbs),
			B: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntAddOneCircuit4(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&AddOneTestCircuit{
			A: setup.From(big.NewInt(-2), numlimbs),
			B: setup.From(big.NewInt(-1), numlimbs),
		},
	)
}

func TestBigIntAddOneCircuit5(t *testing.T) {
	A := commontest.BigInt1()
	B := new(big.Int).Add(A, big.NewInt(1))

	commontest.TestCircuitValid(
		t,
		&AddOneTestCircuit{
			A: setup.From(A, numlimbs),
			B: setup.From(B, numlimbs),
		},
	)
}
