package biguint

import (
	"github.com/consensys/gnark/frontend"
)

type BigUint struct {
	Limbs []frontend.Variable
}

func New(nlimbs int) BigUint {
	b := BigUint{Limbs: make([]frontend.Variable, nlimbs)}
	for i := range nlimbs {
		b.Limbs[i] = frontend.Variable(0)
	}
	return b
}

func (b BigUint) GetLimb(i int) frontend.Variable {
	if i < len(b.Limbs) {
		return b.Limbs[i]
	}
	return frontend.Variable(0)
}

func (api bigUintAPI) SumLimbs(b BigUint) frontend.Variable {
	sum := frontend.Variable(0)
	if len(b.Limbs) == 1 {
		sum = b.Limbs[0]
	} else if len(b.Limbs) == 2 {
		sum = api.core.Add(b.Limbs[0], b.Limbs[1])
	} else {
		sum = api.core.Add(b.Limbs[0], b.Limbs[1], b.Limbs[2:]...)
	}
	return sum
}

func (api bigUintAPI) ToBinary(b BigUint) []frontend.Variable {
	binary := []frontend.Variable{}

	for _, limb := range b.Limbs {
		binary = append(binary, api.core.ToBinary(limb, api.setup.LimbBits)...)
	}

	return binary
}
