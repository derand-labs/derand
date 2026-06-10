package vdf

import (
	"zkvdf/classgroup"

	"github.com/consensys/gnark/frontend"
)

type VDFAPI interface {
	ClassgroupAPI() classgroup.API

	HashToForm(seed frontend.Variable) classgroup.Form
	AssertVerify(x, y, pi classgroup.Form, lBinary, rBinary []frontend.Variable)
}

type vdfAPI struct {
	core          frontend.API
	setup         *Setup
	classgroupAPI classgroup.API
}

func NewAPI(api frontend.API, setup *Setup) VDFAPI {
	return vdfAPI{core: api, setup: setup, classgroupAPI: classgroup.NewAPI(api, setup.Classgroup)}
}

func (api vdfAPI) ClassgroupAPI() classgroup.API {
	return api.classgroupAPI
}
