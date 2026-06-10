package vdfcircuit_test

import (
	"testing"
	"zkvdf/commontest"
	"zkvdf/vdf"
	"zkvdf/vdfcircuit"
)

func TestIntermediatePowCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		vdfcircuit.NewVDFIntermediatePow(vdf.NewDummySetup(114, 1024, 128, 1, 26, 9097)),
		6066935,
	)
}
