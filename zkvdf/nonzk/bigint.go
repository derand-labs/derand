package nonzk

import (
	"fmt"
	"math/big"
	"strings"
)

func BigIntToHex(a *big.Int) string {
	if a.Sign() < 0 {
		return "-0x" + new(big.Int).Neg(a).Text(16)
	}
	return "0x" + a.Text(16)
}

func HexToBigInt(a string) (*big.Int, error) {
	start := 0
	if strings.HasPrefix(a, "0x") {
		start = 2
	}
	if strings.HasPrefix(a, "-0x") {
		start = 3
	}

	if start == 0 {
		return nil, fmt.Errorf("require 0x prefix")
	}

	b, ok := new(big.Int).SetString(a[start:], 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex")
	}

	if start == 3 {
		b.Neg(b)
	}

	return b, nil
}

func MustHexToBigInt(a string) *big.Int {
	b, err := HexToBigInt(a)
	if err != nil {
		panic(err)
	}
	return b
}

func ArrayBigIntToHex(a []*big.Int) []string {
	s := make([]string, len(a))
	for i := range a {
		s[i] = BigIntToHex(a[i])
	}
	return s
}

func ArrayHexToBigInt(s []string) ([]*big.Int, error) {
	a := make([]*big.Int, len(s))
	var err error
	for i := range a {
		a[i], err = HexToBigInt(s[i])
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
	}
	return a, nil
}

func MustArrayHexToBigInt(s []string) []*big.Int {
	a, err := ArrayHexToBigInt(s)
	if err != nil {
		panic(err)
	}

	return a
}
