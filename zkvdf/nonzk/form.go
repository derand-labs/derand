package nonzk

import (
	"encoding/json"
	"fmt"
	"math/big"
	"zkvdf/classgroup"
)

type jsonform struct {
	A string `json:"a"`
	B string `json:"b"`
	C string `json:"c"`
}

type Form struct {
	A *big.Int
	B *big.Int
	C *big.Int
}

func (f Form) ToZK(setup *classgroup.Setup) classgroup.Form {
	return classgroup.Form{
		A: setup.BigInt.From(f.A, setup.GetSmallNumLimbs()),
		B: setup.BigInt.From(f.B, setup.GetSmallNumLimbs()),
		C: setup.BigInt.From(f.C, setup.DNumLimbs),
	}
}

func (f Form) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonform{
		A: BigIntToHex(f.A),
		B: BigIntToHex(f.B),
		C: BigIntToHex(f.C),
	})
}

func (f *Form) UnmarshalJSON(b []byte) error {
	jsonf := jsonform{}

	if err := json.Unmarshal(b, &jsonf); err != nil {
		return err
	}

	var err error
	f.A, err = HexToBigInt(jsonf.A)
	if err != nil {
		return fmt.Errorf("f.A: %w", err)
	}

	f.B, err = HexToBigInt(jsonf.B)
	if err != nil {
		return fmt.Errorf("f.B: %w", err)
	}

	f.C, err = HexToBigInt(jsonf.C)
	if err != nil {
		return fmt.Errorf("f.C: %w", err)
	}

	return nil
}
