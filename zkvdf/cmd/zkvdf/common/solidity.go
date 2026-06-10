package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"zkvdf/nonzk"
	"zkvdf/utils"
	"zkvdf/vdf"
	"zkvdf/vdfcircuit"

	"github.com/consensys/gnark/backend/plonk"
	be_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
)

type SolidityForm struct {
	Asign, Bsign, Csign int8

	// Use uint128 limbs for bigint: each limb is chosen to stay close to 128 bits, but never exceed
	// it, to avoid overflowing intermediate multiplications.
	A, B, C []*big.Int
}

func NewSolidityForm(setup *vdf.Setup, f nonzk.Form) SolidityForm {
	asign, a := extractBigIntToLimbs(f.A, setup.Classgroup.BigInt.BigUint.LimbBits, setup.Classgroup.GetSmallNumLimbs())
	bsign, b := extractBigIntToLimbs(f.B, setup.Classgroup.BigInt.BigUint.LimbBits, setup.Classgroup.GetSmallNumLimbs())
	csign, c := extractBigIntToLimbs(f.C, setup.Classgroup.BigInt.BigUint.LimbBits, setup.Classgroup.DNumLimbs)

	return SolidityForm{
		Asign: asign,
		Bsign: bsign,
		Csign: csign,
		A:     a,
		B:     b,
		C:     c,
	}
}

func (form SolidityForm) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonSolidityFormFromSolidity(form))
}

func (form *SolidityForm) UnmarshalJSON(b []byte) error {
	jsonForm := jsonSolidityForm{}
	if err := json.Unmarshal(b, &jsonForm); err != nil {
		return err
	}

	*form = jsonForm.ToSolidityForm()
	return nil
}

type SolidityProof struct {
	Y, Pi                     SolidityForm
	DeriveChallengeTranscript []*big.Int
	ZkProof                   []byte
}

func (proof SolidityProof) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonSolidityProof{
		Proof:                "0x" + hex.EncodeToString(proof.ZkProof),
		Y:                    proof.Y,
		Pi:                   proof.Pi,
		ChallengeLTranscript: nonzk.ArrayBigIntToHex(proof.DeriveChallengeTranscript),
	})
}

func (proof *SolidityProof) UnmarshalJSON(b []byte) error {
	jsonProof := jsonSolidityProof{}
	if err := json.Unmarshal(b, &jsonProof); err != nil {
		return err
	}

	if !strings.HasPrefix(jsonProof.Proof, "0x") {
		return fmt.Errorf("invalid proof: must has 0x prefix")
	}

	var err error
	proof.ZkProof, err = hex.DecodeString(jsonProof.Proof[2:])
	if err != nil {
		return err
	}

	proof.Y = jsonProof.Y
	proof.Pi = jsonProof.Pi
	proof.DeriveChallengeTranscript = nonzk.MustArrayHexToBigInt(jsonProof.ChallengeLTranscript)

	return nil
}

func NewSolidityProof(setup *vdf.Setup, proof plonk.Proof, transcript *vdfcircuit.Transcript) (*SolidityProof, error) {
	bn254Proof, ok := proof.(*be_bn254.Proof)
	if !ok {
		return nil, fmt.Errorf("invalid proof: not a bn254")
	}

	result := SolidityProof{
		ZkProof:                   bn254Proof.MarshalSolidity(),
		Y:                         NewSolidityForm(setup, transcript.Y),
		Pi:                        NewSolidityForm(setup, transcript.Pi),
		DeriveChallengeTranscript: transcript.ChallengeLTranscript,
	}

	return &result, nil
}

type jsonSolidityForm struct {
	ASign int8     `json:"asign"`
	BSign int8     `json:"bsign"`
	CSign int8     `json:"csign"`
	A     []string `json:"a"`
	B     []string `json:"b"`
	C     []string `json:"c"`
}

func jsonSolidityFormFromSolidity(f SolidityForm) jsonSolidityForm {
	return jsonSolidityForm{
		ASign: f.Asign,
		BSign: f.Bsign,
		CSign: f.Csign,
		A:     nonzk.ArrayBigIntToHex(f.A),
		B:     nonzk.ArrayBigIntToHex(f.B),
		C:     nonzk.ArrayBigIntToHex(f.C),
	}
}

func (f jsonSolidityForm) ToSolidityForm() SolidityForm {
	return SolidityForm{
		Asign: f.ASign,
		Bsign: f.BSign,
		Csign: f.CSign,
		A:     nonzk.MustArrayHexToBigInt(f.A),
		B:     nonzk.MustArrayHexToBigInt(f.B),
		C:     nonzk.MustArrayHexToBigInt(f.C),
	}
}

type jsonSolidityProof struct {
	Y                    SolidityForm `json:"y"`
	Pi                   SolidityForm `json:"pi"`
	Proof                string       `json:"zk_proof"`
	ChallengeLTranscript []string     `json:"challenge_l_transcript"`
}

func extractBigIntToLimbs(f *big.Int, limbBits, size int) (int8, []*big.Int) {
	limbs := make([]*big.Int, size)
	for i := range size {
		limbs[i] = new(big.Int)
	}

	if f.Sign() == 0 {
		return 0, limbs
	}

	x := new(big.Int).Set(f)
	sign := int8(1)
	if x.Sign() < 0 {
		sign = -1
		x.Neg(x)
	}

	mask := utils.MaxValue(limbBits)

	i := 0
	for x.Sign() > 0 {
		limbs[i].And(x, mask)
		x.Rsh(x, uint(limbBits))
		i++
	}

	return sign, limbs
}
