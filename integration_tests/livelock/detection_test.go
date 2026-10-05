// Package livelock covers the detector that ends a run which has stopped making
// progress.
//
// The interesting tests here are the negative ones. A detector that reports a
// livelock in a mesh that was merely slow to converge is worse than no detector
// at all: it kills correct runs, and it does so intermittently.
package livelock

import (
	"context"
	"fmt"
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
	assert.Contains(t, err.Error(), "1 other component(s) have no input signals",
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
	for _, threshold := range []int{1, 2} {
		c := waiter("dropper", component.ErrWaitDroppingInputs)

		// The pending count is taken before the drain clears the dropped
		// inputs, so with a threshold of 1 the first cycle used to look stalled.
		fm, err := fmesh.New("dropping", fmesh.WithLivelockThreshold(threshold))
		require.NoError(t, err)
		require.NoError(t, fm.AddComponents(c))
		require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))

		ri, err := fm.Run(context.Background())

		require.NoError(t, err, "a self-resolving wait is a natural stop, not a livelock (threshold %d)", threshold)
		assert.LessOrEqual(t, ri.Cycles.Len(), 3)
	}
}

func TestLivelock_ProgressingMeshIsUntouched(t *testing.T) {
	// A long but productive run: 200 cycles of real work, never flagged.
	counter := testutil.MustComponent("counter",
		component.WithInputs("in"),
		component.WithOutputs("out", "done"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			n := this.InputByName("in").Signals().FirstPayloadOrDefault(0)
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

func TestLivelock_NamesAFewWaitersAndCountsTheRest(t *testing.T) {
	// With many stuck components the message names enough to see the pattern
	// and counts the remainder, instead of listing every one.
	fm, err := fmesh.New("crowd")
	require.NoError(t, err)
	for _, name := range []string{"w1", "w2", "w3", "w4", "w5", "w6", "w7"} {
		w := waiter(name, component.ErrWaitKeepingInputs)
		require.NoError(t, fm.AddComponents(w))
		require.NoError(t, w.InputByName("in").PutSignals(signal.New(1)))
	}

	_, err = fm.Run(context.Background())

	require.ErrorIs(t, err, fmesh.ErrLivelockDetected)
	assert.Contains(t, err.Error(), `"w5" is waiting`)
	assert.NotContains(t, err.Error(), `"w6" is waiting`)
	assert.Contains(t, err.Error(), "...and 2 more waiting component(s)")
}

// emitter forwards one signal per activation, however many it received.
func emitter(name string) *component.Component {
	return testutil.MustComponent(name,
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutSignals(signal.New(1))
		}))
}

func TestLivelock_ThresholdIsExact(t *testing.T) {
	for _, threshold := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("stalled from the first cycle, threshold %d", threshold), func(t *testing.T) {
			stuck := waiter("stuck", component.ErrWaitKeepingInputs)
			fm, err := fmesh.New("seeded", fmesh.WithLivelockThreshold(threshold))
			require.NoError(t, err)
			require.NoError(t, fm.AddComponents(stuck))
			require.NoError(t, stuck.InputByName("in").PutSignals(signal.New(1)))

			ri, err := fm.Run(context.Background())

			require.ErrorIs(t, err, fmesh.ErrLivelockDetected)
			assert.Equal(t, threshold, ri.Cycles.Len())
		})

		t.Run(fmt.Sprintf("stalled after a productive cycle, threshold %d", threshold), func(t *testing.T) {
			// The source turns three signals into one, so the pending count
			// differs from the one the previous cycle left before its drain:
			// the stall must still be counted from its first cycle.
			src, stuck := emitter("src"), waiter("stuck", component.ErrWaitKeepingInputs)
			fm, err := fmesh.New("after-progress", fmesh.WithLivelockThreshold(threshold))
			require.NoError(t, err)
			require.NoError(t, fm.AddComponents(src, stuck))
			require.NoError(t, src.OutputByName("out").PipeTo(stuck.InputByName("in")))
			require.NoError(t, src.InputByName("in").PutSignals(signal.New(1), signal.New(2), signal.New(3)))

			ri, err := fm.Run(context.Background())

			require.ErrorIs(t, err, fmesh.ErrLivelockDetected)
			assert.Equal(t, 1+threshold, ri.Cycles.Len())
			assert.Contains(t, err.Error(), "1 other component(s) have no input signals",
				"src activated in cycle 1, so it must not be described as never activated")
		})
	}
}

