package vdf_test

import (
	"testing"
	"zkvdf/classgroup"
	"zkvdf/commontest"
	"zkvdf/utils"
	"zkvdf/vdf"

	"github.com/consensys/gnark/frontend"
)

type HashToFormTestCircuit struct {
	X   frontend.Variable `gnark:",public"`
	OUT classgroup.Form
}

func (c *HashToFormTestCircuit) Define(api frontend.API) error {
	vdfapi := vdf.NewAPI(api, setup)

	OUT := vdfapi.HashToForm(c.X)
	vdfapi.ClassgroupAPI().AssertIsEqual(OUT, c.OUT)
	return nil
}

func TestClassgroupHashToFormTestCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[frontend.Variable]{
			A: frontend.Variable(0),
			F: func(api frontend.API, a frontend.Variable) {
				setup := vdf.NewDummySetup(114, 1024, 128, 1, 26, 9097)
				vdf.NewAPI(api, setup).HashToForm(a)
			},
		},
		6413634,
	)
}

func TestVDFHashToFormCorrect(t *testing.T) {
	x := utils.BigIntFromString("1234567890abcdef1122334455667788deadbeefcafebabefeedface01234567", 16)

	aout := utils.BigIntFromString("12f3634bf997", 16)
	bout := utils.BigIntFromString("-a7c255a0aab", 16)
	cout := utils.BigIntFromString("3156e6b6ef015e2d381f61a9119dc9d351ebc0406bb145cb8621f4bddfe6fe04591cd0d98d3e35feaab01da3f2a3ef49fda35207b0ffd71c8f7c66db483b546734f27d374ec44b2ecf579674a112b9b960d66e1f3b5e04d2ae039ad65c66723b220d0ef5cf5e77ed37f7c1eea2c7b5e6f87772fbf12e3c82397b4", 16)

	commontest.TestCircuitValid(
		t,
		&HashToFormTestCircuit{
			X: x,
			OUT: classgroup.Form{
				A: setup.Classgroup.BigInt.From(aout, setup.Classgroup.GetSmallNumLimbs()),
				B: setup.Classgroup.BigInt.From(bout, setup.Classgroup.GetSmallNumLimbs()),
				C: setup.Classgroup.BigInt.From(cout, setup.Classgroup.DNumLimbs),
			},
		},
	)
}
