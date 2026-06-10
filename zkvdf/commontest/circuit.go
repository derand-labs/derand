package commontest

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MinimalCircuitDefineFunc1[T any] func(api frontend.API, a T)

type MinimalCircuit1[T any] struct {
	A T `gnark:",public"`
	F MinimalCircuitDefineFunc1[T]
}

func (c MinimalCircuit1[T]) Define(api frontend.API) error {
	c.F(api, c.A)
	return nil
}

type MinimalCircuitDefineFunc2[T any] func(api frontend.API, a, b T)

type MinimalCircuit2[T any] struct {
	A T `gnark:",public"`
	B T `gnark:",public"`
	F MinimalCircuitDefineFunc2[T]
}

func (c MinimalCircuit2[T]) Define(api frontend.API) error {
	c.F(api, c.A, c.B)
	return nil
}

type MinimalCircuitDefineFunc3[T any] func(api frontend.API, a, b, c T)

type MinimalCircuit3[T any] struct {
	A T `gnark:",public"`
	B T `gnark:",public"`
	C T `gnark:",public"`
	F MinimalCircuitDefineFunc3[T]
}

func (c MinimalCircuit3[T]) Define(api frontend.API) error {
	c.F(api, c.A, c.B, c.C)
	return nil
}

func AssertCircuitConstraints[C frontend.Circuit](t *testing.T, assignment C, num int) {
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, assignment)
	require.NoError(t, err)

	assert.Equal(t, num, ccs.GetNbConstraints())
}

func TestCircuitValid[C frontend.Circuit](t *testing.T, assignment C) {
	assert := test.NewAssert(t)

	assert.CheckCircuit(
		assignment,
		test.WithValidAssignment(assignment),
		test.WithCurves(ecc.BN254),
		test.WithBackends(backend.PLONK),
	)
}

func TestCircuitInvalid[C frontend.Circuit](t *testing.T, assignment C) {
	assert := test.NewAssert(t)

	assert.CheckCircuit(
		assignment,
		test.WithInvalidAssignment(assignment),
		test.WithCurves(ecc.BN254),
		test.WithBackends(backend.PLONK),
	)
}
