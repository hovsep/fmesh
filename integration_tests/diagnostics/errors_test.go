// Package diagnostics covers what a failing mesh actually reports, as opposed to
// what it does. An error that is technically correct but unreadable, or that
// prints a formatting placeholder, costs debugging time in exactly the moment a
// user can least afford it.
package diagnostics

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
)

func meshWith(t *testing.T, name string, af component.ActivationFunc, opts ...fmesh.Option) *fmesh.FMesh {
	t.Helper()
	c := testutil.MustComponent(name,
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(af))

	fm, err := fmesh.New("diagnostics", opts...)
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	return fm
}

func TestRunError_NoFormattingPlaceholders(t *testing.T) {
	// A cycle with errors but no panics used to render
	// "activation panics: %!w(<nil>)" — a nil handed to %w — and that string
	// reached user logs.
	t.Run("errors without panics", func(t *testing.T) {
		fm := meshWith(t, "failing", func(context.Context, *component.Component) error {
			return errors.New("transient network blip")
		})

		_, err := fm.Run(context.Background())

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "%!", "no formatting placeholder may reach the message")
		assert.Contains(t, err.Error(), "activation errors:")
		assert.NotContains(t, err.Error(), "activation panics:",
			"a cycle without panics must not mention them at all")
		assert.Contains(t, err.Error(), "transient network blip")
		t.Logf("reported as: %v", err)
	})

	t.Run("panics without errors", func(t *testing.T) {
		fm := meshWith(t, "panicking", func(context.Context, *component.Component) error {
			panic("kaboom")
		})

		_, err := fm.Run(context.Background())

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "%!")
		assert.Contains(t, err.Error(), "activation panics:")
		assert.NotContains(t, err.Error(), "activation errors:")
		t.Logf("reported as: %v", err)
	})
}

func TestPanicError_IsOneReadableLine(t *testing.T) {
	fm := meshWith(t, "panicking", func(context.Context, *component.Component) error {
		panic("kaboom")
	})

	_, err := fm.Run(context.Background())
	require.Error(t, err)

	// The whole point: the stack no longer lands in the message.
	assert.NotContains(t, err.Error(), "goroutine ")
	assert.NotContains(t, err.Error(), "runtime/debug.Stack")
	assert.Less(t, len(err.Error()), 500,
		"a panic message must stay loggable, not carry a multi-kilobyte stack")
	assert.Contains(t, err.Error(), "panicked: kaboom")
}

func TestPanicError_CarriesStackAndValue(t *testing.T) {
	fm := meshWith(t, "panicking", func(context.Context, *component.Component) error {
		panic("kaboom")
	})

	_, err := fm.Run(context.Background())
	require.Error(t, err)

	var panicErr *component.PanicError
	require.ErrorAs(t, err, &panicErr, "the panic must be reachable with errors.As")

	assert.Equal(t, "panicking", panicErr.ComponentName)
	assert.Equal(t, "kaboom", panicErr.Value)
	assert.NotEmpty(t, panicErr.StackTrace(), "the stack is kept, just not printed")
	assert.Contains(t, string(panicErr.StackTrace()), "goroutine",
		"StackTrace must be a real stack")
}

func TestPanicError_UnwrapsAThrownError(t *testing.T) {
	// Panicking with an error is common (a library that panics on misuse). Being
	// able to errors.Is through the panic to that error is the difference between
	// handling it and string-matching a message.
	sentinel := errors.New("underlying failure")

	fm := meshWith(t, "panicking", func(context.Context, *component.Component) error {
		panic(sentinel)
	})

	_, err := fm.Run(context.Background())
	require.Error(t, err)

	require.ErrorIs(t, err, sentinel, "errors.Is must reach through the panic to what was thrown")

	var panicErr *component.PanicError
	require.ErrorAs(t, err, &panicErr)
	assert.Equal(t, sentinel, panicErr.Value)
}

func TestPanicError_NonErrorValueDoesNotUnwrap(t *testing.T) {
	fm := meshWith(t, "panicking", func(context.Context, *component.Component) error {
		panic(42)
	})

	_, err := fm.Run(context.Background())
	require.Error(t, err)

	var panicErr *component.PanicError
	require.ErrorAs(t, err, &panicErr)
	assert.Equal(t, 42, panicErr.Value)
	require.NoError(t, panicErr.Unwrap(), "an int panic unwraps to nothing")
	assert.Contains(t, err.Error(), "panicked: 42")
}

func TestRunError_ReportsBothErrorsAndPanics(t *testing.T) {
	failing := testutil.MustComponent("failing",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(context.Context, *component.Component) error {
			return errors.New("ordinary failure")
		}))
	panicking := testutil.MustComponent("panicking",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(context.Context, *component.Component) error {
			panic("kaboom")
		}))

	fm, err := fmesh.New("both")
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(failing, panicking))
	require.NoError(t, failing.InputByName("in").PutSignals(signal.New(1)))
	require.NoError(t, panicking.InputByName("in").PutSignals(signal.New(1)))

	_, err = fm.Run(context.Background())

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "%!")
	assert.Contains(t, err.Error(), "activation errors:")
	assert.Contains(t, err.Error(), "activation panics:")
	assert.Contains(t, err.Error(), "ordinary failure")
	assert.Contains(t, err.Error(), "kaboom")
	t.Logf("reported as: %v", err)
}
