package vdfcircuit

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"zkvdf/nonzk"
	"zkvdf/utils"
	"zkvdf/vdf"
)

type IntermediatePowTranscript struct {
	L       *big.Int
	PiL     nonzk.Form
	PiLBase nonzk.Form

	R      *big.Int
	XR     nonzk.Form
	XRBase nonzk.Form
}

type Transcript struct {
	XSeed *big.Int

	X  nonzk.Form
	Y  nonzk.Form
	Pi nonzk.Form

	L, R             *big.Int
	LPhase1, RPhase1 *big.Int

	IntermediatePows []IntermediatePowTranscript

	ChallengeLTranscript []*big.Int
}

func LoadTranscript(setup *vdf.Setup, path string) (*Transcript, error) {
	type dummytranscript struct {
		XSeed string `json:"x_seed"`

		X  nonzk.Form `json:"x"`
		Y  nonzk.Form `json:"y"`
		Pi nonzk.Form `json:"pi"`

		IntermediatePows []struct {
			L       string     `json:"l"`
			PiL     nonzk.Form `json:"pil"`
			PiLBase nonzk.Form `json:"pil_base"`

			R      string     `json:"r"`
			XR     nonzk.Form `json:"xr"`
			XRBase nonzk.Form `json:"xr_base"`
		} `json:"intermediate_pows"`

		ChallengeLTranscript []string `json:"challenge_l_transcript"`
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript file: %w", err)
	}

	var dummyTranscript dummytranscript

	d := json.NewDecoder(f)
	if err := d.Decode(&dummyTranscript); err != nil {
		return nil, fmt.Errorf("failed to decode transcript file: %w", err)
	}

	splitExp := len(dummyTranscript.IntermediatePows)
	if setup.LBits%splitExp != 0 {
		return nil, fmt.Errorf("vdfcircuit.LoadTranscript: invalid intermediate pow length")
	}

	bitSizePerIntermediate := setup.LBits / splitExp
	finalL, finalR := big.NewInt(0), big.NewInt(0)
	lPhase1, rPhase1 := big.NewInt(0), big.NewInt(0)
	var intermediatePows []IntermediatePowTranscript
	for i, item := range dummyTranscript.IntermediatePows {
		L := nonzk.MustHexToBigInt(item.L)
		R := nonzk.MustHexToBigInt(item.R)
		intermediatePows = append(intermediatePows, IntermediatePowTranscript{
			L:       L,
			R:       R,
			PiL:     item.PiL,
			XR:      item.XR,
			PiLBase: item.PiLBase,
			XRBase:  item.XRBase,
		})

		finalL.Add(finalL, new(big.Int).Mul(L, utils.Modulus(bitSizePerIntermediate*i)))
		finalR.Add(finalR, new(big.Int).Mul(R, utils.Modulus(bitSizePerIntermediate*i)))

		if i < splitExp/2 {
			lPhase1.Add(lPhase1, new(big.Int).Mul(L, utils.Modulus(bitSizePerIntermediate*i)))
			rPhase1.Add(rPhase1, new(big.Int).Mul(R, utils.Modulus(bitSizePerIntermediate*i)))
		}
	}

	return &Transcript{
		XSeed:                nonzk.MustHexToBigInt(dummyTranscript.XSeed),
		X:                    dummyTranscript.X,
		Y:                    dummyTranscript.Y,
		Pi:                   dummyTranscript.Pi,
		L:                    finalL,
		R:                    finalR,
		LPhase1:              lPhase1,
		RPhase1:              rPhase1,
		IntermediatePows:     intermediatePows,
		ChallengeLTranscript: nonzk.MustArrayHexToBigInt(dummyTranscript.ChallengeLTranscript),
	}, nil
}
