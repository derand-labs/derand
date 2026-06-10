package biguint_test

import (
	"math/big"
	"testing"
	"zkvdf/biguint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type RangeCheckTestCircuit struct {
	A biguint.BigUint `gnark:",public"`
}

func (circuit *RangeCheckTestCircuit) Define(api frontend.API) error {
	biguintapi := biguint.NewAPI(api, setup)
	biguintapi.AssertRangeCheck(circuit.A)
	return nil
}

func TestBigUintRangeCheckCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[biguint.BigUint]{
			A: biguint.New(numlimbs),
			F: func(api frontend.API, a biguint.BigUint) {
				biguint.NewAPI(api, setup).AssertRangeCheck(a)
			},
		},
		1199,
	)
}

func TestBigUintRangeCheckValidZero(t *testing.T) {
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(big.NewInt(0), numlimbs)})
}

func TestBigUintRangeCheckValidSmall(t *testing.T) {
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(big.NewInt(12), numlimbs)})
}

func TestBigUintRangeCheckValidBig(t *testing.T) {
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(commontest.BigInt1(), numlimbs)})
}

func TestBigUintRangeCheckValidMax(t *testing.T) {
	commontest.TestCircuitValid(t, &RangeCheckTestCircuit{A: setup.From(utils.MaxValue(numlimbs*setup.LimbBits), numlimbs)})
}

func TestBigUintRangeCheckInvalidLimb(t *testing.T) {
	invalidLimbA := setup.From(big.NewInt(12), numlimbs)
	invalidLimbA.Limbs[0] = new(big.Int).Lsh(big.NewInt(1), uint(setup.LimbBits+1))
	commontest.TestCircuitInvalid(t, &RangeCheckTestCircuit{A: invalidLimbA})
}
