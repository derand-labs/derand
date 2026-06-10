package vdfcircuit

import (
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/utils"
	"zkvdf/vdf"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/rangecheck"
	rcplonk "github.com/consensys/gnark/std/recursion/plonk"
)

type VDFRCVerifierPhase1Circuit struct {
	XSeed frontend.Variable `gnark:",public"`
	Pi    classgroup.Form   `gnark:",public"`

	PiLPhase1     classgroup.Form   `gnark:",public"`
	PiLBasePhase1 classgroup.Form   `gnark:",public"`
	XRPhase1      classgroup.Form   `gnark:",public"`
	XRBasePhase1  classgroup.Form   `gnark:",public"`
	LPhase1       frontend.Variable `gnark:",public"`
	RPhase1       frontend.Variable `gnark:",public"`

	HashToFormVK      RCVerifyingKey `gnark:"-"`
	HashToFormProof   RCProof
	HashToFormWitness RCWitness

	IntermediatePowVk RCVerifyingKey `gnark:"-"`

	PiLIntermediatePowProof   []RCProof
	PiLIntermediatePowWitness []RCWitness

	XRIntermediatePowProof   []RCProof
	XRIntermediatePowWitness []RCWitness

	setup *vdf.Setup `gnark:"-"`
}

func NewRCVerifierPhase1(
	setup *vdf.Setup,
	hashToFormCircuitSignature, intermediatePowCircuitSignature CircuitSignature,
) *VDFRCVerifierPhase1Circuit {
	if setup.SplitExp <= 1 {
		panic("invalid split-exp")
	}

	c := VDFRCVerifierPhase1Circuit{setup: setup}
	c.XSeed = frontend.Variable(0)
	c.Pi = classgroup.Form{
		A: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		B: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		C: bigint.New(int(setup.Classgroup.DNumLimbs)),
	}

	c.PiLPhase1 = classgroup.Form{
		A: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		B: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		C: bigint.New(int(setup.Classgroup.DNumLimbs)),
	}
	c.PiLBasePhase1 = classgroup.Form{
		A: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		B: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		C: bigint.New(int(setup.Classgroup.DNumLimbs)),
	}
	c.XRPhase1 = classgroup.Form{
		A: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		B: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		C: bigint.New(int(setup.Classgroup.DNumLimbs)),
	}
	c.XRBasePhase1 = classgroup.Form{
		A: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		B: bigint.New(c.setup.Classgroup.GetSmallNumLimbs()),
		C: bigint.New(int(setup.Classgroup.DNumLimbs)),
	}
	c.LPhase1 = frontend.Variable(0)
	c.RPhase1 = frontend.Variable(0)

	var err error
	c.HashToFormVK, err = ValueOfVerifyingKey(hashToFormCircuitSignature.VK)
	if err != nil {
		panic(err)
	}
	c.HashToFormProof = PlaceholderProof(hashToFormCircuitSignature.CCS)
	c.HashToFormWitness = PlaceholderWitness(hashToFormCircuitSignature.CCS)

	c.IntermediatePowVk, err = ValueOfVerifyingKey(intermediatePowCircuitSignature.VK)
	if err != nil {
		panic(err)
	}

	c.PiLIntermediatePowProof = make([]RCProof, c.setup.SplitExp/2)
	c.PiLIntermediatePowWitness = make([]RCWitness, c.setup.SplitExp/2)
	c.XRIntermediatePowProof = make([]RCProof, c.setup.SplitExp/2)
	c.XRIntermediatePowWitness = make([]RCWitness, c.setup.SplitExp/2)
	for i := range c.setup.SplitExp / 2 {
		c.PiLIntermediatePowProof[i] = PlaceholderProof(intermediatePowCircuitSignature.CCS)
		c.PiLIntermediatePowWitness[i] = PlaceholderWitness(intermediatePowCircuitSignature.CCS)
		c.XRIntermediatePowProof[i] = PlaceholderProof(intermediatePowCircuitSignature.CCS)
		c.XRIntermediatePowWitness[i] = PlaceholderWitness(intermediatePowCircuitSignature.CCS)
	}

	return &c
}

