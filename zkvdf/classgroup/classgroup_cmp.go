package classgroup

import (
	"github.com/consensys/gnark/frontend"
)

func (api classgroupAPI) IsEqual(f1, f2 Form) frontend.Variable {
	return api.core.Mul(
		api.BigIntAPI().IsEqual(f1.A, f2.A),
		api.core.Mul(
			api.BigIntAPI().IsEqual(f1.B, f2.B),
			api.BigIntAPI().IsEqual(f1.C, f2.C),
		),
	)
}

func (api classgroupAPI) AssertIsEqual(f1, f2 Form) {
	api.BigIntAPI().AssertIsEqual(f1.A, f2.A)
	api.BigIntAPI().AssertIsEqual(f1.B, f2.B)
	api.BigIntAPI().AssertIsEqual(f1.C, f2.C)
}
