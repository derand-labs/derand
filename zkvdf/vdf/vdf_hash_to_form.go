package vdf

import (
	"math/big"
	"zkvdf/classgroup"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/cmp"
	"github.com/consensys/gnark/std/permutation/poseidon2"
)

func init() {
	solver.RegisterHint(hintDivQuoRem)
}

func (api vdfAPI) HashToForm(seed frontend.Variable) classgroup.Form {
	perm, err := poseidon2.NewPoseidon2FromParameters(api.core, 2, 6, 50)
	if err != nil {
		panic(err)
	}

	lookuptable := classgroup.NewLookupTable(api.core)
	lookuptable.Insert(api.setup.GetHashToFormGenerators()...)

	acc := api.ClassgroupAPI().GetPrincipalForm()

	for range api.setup.HashToFormSteps {
		seed = api.core.Add(seed, 1)

		idxHash := perm.Compress(0, seed)
		_, idx := quorem(api.core, idxHash, uint64(len(api.setup.HashToFormGenerators)))

		selected := lookuptable.Lookup(idx)

		acc = api.ClassgroupAPI().Compose(acc, selected)
		acc = api.ClassgroupAPI().Reduce(acc)
	}

	return acc
}

func quorem(api frontend.API, x frontend.Variable, m uint64) (frontend.Variable, frontend.Variable) {
	quoremHint, err := api.NewHint(hintDivQuoRem, 2, x, m)
	if err != nil {
		panic(err)
	}

	q, r := quoremHint[0], quoremHint[1]

	comparator := cmp.NewBoundedComparator(api, big.NewInt(int64(m)), false)

	// r = x - q*k
	api.AssertIsEqual(x, api.Add(api.Mul(q, m), r))

	//  < k
	comparator.AssertIsLess(r, m)

	return q, r
}

func hintDivQuoRem(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	q, r := new(big.Int).QuoRem(inputs[0], inputs[1], new(big.Int))
	outputs[0] = q
	outputs[1] = r
	return nil
}