func (c *VDFRCVerifierPhase1Circuit) Assign(
	t *Transcript,
	hashToFormProof CircuitProof,
	pilIntermediatePowProof, xrIntermediatePowProof []CircuitProof,
) error {
	c.XSeed = t.XSeed
	c.Pi = t.Pi.ToZK(c.setup.Classgroup)

	c.PiLPhase1 = t.IntermediatePows[c.setup.SplitExp/2-1].PiL.ToZK(c.setup.Classgroup)
	c.PiLBasePhase1 = t.IntermediatePows[c.setup.SplitExp/2-1].PiLBase.ToZK(c.setup.Classgroup)
	c.XRPhase1 = t.IntermediatePows[c.setup.SplitExp/2-1].XR.ToZK(c.setup.Classgroup)
	c.XRBasePhase1 = t.IntermediatePows[c.setup.SplitExp/2-1].XRBase.ToZK(c.setup.Classgroup)
	c.LPhase1 = t.LPhase1
	c.RPhase1 = t.RPhase1

	var err error
	c.HashToFormProof, err = ValueOfProof(hashToFormProof.Proof)
	if err != nil {
		return err
	}

	c.HashToFormWitness, err = ValueOfWitness(hashToFormProof.Witness)
	if err != nil {
		return err
	}

	for i := range c.PiLIntermediatePowProof {
		c.PiLIntermediatePowProof[i], err = ValueOfProof(pilIntermediatePowProof[i].Proof)
		if err != nil {
			return err
		}
		c.PiLIntermediatePowWitness[i], err = ValueOfWitness(pilIntermediatePowProof[i].Witness)
		if err != nil {
			return err
		}

		c.XRIntermediatePowProof[i], err = ValueOfProof(xrIntermediatePowProof[i].Proof)
		if err != nil {
			return err
		}
		c.XRIntermediatePowWitness[i], err = ValueOfWitness(xrIntermediatePowProof[i].Witness)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *VDFRCVerifierPhase1Circuit) Define(api frontend.API) error {
	// Setup API
	vdfapi := vdf.NewAPI(api, c.setup)
	rcapi := rangecheck.New(api)

	// Check Input
	vdfapi.ClassgroupAPI().AssertValid(c.Pi)
	rcapi.Check(c.LPhase1, c.setup.LBits/2)
	rcapi.Check(c.RPhase1, c.setup.LBits/2)

	// Verify RC proof
	rcVerifier, err := rcplonk.NewVerifier[bn254Fr, bn254G1Aff, bn254G2Aff, bn254GT](api)
	if err != nil {
		return err
	}

	rcVerifier.AssertProof(c.HashToFormVK, c.HashToFormProof, c.HashToFormWitness, rcplonk.WithCompleteArithmetic())

	intermediatePowProofs := []RCProof{}
	intermediatePowProofs = append(intermediatePowProofs, c.PiLIntermediatePowProof...)
	intermediatePowProofs = append(intermediatePowProofs, c.XRIntermediatePowProof...)

	intermediatePowWitness := []RCWitness{}
	intermediatePowWitness = append(intermediatePowWitness, c.PiLIntermediatePowWitness...)
	intermediatePowWitness = append(intermediatePowWitness, c.XRIntermediatePowWitness...)

	rcVerifier.AssertSameProofs(c.IntermediatePowVk, intermediatePowProofs, intermediatePowWitness)

	// Check recursive public witness
	api.AssertIsEqual(c.XSeed, publicToVariable(api, c.HashToFormWitness.Public[0]))
	X, _ := publicsToForm(api, c.setup, c.HashToFormWitness.Public, 1)

	var xr, pil = vdfapi.ClassgroupAPI().GetPrincipalForm(), vdfapi.ClassgroupAPI().GetPrincipalForm()

	var xrBase, pilBase = X, c.Pi
	var accL, accR = frontend.Variable(0), frontend.Variable(0)
	for round := range c.PiLIntermediatePowWitness {
		// PiL
		lExp := publicToVariable(api, c.PiLIntermediatePowWitness[round].Public[0])

		offset := 1
		pilPrevValue, offset := publicsToForm(api, c.setup, c.PiLIntermediatePowWitness[round].Public, offset)
		pilPrevBase, offset := publicsToForm(api, c.setup, c.PiLIntermediatePowWitness[round].Public, offset)
		pilCurValue, offset := publicsToForm(api, c.setup, c.PiLIntermediatePowWitness[round].Public, offset)
		pilCurBase, _ := publicsToForm(api, c.setup, c.PiLIntermediatePowWitness[round].Public, offset)

		// XR
		rExp := publicToVariable(api, c.XRIntermediatePowWitness[round].Public[0])

		offset = 1
		xrPrevValue, offset := publicsToForm(api, c.setup, c.XRIntermediatePowWitness[round].Public, offset)
		xrPrevBase, offset := publicsToForm(api, c.setup, c.XRIntermediatePowWitness[round].Public, offset)
		xrCurValue, offset := publicsToForm(api, c.setup, c.XRIntermediatePowWitness[round].Public, offset)
		xrCurBase, _ := publicsToForm(api, c.setup, c.XRIntermediatePowWitness[round].Public, offset)

		vdfapi.ClassgroupAPI().AssertIsEqual(pilPrevValue, pil)
		vdfapi.ClassgroupAPI().AssertIsEqual(xrPrevValue, xr)

		vdfapi.ClassgroupAPI().AssertIsEqual(pilPrevBase, pilBase)
		vdfapi.ClassgroupAPI().AssertIsEqual(xrPrevBase, xrBase)

		pil = pilCurValue
		xr = xrCurValue

		pilBase = pilCurBase
		xrBase = xrCurBase

		expBitSize := c.setup.LBits / c.setup.SplitExp
		accL = api.Add(accL, api.Mul(lExp, utils.Modulus(expBitSize*round)))
		accR = api.Add(accR, api.Mul(rExp, utils.Modulus(expBitSize*round)))
	}

	api.AssertIsEqual(c.LPhase1, accL)
	api.AssertIsEqual(c.RPhase1, accR)

	vdfapi.ClassgroupAPI().AssertIsEqual(pil, c.PiLPhase1)
	vdfapi.ClassgroupAPI().AssertIsEqual(pilBase, c.PiLBasePhase1)
	vdfapi.ClassgroupAPI().AssertIsEqual(xr, c.XRPhase1)
	vdfapi.ClassgroupAPI().AssertIsEqual(xrBase, c.XRBasePhase1)

	return nil
}
