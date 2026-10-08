package component

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

func TestNew_OnCreationHookFailureFailsConstruction(t *testing.T) {
	_, err := New("c", WithHooks(func(h *Hooks) {
		h.OnCreation(func(context.Context, *Component) error { return errors.New("not today") })
	}))
	require.ErrorContains(t, err, "onCreation hook failed")
	require.ErrorContains(t, err, "not today")
}

func TestComponent_SetLogger_RejectsNil(t *testing.T) {
	require.ErrorContains(t, mustNew("c").SetLogger(nil), "logger cannot be nil")
}

func TestComponent_LoopbackPipe_NamesTheMissingPort(t *testing.T) {
	c := mustNew("c", WithInputs("in"), WithOutputs("out"))

	require.ErrorContains(t, c.LoopbackPipe("nope", "in"), `output port "nope" not found`)
	require.ErrorContains(t, c.LoopbackPipe("out", "nope"), `input port "nope" not found`)
}

func TestWithIndexedPorts(t *testing.T) {
	c := mustNew("c", WithIndexedInputs("i", 1, 3), WithIndexedOutputs("o", 1, 2))

	// Exact names, so an off-by-one at either end of the range fails.
	portNames := func(ports *port.Collection) []string {
		names := make([]string, 0, ports.Len())
		for p := range ports.All() {
			names = append(names, p.Name())
		}
		return names
	}
	assert.Equal(t, []string{"i1", "i2", "i3"}, portNames(c.Inputs()))
	assert.Equal(t, []string{"o1", "o2"}, portNames(c.Outputs()))

	_, err := New("bad", WithIndexedInputs("i", 3, 1))
	require.ErrorIs(t, err, port.ErrInvalidRangeForIndexedGroup)
}

func TestComponent_Validate_PortParents(t *testing.T) {
	noop := WithActivationFunc(func(context.Context, *Component) error { return nil })
	t.Run("a port attached to two components belongs to the last one", func(t *testing.T) {
		shared, err := port.NewInput("in")
		require.NoError(t, err)
		first, second := mustNew("first", noop), mustNew("second", noop)
		require.NoError(t, first.AttachInputPorts(shared))
		require.NoError(t, second.AttachInputPorts(shared))

		require.ErrorContains(t, first.Validate(), `input port "in" has wrong parent component`)
		require.NoError(t, second.Validate())
	})

	t.Run("the same for an output port", func(t *testing.T) {
		shared, err := port.NewOutput("out")
		require.NoError(t, err)
		first, second := mustNew("first", noop), mustNew("second", noop)
		require.NoError(t, first.AttachOutputPorts(shared))
		require.NoError(t, second.AttachOutputPorts(shared))

		require.ErrorContains(t, first.Validate(), `output port "out" has wrong parent component`)
	})

	t.Run("a port added behind the component's back has no parent", func(t *testing.T) {
		in, err := port.NewInput("in")
		require.NoError(t, err)
		out, err := port.NewOutput("out")
		require.NoError(t, err)

		c := mustNew("c", noop)
		require.NoError(t, c.Inputs().Add(in))
		require.ErrorContains(t, c.Validate(), `input port "in" has no parent component`)

		d := mustNew("d", noop)
		require.NoError(t, d.Outputs().Add(out))
		require.ErrorContains(t, d.Validate(), `output port "out" has no parent component`)
	})
}

func TestCollection_AnyAndMetadata(t *testing.T) {
	col := NewCollection()
	assert.Nil(t, col.First(), "empty collection has nothing to return")

	require.NoError(t, col.Add(mustNew("b"), mustNew("a")))
	assert.Equal(t, "a", col.First().Name(), "name order, so the answer is stable")

	col.Meta().Set("k", "v")
	col.Meta().Set("s", 1.0)
	assert.True(t, col.Meta().ValueIs("k", "v"))
	assert.True(t, col.Meta().ValueIs("s", 1.0))
}

func TestActivationResultCode_String(t *testing.T) {
	tests := map[ActivationResultCode]string{
		ActivationCodeUndefined:             "Undefined",
		ActivationCodeOK:                    "Success",
		ActivationCodeNoInput:               "No input",
		ActivationCodeReturnedError:         "Finished with error",
		ActivationCodePanicked:              "Finished with panic",
		ActivationCodeWaitingForInputsClear: "Waiting for input (clear)",
		ActivationCodeWaitingForInputsKeep:  "Waiting for input (keep)",
		ActivationCodeHookFailed:            "Hook failed",
		ActivationResultCode(99):            "Unknown code",
	}
	for code, want := range tests {
		assert.Equal(t, want, code.String())
	}
}

func TestComponent_FlushOutputs_ReportsARefusedDelivery(t *testing.T) {
	src := mustNew("src", WithOutputs("out"))
	dst := mustNew("dst", WithInputs("in"))
	dst.InputByName("in").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error { return errors.New("refused") })
	})
	require.NoError(t, src.OutputByName("out").PipeTo(dst.InputByName("in")))
	require.NoError(t, src.OutputByName("out").PutSignals(signal.New(1)))

	err := src.FlushOutputs(context.Background())
	require.ErrorContains(t, err, `failed to flush output port "out"`)
	require.ErrorContains(t, err, "refused")
}
