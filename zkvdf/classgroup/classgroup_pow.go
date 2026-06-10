package classgroup

import (
	"github.com/consensys/gnark/frontend"
)

func (api classgroupAPI) Pow(x Form, eBinary []frontend.Variable) Form {
	acc := api.PrincipalForm
	cur := x
	for _, bit := range eBinary {
		accIfBit := api.Compose(acc, cur)
		accIfBit = api.Reduce(accIfBit)

		acc = api.Select(bit, accIfBit, acc)

		cur = api.Square(cur)
		cur = api.Reduce(cur)
	}

	return acc
}

func (api classgroupAPI) ParitalPow(x, base Form, eBinary []frontend.Variable) (Form, Form) {
	acc := x
	cur := base
	for _, bit := range eBinary {
		accIfBit := api.Compose(acc, cur)
		accIfBit = api.Reduce(accIfBit)

		acc = api.Select(bit, accIfBit, acc)

		cur = api.Square(cur)
		cur = api.Reduce(cur)
	}

	return acc, cur
}
