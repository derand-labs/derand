package vdf

import (
	"encoding/json"
	"math/big"
	"os"
	"zkvdf/classgroup"
	"zkvdf/nonzk"
)

type Setup struct {
	LBits    int
	SplitExp int

	HashToFormSteps      int
	HashToFormGenerators []nonzk.Form

	Classgroup *classgroup.Setup
}

func (s *Setup) GetHashToFormGenerators() []classgroup.Form {
	entries := []classgroup.Form{}
	for i := range s.HashToFormGenerators {
		entries = append(entries,
			s.HashToFormGenerators[i].ToZK(s.Classgroup),
		)
	}
	return entries
}

func NewSetup(
	limbbits int,
	D *big.Int,
	dbits, lbits, splitExp int,
	hashToFormSteps int,
	hashToFormGenerators []nonzk.Form,
) *Setup {
	return &Setup{
		LBits:                lbits,
		SplitExp:             splitExp,
		Classgroup:           classgroup.NewSetup(limbbits, D, dbits),
		HashToFormSteps:      hashToFormSteps,
		HashToFormGenerators: hashToFormGenerators,
	}
}

func NewDummySetup(
	limbbits int,
	dbits, lbits int,
	splitExp int,
	hashToFormSteps int,
	hashToFormNbGenerators int,
) *Setup {
	entries := []nonzk.Form{}
	for range hashToFormNbGenerators {
		entries = append(entries, nonzk.Form{A: big.NewInt(0), B: big.NewInt(0), C: big.NewInt(0)})
	}

	return NewSetup(limbbits, big.NewInt(0), dbits, lbits, splitExp, hashToFormSteps, entries)
}

func LoadSetup(systemPath string) (*Setup, error) {
	type system struct {
		D                    string       `json:"D"`
		DBits                int          `json:"d_bits"`
		LBits                int          `json:"l_bits"`
		LimbBits             int          `json:"limb_bits"`
		SplitExp             int          `json:"split_exp"`
		HashToFormSteps      int          `json:"hash_to_form_steps"`
		HashToFormGenerators []nonzk.Form `json:"hash_to_form_generators"`
	}

	data, err := os.ReadFile(systemPath)
	if err != nil {
		return nil, err
	}

	var s system
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}

	hashToFormGenerators := []nonzk.Form{}
	for i := range s.HashToFormGenerators {
		hashToFormGenerators = append(hashToFormGenerators, s.HashToFormGenerators[i])
	}

	return NewSetup(
		s.LimbBits,
		nonzk.MustHexToBigInt(s.D),
		s.DBits,
		s.LBits,
		s.SplitExp,
		s.HashToFormSteps,
		hashToFormGenerators,
	), nil
}
