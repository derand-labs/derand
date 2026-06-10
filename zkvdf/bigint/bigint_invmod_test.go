package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

type InvModTestCircuit struct {
	A   bigint.BigInt `gnark:",public"`
	B   bigint.BigInt `gnark:",public"`
	INV bigint.BigInt `gnark:",public"`
}

func (circuit *InvModTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B, circuit.INV)

	INV := bigintapi.InvMod(circuit.A, circuit.B)
	bigintapi.AssertIsEqual(INV, circuit.INV)

	return nil
}

func TestBigIntInvModCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit2[bigint.BigInt]{
			A: bigint.New(numlimbs),
			B: bigint.New(numlimbs),
			F: func(api frontend.API, a, b bigint.BigInt) {
				bigint.NewAPI(api, setup).InvMod(a, b)
			},
		},
		19367,
	)
}

func TestBigIntInvModCircuit_Small_Correct(t *testing.T) {
	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(12), numlimbs),
		B:   setup.From(big.NewInt(17), numlimbs),
		INV: setup.From(big.NewInt(10), numlimbs),
	})
}

func TestBigIntInvModCircuit_Medium_Correct(t *testing.T) {
	a := big.NewInt(123456789)
	m := big.NewInt(987654323)
	inv := new(big.Int).ModInverse(a, m)

	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(a, numlimbs*2),
		B:   setup.From(m, numlimbs*2),
		INV: setup.From(inv, numlimbs*2),
	})
}

func TestBigIntInvModCircuit_A_IsOne_Correct(t *testing.T) {
	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(1), numlimbs),
		B:   setup.From(big.NewInt(17), numlimbs),
		INV: setup.From(big.NewInt(1), numlimbs),
	})
}

func TestBigIntInvModCircuit_Large_Correct(t *testing.T) {
	a := big.NewInt(0)
	a.SetString("123456789012345678901234567890123456789", 10)
	m := big.NewInt(0)
	m.SetString("100000000000000000000000000000000000003", 10)
	inv := new(big.Int).ModInverse(a, m)

	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(a, 64),
		B:   setup.From(m, 64),
		INV: setup.From(inv, 64),
	})
}

func TestBigIntInvModCircuit_WrongInverse_Invalid(t *testing.T) {
	commontest.TestCircuitInvalid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(12), numlimbs),
		B:   setup.From(big.NewInt(17), numlimbs),
		INV: setup.From(big.NewInt(5), numlimbs),
	})
}

func TestBigIntInvModCircuit_NoInverse_GCD_GT1_Invalid(t *testing.T) {
	commontest.TestCircuitInvalid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(12), numlimbs),
		B:   setup.From(big.NewInt(18), numlimbs),
		INV: setup.From(big.NewInt(0), numlimbs),
	})
}

func TestBigIntInvModCircuit_A_IsZero_Invalid(t *testing.T) {
	commontest.TestCircuitInvalid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(0), numlimbs),
		B:   setup.From(big.NewInt(17), numlimbs),
		INV: setup.From(big.NewInt(1), numlimbs),
	})
}

func TestBigIntInvModCircuit_B_IsOne_Invalid(t *testing.T) {
	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(5), numlimbs),
		B:   setup.From(big.NewInt(1), numlimbs),
		INV: setup.From(big.NewInt(0), numlimbs),
	})
}

func TestBigIntInvModCircuit_NegativeA_Correct(t *testing.T) {
	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(-12), numlimbs),
		B:   setup.From(big.NewInt(17), numlimbs),
		INV: setup.From(big.NewInt(7), numlimbs),
	})
}

func TestBigIntInvModCircuit_NegativeA_Medium_Correct(t *testing.T) {
	a := big.NewInt(-123456789)
	m := big.NewInt(987654323)

	inv := new(big.Int).ModInverse(a, m)
	if inv == nil {
		t.Fatal("No inverse")
	}

	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(a, numlimbs*2),
		B:   setup.From(m, numlimbs*2),
		INV: setup.From(inv, numlimbs*2),
	})
}

func TestBigIntInvModCircuit_NegativeA_Large_Correct(t *testing.T) {
	a := big.NewInt(0)
	a.SetString("-123456789012345678901234567890123456789", 10)
	m := big.NewInt(0)
	m.SetString("100000000000000000000000000000000000003", 10)

	inv := new(big.Int).ModInverse(a, m)

	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(a, numlimbs),
		B:   setup.From(m, numlimbs),
		INV: setup.From(inv, numlimbs),
	})
}

func TestBigIntInvModCircuit_BothNegative_Correct(t *testing.T) {
	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(-12), numlimbs),
		B:   setup.From(big.NewInt(17), numlimbs),
		INV: setup.From(big.NewInt(7), numlimbs),
	})
}

func TestBigIntInvModCircuit_Done(t *testing.T) {
	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(-12), numlimbs),
		B:   setup.From(big.NewInt(1), numlimbs),
		INV: setup.From(big.NewInt(0), numlimbs),
	})
}

func TestBigIntInvModCircuit_2(t *testing.T) {
	commontest.TestCircuitValid(t, &InvModTestCircuit{
		A:   setup.From(big.NewInt(1), numlimbs),
		B:   setup.From(big.NewInt(1), numlimbs),
		INV: setup.From(big.NewInt(0), numlimbs),
	})
}
