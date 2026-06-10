package vdf

import (
	"zkvdf/classgroup"

	"github.com/consensys/gnark/frontend"
)

func (api vdfAPI) AssertVerify(x, y, pi classgroup.Form, lBinary, rBinary []frontend.Variable) {
	pil := api.ClassgroupAPI().Pow(pi, lBinary)
	xr := api.ClassgroupAPI().Pow(x, rBinary)

	rhs := api.ClassgroupAPI().Compose(pil, xr)
	rhs = api.ClassgroupAPI().Reduce(rhs)

	api.ClassgroupAPI().AssertIsEqual(y, rhs)
}
