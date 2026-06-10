package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type NegTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
	B biguint.BigUint `gnark:",public"`
}

func (circuit *NegTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)

	biguintapi.AssertRangeCheck(circuit.A, circuit.B)

	B := biguintapi.Neg(circuit.A)
	biguintapi.AssertIsEqual(B, circuit.B)

	return nil
}

func TestBigUintNegCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).Neg(a)
			},
		},
		112,
	)
}

func TestBigUintNegCircuitSmallCorrect1(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&NegTestCircuit{
			A: setup.From(big.NewInt(12), numlimbs),
			B: setup.From(new(big.Int).Sub(utils.Modulus(numlimbs*setup.LimbBits), big.NewInt(12)), numlimbs),
		},
	)
}

func TestBigUintNegCircuitSmallCorrect2(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&NegTestCircuit{
			A: setup.From(new(big.Int).Sub(utils.Modulus(numlimbs*setup.LimbBits), big.NewInt(12)), numlimbs),
			B: setup.From(big.NewInt(12), numlimbs),
		},
	)
}

func TestBigUintNegCircuitSmallZero(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&NegTestCircuit{
			A: setup.From(big.NewInt(0), 1),
			B: setup.From(big.NewInt(0), 1),
		},
	)
}

func TestBigUintNegCircuitSmallIncorrect1(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&NegTestCircuit{
			A: setup.From(big.NewInt(12), 1),
			B: setup.FromUnsafe(new(big.Int).Sub(big.NewInt(12), utils.Modulus(setup.LimbBits)), 1),
		},
	)
}

func TestBigUintNegCircuitSmallIncorrect2(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&NegTestCircuit{
			A: setup.From(big.NewInt(12), 1),
			B: setup.From(big.NewInt(12), 1),
		},
	)
}
