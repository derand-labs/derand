package bigint

import (
	"math/big"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

func init() {
	solver.RegisterHint(hintGCD)
}

func (api bigIntAPI) GCD(a, b BigInt) (g, ga, gb, x, y BigInt) {
	gnlimbs := min(len(a.Mag.Limbs), len(b.Mag.Limbs))

	hintInputs := []frontend.Variable{api.setup.BigUint.LimbBits, gnlimbs, len(a.Mag.Limbs)}
	hintInputs = append(hintInputs, a.Sign)
	hintInputs = append(hintInputs, a.Mag.Limbs...)
	hintInputs = append(hintInputs, b.Sign)
	hintInputs = append(hintInputs, b.Mag.Limbs...)

	// hintOutputs = g ga gb x y
	gcdHint, err := api.core.Compiler().NewHint(hintGCD, gnlimbs+len(a.Mag.Limbs)*2+len(b.Mag.Limbs)*2+5, hintInputs...)
	if err != nil {
		panic(err)
	}

	offset := 0
	g = FromHint(gcdHint[offset : offset+gnlimbs+1])
	offset += gnlimbs + 1

	ga = FromHint(gcdHint[offset : offset+len(a.Mag.Limbs)+1])
	offset += len(a.Mag.Limbs) + 1

	gb = FromHint(gcdHint[offset : offset+len(b.Mag.Limbs)+1])
	offset += len(b.Mag.Limbs) + 1

	x = FromHint(gcdHint[offset : offset+len(b.Mag.Limbs)+1])
	offset += len(b.Mag.Limbs) + 1

	y = FromHint(gcdHint[offset : offset+len(a.Mag.Limbs)+1])

	api.AssertRangeCheck(g, ga, gb, x, y)

	api.AssertIsEqual(a, api.Mul(ga, g))
	api.AssertIsEqual(b, api.Mul(gb, g))

	ax := api.Mul(a, x)
	by := api.Mul(b, y)
	axby := api.Add(ax, by)
	api.AssertIsEqual(axby, g)

	aIsZero := api.IsZero(a)
	bIsZero := api.IsZero(b)

	g = api.Select(aIsZero, b, g)
	g = api.Select(bIsZero, a, g)

	return g, ga, gb, x, y
}

func hintGCD(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	const prefixInputs = 3

	limbbits := int(inputs[0].Int64())
	gNumLimbs := int(inputs[1].Int64())
	aNumLimbs := int(inputs[2].Int64())
	bNumLimbs := len(inputs) - aNumLimbs - prefixInputs - 2

	a := FromHintInput(inputs[prefixInputs:aNumLimbs+prefixInputs+1], limbbits)
	b := FromHintInput(inputs[aNumLimbs+prefixInputs+1:], limbbits)

	g := new(big.Int)
	x := new(big.Int)
	y := new(big.Int)

	g.GCD(x, y, a, b)
	ga := big.NewInt(0)
	gb := big.NewInt(0)
	if g.Sign() != 0 {
		ga = new(big.Int).Div(a, g)
		gb = new(big.Int).Div(b, g)
	}

	offset := 0

	ToHintOutput(outputs[offset:], g, gNumLimbs, limbbits)
	offset += gNumLimbs + 1

	ToHintOutput(outputs[offset:], ga, aNumLimbs, limbbits)
	offset += aNumLimbs + 1

	ToHintOutput(outputs[offset:], gb, bNumLimbs, limbbits)
	offset += bNumLimbs + 1

	ToHintOutput(outputs[offset:], x, bNumLimbs, limbbits)
	offset += bNumLimbs + 1

	ToHintOutput(outputs[offset:], y, aNumLimbs, limbbits)

	return nil
}
