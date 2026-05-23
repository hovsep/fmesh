package errorhandling_test

import (
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorReturns_DuplicateComponent(t *testing.T) {
	fm, err := fmesh.New("fm")
	require.NoError(t, err)
	noop := component.WithActivationFunc(func(c *component.Component) error { return nil })
	c1, err := component.New("c", noop)
	require.NoError(t, err)
	c2, err := component.New("c", noop)
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(c1))
	assert.Error(t, fm.AddComponents(c2))
}

func TestErrorReturns_WrongPortDirection(t *testing.T) {
	c, err := component.New("c")
	require.NoError(t, err)
	outPort, err := port.NewOutput("out")
	require.NoError(t, err)
	// Attaching an output port as input should return an error
	assert.Error(t, c.AttachInputPorts(outPort))
}

func TestErrorReturns_FMeshRun(t *testing.T) {
	fm, err := fmesh.New("fm")
	require.NoError(t, err)
	c, err := component.New("c",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(this *component.Component) error {
			payload := this.InputByName("in").Signals().FirstPayloadOrNil()
			return this.OutputByName("out").PutSignals(signal.New(payload))
		}),
	)
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(42)))
	_, err = fm.Run()
	require.NoError(t, err)
	assert.Equal(t, 42, c.OutputByName("out").Signals().FirstPayloadOrNil())
}
