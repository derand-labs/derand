package vdfcircuit

import (
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/vdf"

	"github.com/consensys/gnark/frontend"
)

type VDFVerifierCircuit struct {
	XSeed frontend.Variable `gnark:",public"`
	Y     classgroup.Form   `gnark:",public"`
	Pi    classgroup.Form   `gnark:",public"`
	L     frontend.Variable `gnark:",public"`
	R     frontend.Variable `gnark:",public"`

	setup *vdf.Setup `gnark:"-"`
}

func NewVerifier(setup *vdf.Setup) *VDFVerifierCircuit {
	c := VDFVerifierCircuit{setup: setup}
	c.XSeed = frontend.Variable(0)
	c.Y = classgroup.Form{
		A: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		B: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		C: bigint.New(int(setup.Classgroup.DNumLimbs)),
	}
	c.Pi = classgroup.Form{
		A: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		B: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		C: bigint.New(int(setup.Classgroup.DNumLimbs)),
	}
	c.L = frontend.Variable(0)
	c.R = frontend.Variable(0)

	return &c
}

func (c *VDFVerifierCircuit) Assign(t *Transcript) {
	c.XSeed = t.XSeed
	c.Y = t.Y.ToZK(c.setup.Classgroup)
	c.Pi = t.Pi.ToZK(c.setup.Classgroup)
	c.L = t.L
	c.R = t.R
}

func (c *VDFVerifierCircuit) Define(api frontend.API) error {
	vdfapi := vdf.NewAPI(api, c.setup)

	vdfapi.ClassgroupAPI().AssertValid(c.Y, c.Pi)

	XForm := vdfapi.HashToForm(c.XSeed)
	vdfapi.AssertVerify(XForm, c.Y, c.Pi, api.ToBinary(c.L, c.setup.LBits), api.ToBinary(c.R, c.setup.LBits))

	return nil
}
