package classgroup

import (
	"github.com/consensys/gnark/frontend"
)

func (api classgroupAPI) Select(selector frontend.Variable, f1, f2 Form) Form {
	return Form{
		A: api.BigIntAPI().Select(selector, f1.A, f2.A),
		B: api.BigIntAPI().Select(selector, f1.B, f2.B),
		C: api.BigIntAPI().Select(selector, f1.C, f2.C),
	}
}