func TestLivelock_ConsumingWaiterIsProgress(t *testing.T) {
	// A waiter that keeps its inputs but takes one signal per cycle is working
	// through a backlog, not stuck.
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
			in := this.InputByName("in")
			signals := in.Signals()
			if signals.Len() == 1 {
				return this.OutputByName("out").PutSignals(signal.New("drained"))
			}
			first := signals.First()
			if err := in.Clear(ctx); err != nil {
				return err
			}
			if err := in.PutSignalGroups(signals.Filter(func(s *signal.Signal) bool { return s != first })); err != nil {
				return err
			}
			return component.ErrWaitKeepingInputs
		}))

	fm, err := fmesh.New("backlog", fmesh.WithLivelockThreshold(1))
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(consumer))
	require.NoError(t, consumer.InputByName("in").PutSignals(signal.New(1), signal.New(2), signal.New(3), signal.New(4)))

	_, err = fm.Run(context.Background())

	require.NoError(t, err)
	assert.True(t, consumer.OutputByName("out").HasSignals())
}

func TestLivelock_StatefulWaiterIsReported(t *testing.T) {
	// The documented limit: detection assumes a waiter decides from its inputs
	// alone. One that proceeds on its own state looks stuck while it waits.
	newMesh := func(t *testing.T, opts ...fmesh.Option) (*fmesh.FMesh, *component.Component) {
		t.Helper()
		patient := testutil.MustComponent("patient",
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				waited, _ := this.State().GetOrDefault("waited", 0).(int)
				if waited < 3 {
					this.State().Set("waited", waited+1)
					return component.ErrWaitKeepingInputs
				}
				return this.OutputByName("out").PutSignals(signal.New("done"))
			}))
		fm, err := fmesh.New("patient", opts...)
		require.NoError(t, err)
		require.NoError(t, fm.AddComponents(patient))
		require.NoError(t, patient.InputByName("in").PutSignals(signal.New(1)))
		return fm, patient
	}

	fm, _ := newMesh(t)
	_, err := fm.Run(context.Background())
	require.ErrorIs(t, err, fmesh.ErrLivelockDetected)

	fm, patient := newMesh(t, fmesh.WithoutLivelockDetection())
	_, err = fm.Run(context.Background())
	require.NoError(t, err)
	assert.True(t, patient.OutputByName("out").HasSignals())
}

func TestLivelock_AnyOtherActivationIsNotAStall(t *testing.T) {
	// A stuck waiter next to a component that keeps working: the mesh as a
	// whole is busy, so only the cycle limit stops it.
	stuck, busy := waiter("stuck", component.ErrWaitKeepingInputs), emitter("busy")
	require.NoError(t, busy.LoopbackPipe("out", "in"))

	fm, err := fmesh.New("busy", fmesh.WithLivelockThreshold(1), fmesh.WithCyclesLimit(20))
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(stuck, busy))
	require.NoError(t, stuck.InputByName("in").PutSignals(signal.New(1)))
	require.NoError(t, busy.InputByName("in").PutSignals(signal.New(1)))

	_, err = fm.Run(context.Background())

	require.ErrorIs(t, err, fmesh.ErrReachedMaxAllowedCycles)
	require.NotErrorIs(t, err, fmesh.ErrLivelockDetected)
}

func TestLivelock_CyclesLimitWinsATie(t *testing.T) {
	stuck := waiter("stuck", component.ErrWaitKeepingInputs)
	fm, err := fmesh.New("tie", fmesh.WithCyclesLimit(2))
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(stuck))
	require.NoError(t, stuck.InputByName("in").PutSignals(signal.New(1)))

	_, err = fm.Run(context.Background())

	require.ErrorIs(t, err, fmesh.ErrReachedMaxAllowedCycles)
	require.NotErrorIs(t, err, fmesh.ErrLivelockDetected)
}

func TestLivelock_EachRunCountsFromZero(t *testing.T) {
	// Inputs survive a failed run, so a rerun stalls again, but it must take
	// the full threshold again rather than inherit the previous run's count.
	const threshold = 3
	stuck := waiter("stuck", component.ErrWaitKeepingInputs)
	fm, err := fmesh.New("rerun", fmesh.WithLivelockThreshold(threshold))
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(stuck))
	require.NoError(t, stuck.InputByName("in").PutSignals(signal.New(1)))

	for run := 1; run <= 2; run++ {
		ri, err := fm.Run(context.Background())
		require.ErrorIs(t, err, fmesh.ErrLivelockDetected)
		assert.Equal(t, threshold, ri.Cycles.Len(), "run %d", run)
	}
}
