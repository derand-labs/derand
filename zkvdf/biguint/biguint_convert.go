package biguint

import (
	"fmt"
	"math/big"
	"zkvdf/utils"

	"github.com/consensys/gnark/frontend"
)

func (setup *Setup) From(in *big.Int, nlimbs int) BigUint {
	tmp := new(big.Int).Set(in)

	out := New(nlimbs)
	for i := range out.Limbs {
		out.Limbs[i] = new(big.Int).And(tmp, utils.MaxValue(setup.LimbBits))
		tmp.Rsh(tmp, uint(setup.LimbBits))
	}

	if tmp.Sign() != 0 {
		panic(fmt.Sprintf("biguint.From: not enough limbs, remaining %s", tmp))
	}

	return out
}

func (setup *Setup) FromUnsafe(in *big.Int, nlimbs int) BigUint {
	tmp := new(big.Int).Set(in)

	out := New(nlimbs)
	for i := range out.Limbs {
		out.Limbs[i] = new(big.Int).And(tmp, utils.MaxValue(setup.LimbBits))
		tmp.Rsh(tmp, uint(setup.LimbBits))
	}

	return out
}

func (api bigUintAPI) Cast(in BigUint, outnlimbs int) BigUint {
	out := New(outnlimbs)
	for i := range outnlimbs {
		out.Limbs[i] = in.GetLimb(i)
	}

	for i := outnlimbs; i < len(in.Limbs); i++ {
		api.core.AssertIsEqual(in.Limbs[i], 0)
	}

	return out
}

func FromHint(hint []frontend.Variable) BigUint {
	out := New(len(hint))
	copy(out.Limbs, hint)
	return out
}

func FromHintInput(limbs []*big.Int, limbbits int) *big.Int {
	out := new(big.Int)

	for i := len(limbs) - 1; i >= 0; i-- {
		out.Lsh(out, uint(limbbits))
		out.Add(out, limbs[i])
	}

	return out
}

func ToHintOutput(hintOutput []*big.Int, in *big.Int, numLimbs, limbbits int) {
	tmp := new(big.Int).Set(in)

	if tmp.Sign() == 0 {
		tmp = big.NewInt(0)
	}

	for i := range numLimbs {
		hintOutput[i].Set(new(big.Int).And(tmp, utils.MaxValue(limbbits)))
		tmp.Rsh(tmp, uint(limbbits))
	}
}
