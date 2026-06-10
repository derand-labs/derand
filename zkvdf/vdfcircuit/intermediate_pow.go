package vdfcircuit

import (
	"math/big"
	"zkvdf/classgroup"
	"zkvdf/vdf"

	"github.com/consensys/gnark/frontend"
)

type VDFIntermediatePowCircuit struct {
	Exp frontend.Variable `gnark:",public"`

	PrevValue classgroup.Form `gnark:",public"`
	PrevBase  classgroup.Form `gnark:",public"`

	CurValue classgroup.Form `gnark:",public"`
	CurBase  classgroup.Form `gnark:",public"`

	setup *vdf.Setup `gnark:"-"`
}

func NewVDFIntermediatePow(setup *vdf.Setup) *VDFIntermediatePowCircuit {
	return &VDFIntermediatePowCircuit{
		setup: setup,
		Exp:   frontend.Variable(0),
		PrevValue: classgroup.Form{
			A: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			B: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			C: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.DNumLimbs),
		},
		PrevBase: classgroup.Form{
			A: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			B: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			C: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.DNumLimbs),
		},
		CurValue: classgroup.Form{
			A: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			B: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			C: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.DNumLimbs),
		},
		CurBase: classgroup.Form{
			A: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			B: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.GetSmallNumLimbs()),
			C: setup.Classgroup.BigInt.From(big.NewInt(0), setup.Classgroup.DNumLimbs),
		},
	}
}

func (c *VDFIntermediatePowCircuit) AssignPiL(t *Transcript, round int) {
	if round == 0 {
		pa, pb, pc := c.setup.Classgroup.GetPrincipalForm()

		c.PrevValue = classgroup.Form{
			A: c.setup.Classgroup.BigInt.From(pa, c.setup.Classgroup.GetSmallNumLimbs()),
			B: c.setup.Classgroup.BigInt.From(pb, c.setup.Classgroup.GetSmallNumLimbs()),
			C: c.setup.Classgroup.BigInt.From(pc, c.setup.Classgroup.DNumLimbs),
		}

		c.PrevBase = t.Pi.ToZK(c.setup.Classgroup)
	} else {
		c.PrevValue = t.IntermediatePows[round-1].PiL.ToZK(c.setup.Classgroup)
		c.PrevBase = t.IntermediatePows[round-1].PiLBase.ToZK(c.setup.Classgroup)
	}

	c.CurValue = t.IntermediatePows[round].PiL.ToZK(c.setup.Classgroup)
	c.CurBase = t.IntermediatePows[round].PiLBase.ToZK(c.setup.Classgroup)
	c.Exp = t.IntermediatePows[round].L
}

func (c *VDFIntermediatePowCircuit) AssignXR(t *Transcript, round int) {
	if round == 0 {
		pa, pb, pc := c.setup.Classgroup.GetPrincipalForm()

		c.PrevValue = classgroup.Form{
			A: c.setup.Classgroup.BigInt.From(pa, c.setup.Classgroup.GetSmallNumLimbs()),
			B: c.setup.Classgroup.BigInt.From(pb, c.setup.Classgroup.GetSmallNumLimbs()),
			C: c.setup.Classgroup.BigInt.From(pc, c.setup.Classgroup.DNumLimbs),
		}

		c.PrevBase = t.X.ToZK(c.setup.Classgroup)
	} else {
		c.PrevValue = t.IntermediatePows[round-1].XR.ToZK(c.setup.Classgroup)
		c.PrevBase = t.IntermediatePows[round-1].XRBase.ToZK(c.setup.Classgroup)
	}

	c.CurValue = t.IntermediatePows[round].XR.ToZK(c.setup.Classgroup)
	c.CurBase = t.IntermediatePows[round].XRBase.ToZK(c.setup.Classgroup)
	c.Exp = t.IntermediatePows[round].R
}

func (c *VDFIntermediatePowCircuit) Define(api frontend.API) error {
	vdfapi := vdf.NewAPI(api, c.setup)
	expBinary := api.ToBinary(c.Exp, c.setup.LBits/c.setup.SplitExp)

	vdfapi.ClassgroupAPI().AssertValid(c.PrevValue, c.PrevBase, c.CurValue, c.CurBase)

	value, base := vdfapi.ClassgroupAPI().ParitalPow(c.PrevValue, c.PrevBase, expBinary)

	vdfapi.ClassgroupAPI().AssertIsEqual(value, c.CurValue)
	vdfapi.ClassgroupAPI().AssertIsEqual(base, c.CurBase)

	return nil
}
