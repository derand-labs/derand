package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type FloorDivQuoRemTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
	Q bigint.BigInt `gnark:",public"`
	R bigint.BigInt `gnark:",public"`
}

func (circuit *FloorDivQuoRemTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.Q, circuit.R)

	Q, R := bigintapi.FloorDivQuoRem(circuit.A, circuit.B)
	bigintapi.AssertIsEqual(Q, circuit.Q)
	bigintapi.AssertIsEqual(R, circuit.R)

	return nil
}

type FloorDiv2QuoRemTestCircuit struct {
	A bigint.BigInt     `gnark:",public"`
	Q bigint.BigInt     `gnark:",public"`
	R frontend.Variable `gnark:",public"`
}

func (circuit *FloorDiv2QuoRemTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.Q)
	api.AssertIsBoolean(circuit.R)

	Q, R := bigintapi.FloorDiv2QuoRem(circuit.A)
	bigintapi.AssertIsEqual(Q, circuit.Q)
	api.AssertIsEqual(R, circuit.R)

	return nil
}

type FloorDiv4QuoRemTestCircuit struct {
	A bigint.BigInt     `gnark:",public"`
	Q bigint.BigInt     `gnark:",public"`
	R frontend.Variable `gnark:",public"`
}

func (circuit *FloorDiv4QuoRemTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.Q)

	Q, R := bigintapi.FloorDiv4QuoRem(circuit.A)
	bigintapi.AssertIsEqual(Q, circuit.Q)
	api.AssertIsEqual(R, circuit.R)

	return nil
}

func TestBigIntFloorDivQuoRemCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs * 2),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).FloorDivQuoRem(a, b)
			},
		},
		14461,
	)
}

func TestBigIntFloorDiv2QuoRemCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs * 2),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).FloorDiv2QuoRem(a)
			},
		},
		4525,
	)
}

func TestBigIntFloorDiv4QuoRemCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs * 2),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).FloorDiv4QuoRem(a)
			},
		},
		7186,
	)
}

