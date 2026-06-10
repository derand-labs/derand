package classgroup_test

import (
	"testing"
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type SquareTestCircuit struct {
	F1 classgroup.Form `gnark:",public"`
}

func (c *SquareTestCircuit) Define(api frontend.API) error {
	cgapi := classgroup.NewAPI(api, setup)

	cgapi.AssertValid(c.F1)

	OUT := cgapi.Square(c.F1)
	cgapi.AssertIsEqual(OUT, cgapi.Compose(c.F1, c.F1))
	return nil
}

func TestClassgroupSquareCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[classgroup.Form]{
			A: classgroup.Form{
				A: bigint.New(setup.GetSmallNumLimbs()),
				B: bigint.New(setup.GetSmallNumLimbs()),
				C: bigint.New(setup.DNumLimbs),
			},
			F: func(api frontend.API, a classgroup.Form) {
				classgroup.NewAPI(api, setup).Square(a)
			},
		},
		116805,
	)
}

func TestClassgroupSquareCorrect(t *testing.T) {
	a1 := utils.BigIntFromString("521639cd8c0f6f2168f9e37be67945ee819685e86c35a027f35f275cf0a79e812b00eac2b958cb8126abbd40231832d5189e6edd9ef3f26bb49b46c59e464f77", 16)
	b1 := utils.BigIntFromString("-f2532f84467be827f3432ca62d2ab29de1644627af062ad092eab9edaa631ba04426d23da1e7dbb08016a676a1109fcb5e3c89b31662545e125e9bf84d8dadd", 16)
	c1 := utils.BigIntFromString("b6f2c6ffc9285623d27a70d98e0ae08d3ec639fb25b1767d5ff23a6518ec00f1e4cd6007650a014fcf76ce5c7b3985635ccd81191c18613d29e8b04e0605f230", 16)

	commontest.TestCircuitValid(
		t,
		&SquareTestCircuit{
			F1: classgroup.Form{
				A: setup.BigInt.From(a1, setup.DNumLimbs/2),
				B: setup.BigInt.From(b1, setup.DNumLimbs/2),
				C: setup.BigInt.From(c1, setup.DNumLimbs),
			},
		},
	)
}
