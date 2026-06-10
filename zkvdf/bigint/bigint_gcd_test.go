package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type GCDTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
	G bigint.BigInt `gnark:",public"`
}

func (circuit *GCDTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.G)

	G, _, _, _, _ := bigintapi.GCD(circuit.A, circuit.B)
	bigintapi.AssertIsEqual(G, circuit.G)

	return nil
}

func TestBigIntGCDCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).GCD(a, b)
			},
		},
		21709,
	)
}

func TestBigIntGCDCircuitSmallCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			G: setup.From(big.NewInt(2), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitSmallCoprime(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(13), numlimbs),
			B: setup.From(big.NewInt(17), numlimbs),
			G: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitOneDividesOther(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(21), numlimbs),
			B: setup.From(big.NewInt(7), numlimbs),
			G: setup.From(big.NewInt(7), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitSwappedInputs(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(10), numlimbs),
			B: setup.From(big.NewInt(12), numlimbs),
			G: setup.From(big.NewInt(2), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitZeroAndNonZeroAZero(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(0), numlimbs),
			B: setup.From(big.NewInt(42), numlimbs),
			G: setup.From(big.NewInt(42), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitZeroAndNonZeroBZero(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(42), numlimbs),
			B: setup.From(big.NewInt(0), numlimbs),
			G: setup.From(big.NewInt(42), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitBothZero(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(0), numlimbs),
			B: setup.From(big.NewInt(0), numlimbs),
			G: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitLargeComposite(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(360), numlimbs),
			B: setup.From(big.NewInt(840), numlimbs),
			G: setup.From(big.NewInt(120), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitPowerOfTwo(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(1<<20), numlimbs),
			B: setup.From(big.NewInt(1<<12), numlimbs),
			G: setup.From(big.NewInt(1<<12), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitBigNumbers(t *testing.T) {
	a := utils.BigIntFromString("123456789012345678901234567890", 10)
	b := utils.BigIntFromString("987654321098765432109876543210", 10)
	g := new(big.Int).GCD(new(big.Int), new(big.Int), a, b)

	commontest.TestCircuitValid(
		t,
		&GCDTestCircuit{
			A: setup.From(a, numlimbs),
			B: setup.From(b, numlimbs),
			G: setup.From(g, numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidWrongTooSmallG(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			G: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidWrongTooLargeG(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			G: setup.From(big.NewInt(4), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidWrongCoprimeG(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(21), numlimbs),
			B: setup.From(big.NewInt(14), numlimbs),
			G: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidWrongZeroCaseAZero(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(0), numlimbs),
			B: setup.From(big.NewInt(42), numlimbs),
			G: setup.From(big.NewInt(41), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidWrongZeroCaseBZero(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(42), numlimbs),
			B: setup.From(big.NewInt(0), numlimbs),
			G: setup.From(big.NewInt(41), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidBothZeroNonZeroG(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(0), numlimbs),
			B: setup.From(big.NewInt(0), numlimbs),
			G: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidWrongLargeComposite(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(360), numlimbs),
			B: setup.From(big.NewInt(840), numlimbs),
			G: setup.From(big.NewInt(60), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidCommonDivisorButNotGreatest(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(24), numlimbs),
			B: setup.From(big.NewInt(36), numlimbs),
			G: setup.From(big.NewInt(6), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidGDividesAOnly(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(20), numlimbs),
			B: setup.From(big.NewInt(15), numlimbs),
			G: setup.From(big.NewInt(4), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidGDividesBOnly(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(20), numlimbs),
			B: setup.From(big.NewInt(15), numlimbs),
			G: setup.From(big.NewInt(3), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidGIsZeroForNonZeroInputs(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(10), numlimbs),
			B: setup.From(big.NewInt(20), numlimbs),
			G: setup.From(big.NewInt(0), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidGGreaterThanInputs(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(5), numlimbs),
			B: setup.From(big.NewInt(10), numlimbs),
			G: setup.From(big.NewInt(20), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidNegativeInputsConceptual(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(7), numlimbs),
			B: setup.From(big.NewInt(13), numlimbs),
			G: setup.From(big.NewInt(7), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidLargeRandomG(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(100), numlimbs),
			B: setup.From(big.NewInt(200), numlimbs),
			G: setup.From(big.NewInt(33), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidOneZeroOneNonZeroWrongG(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(0), numlimbs),
			B: setup.From(big.NewInt(50), numlimbs),
			G: setup.From(big.NewInt(25), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidGIsMultipleOfTrueGCD(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(big.NewInt(18), numlimbs),
			G: setup.From(big.NewInt(12), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidSwapGWithInput(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(14), numlimbs),
			B: setup.From(big.NewInt(28), numlimbs),
			G: setup.From(big.NewInt(28), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidNonZeroInputButGIsOne(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(100), numlimbs),
			B: setup.From(big.NewInt(50), numlimbs),
			G: setup.From(big.NewInt(1), numlimbs),
		},
	)
}

func TestBigIntGCDCircuitInvalidVeryLargeLimbsG(t *testing.T) {
	largeValue := new(big.Int).Lsh(big.NewInt(1), 256)
	commontest.TestCircuitInvalid(
		t,
		&GCDTestCircuit{
			A: setup.From(big.NewInt(10), numlimbs),
			B: setup.From(big.NewInt(20), numlimbs),
			G: setup.From(largeValue, numlimbs),
		},
	)
}
