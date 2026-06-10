package bigint

import (
	"math/big"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

func init() {
	solver.RegisterHint(hintInv)
}

func (api bigIntAPI) InvMod(a, m BigInt) BigInt {
	hintInputs := []frontend.Variable{api.setup.BigUint.LimbBits, len(a.Mag.Limbs)}
	hintInputs = append(hintInputs, a.Sign)
	hintInputs = append(hintInputs, a.Mag.Limbs...)
	hintInputs = append(hintInputs, m.Sign)
	hintInputs = append(hintInputs, m.Mag.Limbs...)

	invHint, err := api.core.Compiler().NewHint(hintInv, len(m.Mag.Limbs)+1, hintInputs...)
	if err != nil {
		panic(err)
	}

	inv := FromHint(invHint)
	api.AssertRangeCheck(inv)

	// a*inv - 1 must be divisible by m

	product := api.Mul(a, inv)
	lhs := api.SubOne(product)

	_, r := api.FloorDivQuoRem(lhs, m)

	api.AssertIsZero(r)

	return inv
}

func hintInv(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 2

	limbbits := int(inputs[0].Int64())
	aNumLimbs := int(inputs[1].Int64())
	bNumLimbs := len(inputs) - aNumLimbs - prefixInputs - 2

	a := FromHintInput(inputs[prefixInputs:aNumLimbs+prefixInputs+1], limbbits)
	b := FromHintInput(inputs[aNumLimbs+prefixInputs+1:], limbbits)
	inv := new(big.Int).ModInverse(a, b)
	if inv == nil {
		// If no invert for input, this will give incorrect check in constraint in circuit.
		inv = big.NewInt(1)
	}

	ToHintOutput(outputs, inv, bNumLimbs, limbbits)
	return nil
}
