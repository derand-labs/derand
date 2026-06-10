package vdfcircuit_test

import (
	"testing"
	"zkvdf/commontest"
	"zkvdf/vdf"
	"zkvdf/vdfcircuit"
)

func TestVerifierCircuitNbConstraints(t *testing.T) {
	commontest.AssertCircuitConstraints(
		t,
		vdfcircuit.NewVerifier(vdf.NewDummySetup(64, 512, 16, 1, 8, 1024)),
		7257250,
	)
}
