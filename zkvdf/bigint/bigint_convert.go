package bigint

import (
	"math/big"
	"zkvdf/biguint"

	"github.com/consensys/gnark/frontend"
)

func New(numlimbs int) BigInt {
	return BigInt{Sign: frontend.Variable(0), Mag: biguint.New(numlimbs)}
}

func (setup *Setup) From(v *big.Int, numlimbs int) BigInt {
	var out BigInt

	tmp := new(big.Int).Set(v)
	out.Sign = tmp.Sign()

	if tmp.Sign() < 0 {
		tmp.Neg(tmp)
	}

	out.Mag = setup.BigUint.From(tmp, numlimbs)
	return out
}

func (setup *Setup) FromUnsafe(v *big.Int, numlimbs int) BigInt {
	var out BigInt

	tmp := new(big.Int).Set(v)
	out.Sign = tmp.Sign()

	if tmp.Sign() < 0 {
		tmp.Neg(tmp)
	}

	out.Mag = setup.BigUint.FromUnsafe(tmp, numlimbs)
	return out
}

func (api bigIntAPI) Cast(in BigInt, outnlimbs int) BigInt {
	return BigInt{
		Sign: in.Sign,
		Mag:  api.BigUintAPI().Cast(in.Mag, outnlimbs),
	}
}

func FromHint(hint []frontend.Variable) BigInt {
	var out BigInt

	out.Sign = hint[0]
	out.Mag = biguint.FromHint(hint[1:])

	return out
}

func FromHintInput(hintInput []*big.Int, limbbits int) *big.Int {
	v := biguint.FromHintInput(hintInput[1:], limbbits)
	sign := hintInput[0].Int64()
	if sign == 0 && v.Int64() != 0 {
		panic("bigint.FromHintInput: invalid zero sign")
	}

	if sign != 0 && sign != 1 {
		v.Neg(v)
	}

	return v
}

func ToHintOutput(hintOutput []*big.Int, v *big.Int, numLimbs, limbbits int) {
	tmp := new(big.Int).Set(v)
	hintOutput[0].Set(big.NewInt(int64(tmp.Sign())))

	if tmp.Sign() < 0 {
		tmp.Neg(tmp)
	}

	if tmp.Sign() == 0 {
		tmp = big.NewInt(0)
	}

	biguint.ToHintOutput(hintOutput[1:], tmp, numLimbs, limbbits)
}
