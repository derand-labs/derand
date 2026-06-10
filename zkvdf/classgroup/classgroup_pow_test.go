package classgroup_test

import (
	"testing"
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/commontest"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

type PowTestCircuit struct {
	F1  classgroup.Form   `gnark:",public"`
	EXP frontend.Variable `gnark:",public"`
	OUT classgroup.Form   `gnark:",public"`

	eBits int `gnark:"-"`
}

func (c *PowTestCircuit) Define(api frontend.API) error {
	cgapi := classgroup.NewAPI(api, setup)

	cgapi.AssertValid(c.F1, c.OUT)

	OUT := cgapi.Pow(c.F1, api.ToBinary(c.EXP, c.eBits))
	cgapi.AssertIsEqual(OUT, c.OUT)
	return nil
}

type PartialPowTestCircuit struct {
	PrevF, PrevBase classgroup.Form   `gnark:",public"`
	F, Base         classgroup.Form   `gnark:",public"`
	EXP             frontend.Variable `gnark:",public"`

	eBits int `gnark:"-"`
}

func (c *PartialPowTestCircuit) Define(api frontend.API) error {
	cgapi := classgroup.NewAPI(api, setup)

	cgapi.AssertValid(c.PrevF, c.PrevBase, c.F, c.Base)

	F, _ := cgapi.ParitalPow(c.PrevF, c.PrevBase, api.ToBinary(c.EXP, c.eBits))
	cgapi.AssertIsEqual(F, c.F)
	// cgapi.AssertIsEqual(Base, c.Base)
	return nil
}

func TestClassgroupPowCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[classgroup.Form]{
			A: classgroup.Form{
				A: bigint.New(setup.GetSmallNumLimbs()),
				B: bigint.New(setup.GetSmallNumLimbs()),
				C: bigint.New(setup.DNumLimbs),
			},
			F: func(api frontend.API, a classgroup.Form) {
				eBinary := make([]frontend.Variable, 16)
				for i := range eBinary {
					eBinary[i] = 0
				}
				classgroup.NewAPI(api, setup).Pow(a, eBinary)
			},
		},
		6283641,
	)
}

func TestClassgroupPowCorrect(t *testing.T) {
	a := utils.BigIntFromString("521639cd8c0f6f2168f9e37be67945ee819685e86c35a027f35f275cf0a79e812b00eac2b958cb8126abbd40231832d5189e6edd9ef3f26bb49b46c59e464f77", 16)
	b := utils.BigIntFromString("-f2532f84467be827f3432ca62d2ab29de1644627af062ad092eab9edaa631ba04426d23da1e7dbb08016a676a1109fcb5e3c89b31662545e125e9bf84d8dadd", 16)
	c := utils.BigIntFromString("b6f2c6ffc9285623d27a70d98e0ae08d3ec639fb25b1767d5ff23a6518ec00f1e4cd6007650a014fcf76ce5c7b3985635ccd81191c18613d29e8b04e0605f230", 16)

	aout := utils.BigIntFromString("588a07f93cb408fec4cdc92a873f7e5dc65a2c3950be69c7e807fb5b67a11229fbcccefe098c0312d3460532a3d9a4aaba43cb661d225caa28af6c9dc50d8e30", 16)
	bout := utils.BigIntFromString("-51c7f5bd9055abab7bce897726650e334fede86676a730dc0b7e4cdba7ca63b345060bb84a9a6e3046a812e291b7b2d2e47089d2723a09d57682439e1775cd5d", 16)
	cout := utils.BigIntFromString("bbda55a95663c694fd2580b6629839840551efb0d0fca9079742dcea03c00932b0641973f0ceea4637255277688c5176326fd1b2d0334ecd923f9e63a2971923", 16)

	commontest.TestCircuitValid(
		t,
		&PowTestCircuit{
			F1: classgroup.Form{
				A: setup.BigInt.From(a, setup.DNumLimbs/2),
				B: setup.BigInt.From(b, setup.DNumLimbs/2),
				C: setup.BigInt.From(c, setup.DNumLimbs),
			},
			EXP: 1000000,
			OUT: classgroup.Form{
				A: setup.BigInt.From(aout, setup.DNumLimbs/2),
				B: setup.BigInt.From(bout, setup.DNumLimbs/2),
				C: setup.BigInt.From(cout, setup.DNumLimbs),
			},
			eBits: 32,
		},
	)
}

func TestClassgroupPartialPowCorrect(t *testing.T) {
	pa, pb, pc := setup.GetPrincipalForm()

	a := utils.BigIntFromString("521639cd8c0f6f2168f9e37be67945ee819685e86c35a027f35f275cf0a79e812b00eac2b958cb8126abbd40231832d5189e6edd9ef3f26bb49b46c59e464f77", 16)
	b := utils.BigIntFromString("-f2532f84467be827f3432ca62d2ab29de1644627af062ad092eab9edaa631ba04426d23da1e7dbb08016a676a1109fcb5e3c89b31662545e125e9bf84d8dadd", 16)
	c := utils.BigIntFromString("b6f2c6ffc9285623d27a70d98e0ae08d3ec639fb25b1767d5ff23a6518ec00f1e4cd6007650a014fcf76ce5c7b3985635ccd81191c18613d29e8b04e0605f230", 16)

	aout := utils.BigIntFromString("588a07f93cb408fec4cdc92a873f7e5dc65a2c3950be69c7e807fb5b67a11229fbcccefe098c0312d3460532a3d9a4aaba43cb661d225caa28af6c9dc50d8e30", 16)
	bout := utils.BigIntFromString("-51c7f5bd9055abab7bce897726650e334fede86676a730dc0b7e4cdba7ca63b345060bb84a9a6e3046a812e291b7b2d2e47089d2723a09d57682439e1775cd5d", 16)
	cout := utils.BigIntFromString("bbda55a95663c694fd2580b6629839840551efb0d0fca9079742dcea03c00932b0641973f0ceea4637255277688c5176326fd1b2d0334ecd923f9e63a2971923", 16)

	commontest.TestCircuitValid(
		t,
		&PartialPowTestCircuit{
			PrevF: classgroup.Form{
				A: setup.BigInt.From(pa, setup.DNumLimbs/2),
				B: setup.BigInt.From(pb, setup.DNumLimbs/2),
				C: setup.BigInt.From(pc, setup.DNumLimbs),
			},
			PrevBase: classgroup.Form{
				A: setup.BigInt.From(a, setup.DNumLimbs/2),
				B: setup.BigInt.From(b, setup.DNumLimbs/2),
				C: setup.BigInt.From(c, setup.DNumLimbs),
			},
			EXP: 1000000,
			F: classgroup.Form{
				A: setup.BigInt.From(aout, setup.DNumLimbs/2),
				B: setup.BigInt.From(bout, setup.DNumLimbs/2),
				C: setup.BigInt.From(cout, setup.DNumLimbs),
			},
			Base: classgroup.Form{
				A: setup.BigInt.From(aout, setup.DNumLimbs/2),
				B: setup.BigInt.From(bout, setup.DNumLimbs/2),
				C: setup.BigInt.From(cout, setup.DNumLimbs),
			},
			eBits: 32,
		},
	)
}
