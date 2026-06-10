package biguint_test

import (
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type CastTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
	B biguint.BigUint `gnark:",public"`
}

func (circuit *CastTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A, circuit.B)

	B := biguintapi.Cast(circuit.A, len(circuit.B.Limbs))
	biguintapi.AssertIsEqual(B, circuit.B)

	return nil
}

func TestBigUintMinimalCastCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs * 2),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).Cast(a, numlimbs)
			},
		},
		16,
	)
}

func TestBigUintCastCircuitCorrect(t *testing.T) {
	commontest.TestCircuitValid(
		t,
		&CastTestCircuit{
			A: setup.From(utils.MaxValue(numlimbs*setup.LimbBits), numlimbs*2),
			B: setup.From(utils.MaxValue(numlimbs*setup.LimbBits), numlimbs),
		},
	)
}

func TestBigUintCastCircuitIncorrect(t *testing.T) {
	commontest.TestCircuitInvalid(
		t,
		&CastTestCircuit{
			A: setup.From(utils.MaxValue(numlimbs*2*setup.LimbBits), numlimbs*2),
			B: setup.From(utils.MaxValue(numlimbs*setup.LimbBits), numlimbs),
		},
	)
}
