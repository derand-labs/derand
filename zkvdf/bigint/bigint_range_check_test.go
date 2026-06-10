package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type RangeCheckTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
}

func (circuit *RangeCheckTestCircuit) Define(api frontend.API) error {
	bigint.NewAPI(api, setup).AssertRangeCheck(circuit.A)
	return nil
}

func TestBigIntRangeCheckCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).AssertRangeCheck(a)
			},
		},
		1224,
	)
}

func TestBigIntRangeCheckSmallValid(t *testing.T) {
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(big.NewInt(12), numlimbs)})
}

func TestBigIntRangeCheckValidMax(t *testing.T) {
	Max := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(Max, numlimbs)})
}

func TestBigIntRangeCheckValidMin(t *testing.T) {
	Max := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	Min := new(big.Int).Add(new(big.Int).Neg(Max), big.NewInt(1))
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(Min, numlimbs)})
}

func TestBigIntRangeCheckValidMinAdd1(t *testing.T) {
	Max := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	Min := new(big.Int).Add(new(big.Int).Neg(Max), big.NewInt(1))
	C := new(big.Int).Add(Min, big.NewInt(1))
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(C, numlimbs)})
}

func TestBigIntRangeCheckInvalidOverflowMax(t *testing.T) {
	a := utils.Modulus(numlimbs * setup.BigUint.LimbBits)
	commontest.TestCircuitInvalid(t, &RangeCheckTestCircuit{A: setup.FromUnsafe(a, numlimbs)})
}

func TestBigIntRangeCheckInvalidOverflowMin(t *testing.T) {
	b := utils.Modulus(numlimbs * setup.BigUint.LimbBits)
	b.Neg(b)
	commontest.TestCircuitInvalid(t, &RangeCheckTestCircuit{A: setup.FromUnsafe(b, numlimbs)})
}

func TestBigIntRangeCheckInvalidSignZeroNegative(t *testing.T) {
	zero1 := bigint.New(numlimbs)
	zero1.Sign = -1
	commontest.TestCircuitInvalid(t, &RangeCheckTestCircuit{A: zero1})
}

func TestBigIntRangeCheckInvalidSignZeroPositive(t *testing.T) {
	zero2 := bigint.New(numlimbs)
	zero2.Sign = 1
	commontest.TestCircuitInvalid(t, &RangeCheckTestCircuit{A: zero2})
}

func TestBigIntRangeCheckInvalidSignPositiveZero(t *testing.T) {
	a := commontest.BigInt1()
	A := setup.From(a, numlimbs)
	A.Sign = 0
	commontest.TestCircuitInvalid(t, &RangeCheckTestCircuit{A: A})
}

func TestBigIntRangeCheckInvalidSignNegativeZero(t *testing.T) {
	b := commontest.BigInt2()
	b.Neg(b)
	B := setup.From(b, numlimbs)
	B.Sign = 0
	commontest.TestCircuitInvalid(t, &RangeCheckTestCircuit{A: B})
}
