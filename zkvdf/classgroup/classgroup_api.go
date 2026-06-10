package classgroup

import (
	"zkvdf/bigint"

	"github.com/consensys/gnark/frontend"
)

type API interface {
	BigIntAPI() bigint.API

	GetPrincipalForm() Form
	GetDisciminant() bigint.BigInt

	AssertValid(fs ...Form)
	AssertIsReduced(f Form)

	Select(selector frontend.Variable, f1, f2 Form) Form

	Compose(f1, f2 Form) Form
	Square(f Form) Form
	Reduce(f Form) Form
	Pow(f Form, eBinary []frontend.Variable) Form
	ParitalPow(f, base Form, eBinary []frontend.Variable) (Form, Form)

	IsEqual(f1, f2 Form) frontend.Variable
	AssertIsEqual(f1, f2 Form)
}

type classgroupAPI struct {
	core      frontend.API
	setup     *Setup
	bigintAPI bigint.API

	D             bigint.BigInt
	PrincipalForm Form
}

func NewAPI(api frontend.API, setup *Setup) API {
	D := setup.BigInt.From(setup.D, setup.DNumLimbs)

	pa, pb, pc := setup.GetPrincipalForm()

	p := Form{
		A: setup.BigInt.From(pa, 1),
		B: setup.BigInt.From(pb, 1),
		C: setup.BigInt.From(pc, setup.DNumLimbs),
	}

	return classgroupAPI{
		core:          api,
		setup:         setup,
		bigintAPI:     bigint.NewAPI(api, setup.BigInt),
		D:             D,
		PrincipalForm: p,
	}
}

func (api classgroupAPI) BigIntAPI() bigint.API {
	return api.bigintAPI
}

func (api classgroupAPI) GetDisciminant() bigint.BigInt {
	return api.D
}

func (api classgroupAPI) GetPrincipalForm() Form {
	return api.PrincipalForm
}
