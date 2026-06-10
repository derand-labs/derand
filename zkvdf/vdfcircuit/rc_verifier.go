package vdfcircuit

import (
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/utils"
	"zkvdf/vdf"

	beplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bn254"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/rangecheck"
	rcplonk "github.com/consensys/gnark/std/recursion/plonk"
)

type (
	bn254Fr    = sw_bn254.ScalarField
	bn254G1Aff = sw_bn254.G1Affine
	bn254G2Aff = sw_bn254.G2Affine
	bn254GT    = sw_bn254.GTEl
)

var (
	ValueOfVerifyingKey = rcplonk.ValueOfVerifyingKey[bn254Fr, bn254G1Aff, bn254G2Aff]
	ValueOfProof        = rcplonk.ValueOfProof[bn254Fr, bn254G1Aff, bn254G2Aff]
	ValueOfWitness      = rcplonk.ValueOfWitness[bn254Fr]

	PlaceholderProof   = rcplonk.PlaceholderProof[bn254Fr, bn254G1Aff, bn254G2Aff]
	PlaceholderWitness = rcplonk.PlaceholderWitness[bn254Fr]
)

type (
	RCVerifyingKey = rcplonk.VerifyingKey[bn254Fr, bn254G1Aff, bn254G2Aff]
	RCProof        = rcplonk.Proof[bn254Fr, bn254G1Aff, bn254G2Aff]
	RCWitness      = rcplonk.Witness[bn254Fr]
)

type CircuitSignature struct {
	VK  beplonk.VerifyingKey
	CCS constraint.ConstraintSystem
}

type CircuitProof struct {
	Proof   beplonk.Proof
	Witness witness.Witness
}

type VDFRCVerifierCircuit struct {
	XSeed frontend.Variable `gnark:",public"`
	Y     classgroup.Form   `gnark:",public"`
	Pi    classgroup.Form   `gnark:",public"`
	L     frontend.Variable `gnark:",public"`
	R     frontend.Variable `gnark:",public"`

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

func NewRCVerifier(
	setup *vdf.Setup,
	hashToFormCircuitSignature, intermediatePowCircuitSignature CircuitSignature,
) *VDFRCVerifierCircuit {
	c := VDFRCVerifierCircuit{setup: setup}
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

	c.PiLIntermediatePowProof = make([]RCProof, c.setup.SplitExp)
	c.PiLIntermediatePowWitness = make([]RCWitness, c.setup.SplitExp)
	c.XRIntermediatePowProof = make([]RCProof, c.setup.SplitExp)
	c.XRIntermediatePowWitness = make([]RCWitness, c.setup.SplitExp)
	for i := range c.setup.SplitExp {
		c.PiLIntermediatePowProof[i] = PlaceholderProof(intermediatePowCircuitSignature.CCS)
		c.PiLIntermediatePowWitness[i] = PlaceholderWitness(intermediatePowCircuitSignature.CCS)
		c.XRIntermediatePowProof[i] = PlaceholderProof(intermediatePowCircuitSignature.CCS)
		c.XRIntermediatePowWitness[i] = PlaceholderWitness(intermediatePowCircuitSignature.CCS)
	}

	return &c
}

func (c *VDFRCVerifierCircuit) Assign(
	t *Transcript,
	hashToFormProof CircuitProof,
	pilIntermediatePowProof, xrIntermediatePowProof []CircuitProof,
) error {
	c.XSeed = t.XSeed
	c.Y = t.Y.ToZK(c.setup.Classgroup)
	c.Pi = t.Pi.ToZK(c.setup.Classgroup)
	c.L = t.L
	c.R = t.R

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

func (c *VDFRCVerifierCircuit) Define(api frontend.API) error {
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

	rcVerifier.AssertProof(c.HashToFormVK, c.HashToFormProof, c.HashToFormWitness)

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

	api.AssertIsEqual(c.L, accL)
	api.AssertIsEqual(c.R, accR)

	yhat := vdfapi.ClassgroupAPI().Compose(pil, xr)
	yhat = vdfapi.ClassgroupAPI().Reduce(yhat)
	vdfapi.ClassgroupAPI().AssertIsEqual(c.Y, yhat)

	return nil
}

func publicToVariable(api frontend.API, p emulated.Element[bn254Fr]) frontend.Variable {
	result := frontend.Variable(0)
	base := utils.Modulus(64)

	for i := len(p.Limbs) - 1; i >= 0; i-- {
		result = api.Mul(result, base)
		result = api.Add(result, p.Limbs[i])
	}
	return result
}

func publicsToVariables(api frontend.API, ps []emulated.Element[bn254Fr]) []frontend.Variable {
	result := []frontend.Variable{}

	for _, p := range ps {
		result = append(result, publicToVariable(api, p))
	}

	return result
}

func publicsToForm(
	api frontend.API,
	setup *vdf.Setup,
	ps []emulated.Element[bn254Fr],
	offset int,
) (classgroup.Form, int) {
	smallNumLimbs := setup.Classgroup.GetSmallNumLimbs()

	a := bigint.FromHint(publicsToVariables(api, ps[offset:offset+smallNumLimbs+1]))
	offset += smallNumLimbs + 1

	b := bigint.FromHint(publicsToVariables(api, ps[offset:offset+smallNumLimbs+1]))
	offset += smallNumLimbs + 1

	c := bigint.FromHint(publicsToVariables(api, ps[offset:offset+setup.Classgroup.DNumLimbs+1]))
	offset += setup.Classgroup.DNumLimbs + 1

	return classgroup.Form{A: a, B: b, C: c}, offset
}
