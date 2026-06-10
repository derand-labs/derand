package vdfcircuit_test

import (
	"testing"
	"zkvdf/commontest"
	"zkvdf/vdf"
	"zkvdf/vdfcircuit"
)

func TestHashToFormCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		vdfcircuit.NewVDFHashToFormCircuit(vdf.NewDummySetup(114, 1024, 128, 1, 26, 9097)),
		6413656,
	)
}
