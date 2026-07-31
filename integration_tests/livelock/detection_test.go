// Package livelock covers the detector that ends a run which has stopped making
// progress.
//
// The interesting tests here are the negative ones. A detector that reports a
// livelock in a mesh that was merely slow to converge is worse than no detector
// at all: it kills correct runs, and it does so intermittently.
package livelock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
)

// waiter is a component that suspends itself forever in the given mode.
func waiter(name string, waitErr error) *component.Component {
	return testutil.MustComponent(name,
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, _ *component.Component) error {
			return waitErr
		}))
}

func TestLivelock_MutualWait(t *testing.T) {
	// The classic: a and b each wait for the other, keeping their inputs, so the
	// mesh reproduces the same cycle forever. Before detection this burned the
	// whole 1000-cycle budget and reported "reached max allowed cycles".
	a, b := waiter("alpha", component.ErrWaitKeepingInputs), waiter("bravo", component.ErrWaitKeepingInputs)

	fm, err := fmesh.New("mutual-wait")
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(a, b))
	require.NoError(t, a.OutputByName("out").PipeTo(b.InputByName("in")))
	require.NoError(t, b.OutputByName("out").PipeTo(a.InputByName("in")))
	require.NoError(t, a.InputByName("in").PutSignals(signal.New(1)))

	ri, err := fm.Run(context.Background())

	require.ErrorIs(t, err, fmesh.ErrLivelockDetected)
	require.NotErrorIs(t, err, fmesh.ErrReachedMaxAllowedCycles,
		"the livelock must be reported as such, not as an exhausted cycle budget")
	assert.LessOrEqual(t, ri.Cycles.Len(), 4, "must stop promptly, not burn the cycle budget")

	// The message has to point at something actionable.
	assert.Contains(t, err.Error(), `"alpha" is waiting (keeping inputs)`)
	assert.Contains(t, err.Error(), "empty input ports []")
	assert.Contains(t, err.Error(), "holding signals on [in]")
	assert.Contains(t, err.Error(), "1 other component(s) never activated",
		"the component starving the other one is invisible otherwise")
	t.Logf("reported as:\n%v", err)
}

func TestLivelock_NamesTheUnfedPort(t *testing.T) {
	// A join that never gets its second operand: the mesh is stuck because "b"
	// was never wired to anything. The error should say so.
	joiner := testutil.MustComponent("joiner",
		component.WithInputs("a", "b"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			if !this.Inputs().AllHaveSignals() {
				return component.ErrWaitKeepingInputs
			}
			return this.OutputByName("out").PutSignals(signal.New("joined"))
		}))

	fm, err := fmesh.New("unfed-port")
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(joiner))
	require.NoError(t, joiner.InputByName("a").PutSignals(signal.New("A")))

	_, err = fm.Run(context.Background())

	require.ErrorIs(t, err, fmesh.ErrLivelockDetected)
	assert.Contains(t, err.Error(), "empty input ports [b]", "the never-fed port must be named")
	assert.Contains(t, err.Error(), "holding signals on [a]")
	t.Logf("reported as:\n%v", err)
}

func TestLivelock_NoFalsePositiveWhileAccumulating(t *testing.T) {
	// The false positive that matters: a component legitimately waiting across
	// several cycles while a hook feeds it one signal at a time. Signals keep
	// moving, so this is progress, not a stall — and it must run to completion.
	const needed = 6

	collector := testutil.MustComponent("collector",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			if this.InputByName("in").Signals().Len() < needed {
				return component.ErrWaitKeepingInputs
			}
			return this.OutputByName("out").PutSignals(signal.New("full"))
		}))

	fm, err := fmesh.New("accumulator")
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(collector))
	require.NoError(t, collector.InputByName("in").PutSignals(signal.New(0)))

	fed := 1
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeCycle(func(_ context.Context, _ *fmesh.CycleContext) error {
			if fed < needed {
				fed++
				return collector.InputByName("in").PutSignals(signal.New(fed))
			}
			return nil
		})
	})

	_, err = fm.Run(context.Background())

	require.NoError(t, err, "a mesh making progress must not be reported as livelocked")
	payload, err := collector.OutputByName("out").Signals().FirstPayload()
	require.NoError(t, err)
	assert.Equal(t, "full", payload)
}

func TestLivelock_DroppingWaitersStopNaturally(t *testing.T) {
	// A waiter that drops its inputs clears them, so next cycle it has nothing to
	// activate on and the mesh ends naturally. That is not a livelock and must
	// not be reported as one.
	c := waiter("dropper", component.ErrWaitDroppingInputs)

	fm, err := fmesh.New("dropping")
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))

	ri, err := fm.Run(context.Background())

	require.NoError(t, err, "a self-resolving wait is a natural stop, not a livelock")
	assert.LessOrEqual(t, ri.Cycles.Len(), 3)
}

func TestLivelock_ProgressingMeshIsUntouched(t *testing.T) {
	// A long but productive run: 200 cycles of real work, never flagged.
	counter := testutil.MustComponent("counter",
		component.WithInputs("in"),
		component.WithOutputs("out", "done"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			n := signal.AsOrDefault(this.InputByName("in").Signals().First(), 0)
			if n >= 200 {
				return this.OutputByName("done").PutSignals(signal.New(n))
			}
			return this.OutputByName("out").PutSignals(signal.New(n + 1))
		}))
	require.NoError(t, counter.LoopbackPipe("out", "in"))

	fm, err := fmesh.New("counting")
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(counter))
	require.NoError(t, counter.InputByName("in").PutSignals(signal.New(0)))

	_, err = fm.Run(context.Background())

	require.NoError(t, err)
	payload, err := counter.OutputByName("done").Signals().FirstPayload()
	require.NoError(t, err)
	assert.Equal(t, 200, payload)
}

func TestLivelock_ThresholdAndDisabling(t *testing.T) {
	newStuckMesh := func(t *testing.T, opts ...fmesh.Option) *fmesh.FMesh {
		t.Helper()
		a := waiter("stuck", component.ErrWaitKeepingInputs)
		require.NoError(t, a.LoopbackPipe("out", "in"))
		fm, err := fmesh.New("stuck", opts...)
		require.NoError(t, err)
		require.NoError(t, fm.AddComponents(a))
		require.NoError(t, a.InputByName("in").PutSignals(signal.New(1)))
		return fm
	}

	t.Run("a higher threshold takes more cycles to fire", func(t *testing.T) {
		low, err := newStuckMesh(t, fmesh.WithLivelockThreshold(2)).Run(context.Background())
		require.ErrorIs(t, err, fmesh.ErrLivelockDetected)

		high, err := newStuckMesh(t, fmesh.WithLivelockThreshold(9)).Run(context.Background())
		require.ErrorIs(t, err, fmesh.ErrLivelockDetected)

		assert.Greater(t, high.Cycles.Len(), low.Cycles.Len())
	})

	t.Run("detection can be turned off", func(t *testing.T) {
		fm := newStuckMesh(t, fmesh.WithoutLivelockDetection(), fmesh.WithCyclesLimit(20))
		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrReachedMaxAllowedCycles,
			"with detection off the mesh runs until a limit stops it")
		require.NotErrorIs(t, err, fmesh.ErrLivelockDetected)
	})

	t.Run("threshold must be positive", func(t *testing.T) {
		_, err := fmesh.New("bad", fmesh.WithLivelockThreshold(0))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "WithoutLivelockDetection")
	})
}
