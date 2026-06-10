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

type VDFRCVerifierPhase2Circuit struct {
	XSeed frontend.Variable `gnark:",public"`
	Y     classgroup.Form   `gnark:",public"`
	Pi    classgroup.Form   `gnark:",public"`
	L     frontend.Variable `gnark:",public"`
	R     frontend.Variable `gnark:",public"`

	Phase1VK      RCVerifyingKey `gnark:"-"`
	Phase1Proof   RCProof
	Phase1Witness RCWitness

	IntermediatePowVk RCVerifyingKey `gnark:"-"`

	PiLIntermediatePowProof   []RCProof
	PiLIntermediatePowWitness []RCWitness

	XRIntermediatePowProof   []RCProof
	XRIntermediatePowWitness []RCWitness

	setup *vdf.Setup `gnark:"-"`
}

func NewRCVerifierPhase2(
	setup *vdf.Setup,
	phase1CircuitSignature, intermediatePowCircuitSignature CircuitSignature,
) *VDFRCVerifierPhase2Circuit {
	if setup.SplitExp <= 1 {
		panic("invalid split-exp")
	}

	c := VDFRCVerifierPhase2Circuit{setup: setup}
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

	var err error
	c.Phase1VK, err = ValueOfVerifyingKey(phase1CircuitSignature.VK)
	if err != nil {
		panic(err)
	}
	c.Phase1Proof = PlaceholderProof(phase1CircuitSignature.CCS)
	c.Phase1Witness = PlaceholderWitness(phase1CircuitSignature.CCS)

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

func (c *VDFRCVerifierPhase2Circuit) Assign(
	t *Transcript,
	phase1Proof CircuitProof,
	pilIntermediatePowProof, xrIntermediatePowProof []CircuitProof,
) error {
	c.XSeed = t.XSeed
	c.Y = t.Y.ToZK(c.setup.Classgroup)
	c.Pi = t.Pi.ToZK(c.setup.Classgroup)
	c.L = t.L
	c.R = t.R

	var err error
	c.Phase1Proof, err = ValueOfProof(phase1Proof.Proof)
	if err != nil {
		return err
	}

	c.Phase1Witness, err = ValueOfWitness(phase1Proof.Witness)
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

func (c *VDFRCVerifierPhase2Circuit) Define(api frontend.API) error {
	// Setup API
	vdfapi := vdf.NewAPI(api, c.setup)
	rcapi := rangecheck.New(api)

	// Check Input
	vdfapi.ClassgroupAPI().AssertValid(c.Y, c.Pi)
	rcapi.Check(c.L, c.setup.LBits)
	rcapi.Check(c.R, c.setup.LBits)

	// Verify RC proof
	rcVerifier, err := rcplonk.NewVerifier[bn254Fr, bn254G1Aff, bn254G2Aff, bn254GT](api)
	if err != nil {
		return err
	}

	rcVerifier.AssertProof(c.Phase1VK, c.Phase1Proof, c.Phase1Witness)

	intermediatePowProofs := []RCProof{}
	intermediatePowProofs = append(intermediatePowProofs, c.PiLIntermediatePowProof...)
	intermediatePowProofs = append(intermediatePowProofs, c.XRIntermediatePowProof...)

	intermediatePowWitness := []RCWitness{}
	intermediatePowWitness = append(intermediatePowWitness, c.PiLIntermediatePowWitness...)
	intermediatePowWitness = append(intermediatePowWitness, c.XRIntermediatePowWitness...)

	rcVerifier.AssertSameProofs(c.IntermediatePowVk, intermediatePowProofs, intermediatePowWitness)

	// Check recursive public witness
	api.AssertIsEqual(c.XSeed, publicToVariable(api, c.Phase1Witness.Public[0]))
	Phase1Pi, offset := publicsToForm(api, c.setup, c.Phase1Witness.Public, 1)
	Phase1PiL, offset := publicsToForm(api, c.setup, c.Phase1Witness.Public, offset)
	Phase1PilBase, offset := publicsToForm(api, c.setup, c.Phase1Witness.Public, offset)
	Phase1XR, offset := publicsToForm(api, c.setup, c.Phase1Witness.Public, offset)
	Phase1XRBase, offset := publicsToForm(api, c.setup, c.Phase1Witness.Public, offset)
	Phase1L := publicToVariable(api, c.Phase1Witness.Public[offset])
	Phase1R := publicToVariable(api, c.Phase1Witness.Public[offset+1])

	vdfapi.ClassgroupAPI().AssertIsEqual(c.Pi, Phase1Pi)

	var pil, xr = Phase1PiL, Phase1XR
	var pilBase, xrBase = Phase1PilBase, Phase1XRBase
	var accL, accR = Phase1L, Phase1R
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
		m := utils.Modulus(expBitSize * (round + c.setup.SplitExp/2))
		accL = api.Add(accL, api.Mul(lExp, m))
		accR = api.Add(accR, api.Mul(rExp, m))
	}

	api.AssertIsEqual(c.L, accL)
	api.AssertIsEqual(c.R, accR)

	yhat := vdfapi.ClassgroupAPI().Compose(pil, xr)
	yhat = vdfapi.ClassgroupAPI().Reduce(yhat)
	vdfapi.ClassgroupAPI().AssertIsEqual(c.Y, yhat)

	return nil
}
