package errorhandling

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

// What each ErrorHandlingStrategy does when a component errors or panics.

// countingMesh builds a mesh with one misbehaving component and one that keeps
// counting activations, so a strategy that continues is visibly different from
// one that stops.
func countingMesh(t *testing.T, strategy fmesh.ErrorHandlingStrategy, misbehave component.ActivationFunc) (fm *fmesh.FMesh, activationCount *int) {
	t.Helper()

	activations := 0
	counter := testutil.MustComponent("counter",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			activations++
			return this.OutputByName("out").PutSignals(signal.New(activations))
		}))
	require.NoError(t, counter.LoopbackPipe("out", "in"))

	bad := testutil.MustComponent("bad",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(misbehave))
	require.NoError(t, bad.LoopbackPipe("out", "in"))

	fm = testutil.MustFMesh("strategies",
		fmesh.WithErrorHandlingStrategy(strategy),
		fmesh.WithCyclesLimit(5))
	require.NoError(t, fm.AddComponents(counter, bad))
	require.NoError(t, counter.InputByName("in").PutSignals(signal.New(0)))
	require.NoError(t, bad.InputByName("in").PutSignals(signal.New(0)))

	return fm, &activations
}

func returnsError(_ context.Context, this *component.Component) error {
	// Keep the loop alive so the mesh would continue if the strategy allows it.
	if err := this.OutputByName("out").PutSignals(signal.New(0)); err != nil {
		return err
	}
	return errors.New("deliberate failure")
}

func panics(_ context.Context, this *component.Component) error {
	if err := this.OutputByName("out").PutSignals(signal.New(0)); err != nil {
		return err
	}
	panic("deliberate panic")
}

func TestStrategies_StopOnFirstErrorOrPanic(t *testing.T) {
	t.Run("stops on an error", func(t *testing.T) {
		fm, activations := countingMesh(t, fmesh.StopOnFirstErrorOrPanic, returnsError)

		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrHitAnErrorOrPanic)
		assert.Equal(t, 1, *activations, "the run stops after the failing cycle")
	})

	t.Run("stops on a panic", func(t *testing.T) {
		fm, activations := countingMesh(t, fmesh.StopOnFirstErrorOrPanic, panics)

		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrHitAnErrorOrPanic)
		assert.Equal(t, 1, *activations)
	})
}

func TestStrategies_StopOnFirstPanic(t *testing.T) {
	t.Run("ignores errors and keeps running", func(t *testing.T) {
		fm, activations := countingMesh(t, fmesh.StopOnFirstPanic, returnsError)

		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrReachedMaxAllowedCycles,
			"errors do not stop this strategy, so the cycle limit does")
		assert.Greater(t, *activations, 1, "the healthy component keeps working")
	})

	t.Run("stops on a panic", func(t *testing.T) {
		fm, activations := countingMesh(t, fmesh.StopOnFirstPanic, panics)

		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrHitAPanic)
		assert.Equal(t, 1, *activations)
	})
}

func TestStrategies_IgnoreAll(t *testing.T) {
	for _, tt := range []struct {
		name      string
		misbehave component.ActivationFunc
	}{
		{"error", returnsError},
		{"panic", panics},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fm, activations := countingMesh(t, fmesh.IgnoreAll, tt.misbehave)

			_, err := fm.Run(context.Background())

			require.ErrorIs(t, err, fmesh.ErrReachedMaxAllowedCycles,
				"nothing stops this strategy but a limit")
			assert.Greater(t, *activations, 1)
		})
	}
}

func TestStrategies_FailuresAreAlwaysRecordedInRuntimeInfo(t *testing.T) {
	// Whatever the strategy does about stopping, the results are reported.
	fm, _ := countingMesh(t, fmesh.IgnoreAll, returnsError)

	ri, err := fm.Run(context.Background())
	require.Error(t, err)

	firstCycle := ri.Cycles.First()
	require.NotNil(t, firstCycle)
	assert.True(t, firstCycle.HasActivationErrors())
	require.Error(t, firstCycle.AllErrorsCombined())
	assert.Contains(t, firstCycle.AllErrorsCombined().Error(), "deliberate failure")
}
