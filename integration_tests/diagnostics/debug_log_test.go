package diagnostics

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
)

// Debug mode is the only place the panic stack trace is ever printed — it is
// deliberately kept out of the error message so logs stay readable — so if this
// path rots, the stack is lost with no other way to reach it.
func TestDebugMode_LogsActivationResultsAndPanicStack(t *testing.T) {
	var out bytes.Buffer
	logger := log.New(&out, "", 0)

	panicking := testutil.MustComponent("panicking",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, _ *component.Component) error {
			panic(errors.New("boom"))
		}),
	)

	fm := testutil.MustFMesh("debug fm",
		fmesh.WithDebug(true),
		fmesh.WithLogger(logger),
		fmesh.WithErrorHandlingStrategy(fmesh.IgnoreAll),
		fmesh.WithCyclesLimit(1),
	)
	require.NoError(t, fm.AddComponents(panicking))
	require.NoError(t, panicking.InputByName("in").PutSignals(signal.New(1)))

	_, err := fm.Run(context.Background())
	require.Error(t, err)

	logged := out.String()
	assert.Contains(t, logged, "activation result for component panicking",
		"debug mode must report each activation result")
	assert.Contains(t, logged, "is panic: true")
	assert.Contains(t, logged, "code: Finished with panic",
		"the activation code must render as a name, not an integer")
	assert.Contains(t, logged, "stack trace for component panicking",
		"the stack is only ever printed here")
}

// The counterpart: with debug off nothing is written, which is what makes the
// IsDebug guard worth having rather than always formatting and discarding.
func TestDebugMode_SilentWhenDisabled(t *testing.T) {
	var out bytes.Buffer

	c := testutil.MustComponent("noop",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, _ *component.Component) error { return nil }),
	)

	fm := testutil.MustFMesh("quiet fm", fmesh.WithLogger(log.New(&out, "", 0)))
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))

	_, err := fm.Run(context.Background())
	require.NoError(t, err)
	assert.Empty(t, out.String())
}
