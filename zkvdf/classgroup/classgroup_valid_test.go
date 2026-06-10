package classgroup_test

import (
	"testing"
	"zkvdf/bigint"
	"zkvdf/classgroup"
	"zkvdf/commontest"

	"github.com/consensys/gnark/frontend"
)

func TestClassgroupValidCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		&commontest.MinimalCircuit1[classgroup.Form]{
			A: classgroup.Form{
				A: bigint.New(setup.GetSmallNumLimbs()),
				B: bigint.New(setup.GetSmallNumLimbs()),
				C: bigint.New(setup.DNumLimbs),
			},
			F: func(api frontend.API, a classgroup.Form) {
				classgroup.NewAPI(api, setup).AssertValid(a)
			},
		},
		14123,
	)
}