func TestBigIntFloorDivQuoRemCircuitSmallDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), 1),
			R: setup.From(big.NewInt(0), 1),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitSmallDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), 1),
			R: setup.From(big.NewInt(1), 1),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitSmallNonDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), 1),
			R: setup.From(big.NewInt(1), 1),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitSmallNonDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			Q: setup.From(big.NewInt(12), 1),
			R: setup.From(big.NewInt(0), 1),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitBigDivisibleCorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)

	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitBigNegativeDivisibleCorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	B.Neg(B)
	A := new(big.Int).Mul(B, C)

	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitNegativeBothDivisible(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()

	B.Neg(B)
	C.Neg(C)
	A := new(big.Int).Mul(B, C)

	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitBigNonDivisibleCorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)
	A.Add(A, big.NewInt(1))

	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitBigNonDivisibleIncorrect(t *testing.T) {
	B := commontest.BigInt1()
	C := commontest.BigInt2()
	A := new(big.Int).Mul(B, C)
	A.Add(A, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			B: setup.From(B, numlimbs),
			Q: setup.From(C, numlimbs),
			R: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTableBothPositive(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(7), numlimbs*2),
			B: setup.From(big.NewInt(3), numlimbs),
			Q: setup.From(big.NewInt(2), numlimbs),
			R: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTableNegativePositive1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(-7), numlimbs*2),
			B: setup.From(big.NewInt(3), numlimbs),
			Q: setup.From(big.NewInt(-3), numlimbs),
			R: setup.From(big.NewInt(2), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTableNegativePositive2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(-1), numlimbs*2),
			B: setup.From(big.NewInt(3), numlimbs),
			Q: setup.From(big.NewInt(-1), numlimbs),
			R: setup.From(big.NewInt(2), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTablePositiveNegative1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(7), numlimbs*2),
			B: setup.From(big.NewInt(-3), numlimbs),
			Q: setup.From(big.NewInt(-3), numlimbs),
			R: setup.From(big.NewInt(-2), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTablePositiveNegative2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(1), numlimbs*2),
			B: setup.From(big.NewInt(-3), numlimbs),
			Q: setup.From(big.NewInt(-1), numlimbs),
			R: setup.From(big.NewInt(-2), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTableBothNegative1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(-7), numlimbs*2),
			B: setup.From(big.NewInt(-3), numlimbs),
			Q: setup.From(big.NewInt(2), numlimbs),
			R: setup.From(big.NewInt(-1), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTableBothNegative2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(-10), numlimbs*2),
			B: setup.From(big.NewInt(-3), numlimbs),
			Q: setup.From(big.NewInt(3), numlimbs),
			R: setup.From(big.NewInt(-1), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTableZero1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(1), numlimbs*2),
			B: setup.From(big.NewInt(2), numlimbs),
			Q: setup.From(big.NewInt(0), numlimbs),
			R: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntFloorDivQuoRemCircuitTruthTableZero2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDivQuoRemTestCircuit{
			A: setup.From(big.NewInt(-1), numlimbs*2),
			B: setup.From(big.NewInt(2), numlimbs),
			Q: setup.From(big.NewInt(-1), numlimbs),
			R: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitSmallDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			Q: setup.From(big.NewInt(60), 1),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitSmallDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			Q: setup.From(big.NewInt(60), 1),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitSmallNonDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs),
			Q: setup.From(big.NewInt(60), 1),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitSmallNonDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs),
			Q: setup.From(big.NewInt(60), 1),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitBigDivisibleCorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(2), C)

	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitBigNegativeDivisibleCorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(-2), C)

	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C.Neg(C), numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitNegativeBothDivisible(t *testing.T) {
	C := commontest.BigInt2()
	C.Neg(C)
	A := new(big.Int).Mul(big.NewInt(2), C)

	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitBigNonDivisibleCorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(2), C)
	A.Add(A, big.NewInt(1))

	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitBigNonDivisibleIncorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(2), C)
	A.Add(A, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitTruthTableBothPositive(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(7), numlimbs*2),
			Q: setup.From(big.NewInt(3), numlimbs),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitTruthTableNegativePositive1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(-7), numlimbs*2),
			Q: setup.From(big.NewInt(-4), numlimbs),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitTruthTableNegativePositive2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(-1), numlimbs*2),
			Q: setup.From(big.NewInt(-1), numlimbs),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv2QuoRemCircuitTruthTablePositiveNegative2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(1), numlimbs*2),
			Q: setup.From(big.NewInt(0), numlimbs),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitSmallDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			Q: setup.From(big.NewInt(30), 1),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitSmallDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(big.NewInt(120), numlimbs),
			Q: setup.From(big.NewInt(30), 1),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitSmallNonDivisibleCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs),
			Q: setup.From(big.NewInt(30), 1),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitSmallNonDivisibleIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&FloorDiv2QuoRemTestCircuit{
			A: setup.From(big.NewInt(121), numlimbs),
			Q: setup.From(big.NewInt(30), 1),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitBigDivisibleCorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(4), C)

	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitBigNegativeDivisibleCorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(-4), C)

	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C.Neg(C), numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitNegativeBothDivisible(t *testing.T) {
	C := commontest.BigInt2()
	C.Neg(C)
	A := new(big.Int).Mul(big.NewInt(4), C)

	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitBigNonDivisibleCorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(4), C)
	A.Add(A, big.NewInt(1))

	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitBigNonDivisibleIncorrect(t *testing.T) {
	C := commontest.BigInt2()
	A := new(big.Int).Mul(big.NewInt(4), C)
	A.Add(A, big.NewInt(1))

	commontest.TestCircuitInvalid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(A, numlimbs*2),
			Q: setup.From(C, numlimbs),
			R: 0,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitTruthTableBothPositive(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(big.NewInt(7), numlimbs*2),
			Q: setup.From(big.NewInt(1), numlimbs),
			R: 3,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitTruthTableNegativePositive1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(big.NewInt(-7), numlimbs*2),
			Q: setup.From(big.NewInt(-2), numlimbs),
			R: 1,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitTruthTableNegativePositive2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(big.NewInt(-1), numlimbs*2),
			Q: setup.From(big.NewInt(-1), numlimbs),
			R: 3,
		},
	)
}

func TestBigIntFloorDiv4QuoRemCircuitTruthTablePositiveNegative2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&FloorDiv4QuoRemTestCircuit{
			A: setup.From(big.NewInt(1), numlimbs*2),
			Q: setup.From(big.NewInt(0), numlimbs),
			R: 1,
		},
	)
}
