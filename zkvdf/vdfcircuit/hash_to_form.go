package vdfcircuit

import (
	"math/big"
	"zkvdf/classgroup"
	"zkvdf/vdf"

	"github.com/consensys/gnark/frontend"
)

type VDFHashToFormCircuit struct {
	XSeed frontend.Variable `gnark:",public"`
	XForm classgroup.Form   `gnark:",public"`

	setup *vdf.Setup `gnark:"-"`
}

func NewVDFHashToFormCircuit(setup *vdf.Setup) *VDFHashToFormCircuit {
	return &VDFHashToFormCircuit{
		setup: setup,
		XSeed: frontend.Variable(0),
		XForm: classgroup.Form{
			A: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			B: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			C: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.DNumLimbs),
		},
	}
}

func (c *VDFHashToFormCircuit) Assign(t *Transcript) {
	c.XSeed = t.XSeed
	c.XForm = t.X.ToZK(c.setup.Classgroup)
}

func (c *VDFHashToFormCircuit) Define(api frontend.API) error {
	vdfapi := vdf.NewAPI(api, c.setup)

	XForm := vdfapi.HashToForm(c.XSeed)
	vdfapi.ClassgroupAPI().AssertIsEqual(XForm, c.XForm)

	return nil
}
