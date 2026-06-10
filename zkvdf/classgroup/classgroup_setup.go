package classgroup

import (
	"math/big"
	"zkvdf/bigint"
)

type Setup struct {
	D         *big.Int
	DNumLimbs int
	BigInt    *bigint.Setup
}

func NewSetup(limbbits int, D *big.Int, dbits int) *Setup {
	return &Setup{
		D:         D,
		DNumLimbs: (dbits + limbbits - 1) / limbbits,
		BigInt:    bigint.NewSetup(limbbits),
	}
}

func (s *Setup) GetSmallNumLimbs() int {
	return max(1, (s.DNumLimbs+1)/2)
}

func (s *Setup) GetPrincipalForm() (*big.Int, *big.Int, *big.Int) {
	tmp := new(big.Int).Sub(big.NewInt(1), s.D)
	c := new(big.Int).Div(tmp, big.NewInt(4))

	return big.NewInt(1), big.NewInt(1), c
}
