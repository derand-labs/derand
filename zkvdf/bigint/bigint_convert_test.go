package bigint_test

import (
	"math/big"
	"testing"
	"zkvdf/bigint"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type CastTestCircuit struct {
	A bigint.BigInt `gnark:",public"`
	B bigint.BigInt `gnark:",public"`
}

func (circuit *CastTestCircuit) Define(api frontend.API) error {
	bigintapi := bigint.NewAPI(api, setup)

	bigintapi.AssertRangeCheck(circuit.A, circuit.B)

	B := bigintapi.Cast(circuit.A, len(circuit.B.Mag.Limbs))
	bigintapi.AssertIsEqual(B, circuit.B)

	return nil
}

func TestBigIntCastCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[bigint.BigInt]{
			A: bigint.New(numlimbs * 2),
			F: func(api frontend.API, a bigint.BigInt) {
				bigint.NewAPI(api, setup).Cast(a, numlimbs)
			},
		},
		16,
	)
}

func TestBigIntCastCircuitMaxCorrect(t *testing.T) {
	Max := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)

	commontest.TestCircuitValid(
		t,
		&CastTestCircuit{
			A: setup.From(Max, numlimbs*2),
			B: setup.From(Max, numlimbs),
		},
	)
}

func TestBigIntCastCircuitMinCorrect(t *testing.T) {
	Max := utils.MaxValue(numlimbs * setup.BigUint.LimbBits)
	Min := new(big.Int).Neg(Max)

	commontest.TestCircuitValid(
		t,
		&CastTestCircuit{
			A: setup.From(Min, numlimbs*2),
			B: setup.From(Min, numlimbs),
		},
	)
}

func TestBigIntCastCircuitMaxIncorrect(t *testing.T) {
	Max := utils.MaxValue(17 * setup.BigUint.LimbBits)

	commontest.TestCircuitInvalid(
		t,
		&CastTestCircuit{
			A: setup.From(Max, limbbits*2),
			B: setup.FromUnsafe(Max, numlimbs),
		},
	)
}
