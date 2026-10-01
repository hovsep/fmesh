package hooks

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
)

// Mesh hooks: run and cycle lifecycle, and component registration.

// twoCycleMesh runs for exactly two cycles: the producer activates in the first,
// the consumer in the second once the signal has been drained to it.
func twoCycleMesh(t *testing.T) *fmesh.FMesh {
	t.Helper()

	producer := testutil.MustComponent("producer",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutSignals(signal.New(1))
		}))
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	fm := testutil.MustFMesh("two-cycle")
	require.NoError(t, fm.AddComponents(producer, consumer))
	require.NoError(t, producer.OutputByName("out").PipeTo(consumer.InputByName("in")))
	require.NoError(t, producer.InputByName("in").PutSignals(signal.New(1)))
	return fm
}

func TestMeshHooks_RunAndCycleLifecycle(t *testing.T) {
	var log recorder
	var cycles int

	fm := twoCycleMesh(t)
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeRun(func(context.Context, *fmesh.FMesh) error {
			log.add("beforeRun")
			return nil
		})
		h.BeforeCycle(func(context.Context, *fmesh.CycleContext) error {
			log.add("beforeCycle")
			cycles++
			return nil
		})
		h.AfterCycle(func(context.Context, *fmesh.CycleContext) error {
			log.add("afterCycle")
			return nil
		})
		h.AfterRun(func(context.Context, *fmesh.FMesh) error {
			log.add("afterRun")
			return nil
		})
	})

	ri, err := fm.Run(context.Background())
	require.NoError(t, err)

	// Run hooks bracket the whole run; cycle hooks bracket each cycle.
	assert.Equal(t, "beforeRun", log.events[0])
	assert.Equal(t, "afterRun", log.events[len(log.events)-1])
	assert.Equal(t, ri.Cycles.Len(), cycles, "BeforeCycle fires once per cycle")
	assert.Equal(t, 1, countEvents(log.events, "beforeRun"))
	assert.Equal(t, 1, countEvents(log.events, "afterRun"))
	assert.Equal(t, cycles, countEvents(log.events, "afterCycle"))
}

func countEvents(events []string, want string) int {
	n := 0
	for _, e := range events {
		if e == want {
			n++
		}
	}
	return n
}

func TestMeshHooks_CycleContextCarriesResults(t *testing.T) {
	var lastNumber int
	var activatedInFirstCycle int

	fm := twoCycleMesh(t)
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.AfterCycle(func(_ context.Context, cycleCtx *fmesh.CycleContext) error {
			lastNumber = cycleCtx.Cycle.Number()
			if cycleCtx.Cycle.Number() == 1 {
				activatedInFirstCycle = cycleCtx.Cycle.ActivationResults().Len()
			}
			assert.Equal(t, "two-cycle", cycleCtx.FMesh.Name())
			return nil
		})
	})

	ri, err := fm.Run(context.Background())
	require.NoError(t, err)

	assert.Equal(t, ri.Cycles.Len(), lastNumber, "cycle numbers run 1..n")
	assert.Equal(t, 1, activatedInFirstCycle, "only the producer has input in cycle 1")
}

func TestMeshHooks_OnComponentAddedFiresPerComponent(t *testing.T) {
	var added []string

	fm := testutil.MustFMesh("registry")
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.OnComponentAdded(func(_ context.Context, ctx *fmesh.ComponentAddedContext) error {
			added = append(added, ctx.Component.Name())
			return nil
		})
	})

	c1 := testutil.MustComponent("c1", component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))
	c2 := testutil.MustComponent("c2", component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))
	require.NoError(t, fm.AddComponents(c1, c2))

	assert.Equal(t, []string{"c1", "c2"}, added)
}

func TestMeshHooks_AccumulateInRegistrationOrder(t *testing.T) {
	var log recorder

	fm := twoCycleMesh(t)
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeRun(func(context.Context, *fmesh.FMesh) error {
			log.add("first")
			return nil
		})
		h.BeforeRun(func(context.Context, *fmesh.FMesh) error {
			log.add("second")
			return nil
		})
	}).SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeRun(func(context.Context, *fmesh.FMesh) error {
			log.add("third")
			return nil
		})
	})

	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"first", "second", "third"}, log.events)
}

func TestMeshHooks_AfterRunFiresOnAFailedRun(t *testing.T) {
	var beforeRun, afterRun bool

	// A component whose activation errors fails the run under the default strategy.
	failing := testutil.MustComponent("failing",
		component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error {
			return errors.New("activation failed")
		}))
	fm := testutil.MustFMesh("failing-mesh")
	require.NoError(t, fm.AddComponents(failing))
	testutil.MustPutSignals(failing.InputByName("in"), signal.New(1))
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeRun(func(context.Context, *fmesh.FMesh) error {
			beforeRun = true
			return nil
		})
		h.AfterRun(func(context.Context, *fmesh.FMesh) error {
			afterRun = true
			return nil
		})
	})

	_, err := fm.Run(context.Background())

	require.Error(t, err)
	assert.True(t, beforeRun)
	assert.True(t, afterRun, "AfterRun runs in a defer, so a failed run still triggers it")
}

func TestMeshHooks_AfterRunContextOutlivesTheTimeLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		slow := testutil.MustComponent("slow",
			component.WithInputs("in"), component.WithOutputs("out"),
			component.WithActivationFunc(func(context.Context, *component.Component) error {
				time.Sleep(time.Second)
				return nil
			}))
		require.NoError(t, slow.LoopbackPipe("out", "in"))
		fm := testutil.MustFMesh("slow-mesh", fmesh.WithTimeLimit(500*time.Millisecond))
		require.NoError(t, fm.AddComponents(slow))
		testutil.MustPutSignals(slow.InputByName("in"), signal.New(1))
		var afterRunErr error
		fm.SetupHooks(func(h *fmesh.Hooks) {
			h.AfterRun(func(ctx context.Context, _ *fmesh.FMesh) error {
				afterRunErr = ctx.Err()
				return nil
			})
		})

		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrTimeLimitExceeded)
		assert.NoError(t, afterRunErr, "AfterRun can still do I/O after the time limit")
	})
}

func TestMeshHooks_BeforeRunFailureAbortsTheRun(t *testing.T) {
	var activated bool

	c := testutil.MustComponent("c", component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error {
			activated = true
			return nil
		}))

	fm := testutil.MustFMesh("gated")
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeRun(func(context.Context, *fmesh.FMesh) error {
			return errors.New("preflight failed")
		})
	})

	ri, err := fm.Run(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "preflight failed")
	assert.False(t, activated, "no component may activate when BeforeRun fails")
	assert.Zero(t, ri.Cycles.Len())
}

func TestMeshHooks_CycleHookFailuresAbortTheRun(t *testing.T) {
	t.Run("BeforeCycle", func(t *testing.T) {
		fm := twoCycleMesh(t)
		fm.SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeCycle(func(context.Context, *fmesh.CycleContext) error { return errors.New("no cycles today") })
		})

		ri, err := fm.Run(context.Background())

		require.ErrorContains(t, err, "beforeCycle hook failed")
		require.ErrorContains(t, err, "no cycles today")
		assert.Equal(t, 1, ri.Cycles.Len(), "the cycle that failed to start is still recorded")
	})

	t.Run("AfterCycle", func(t *testing.T) {
		fm := twoCycleMesh(t)
		fm.SetupHooks(func(h *fmesh.Hooks) {
			h.AfterCycle(func(context.Context, *fmesh.CycleContext) error { return errors.New("bookkeeping failed") })
		})

		_, err := fm.Run(context.Background())

		require.ErrorContains(t, err, "afterCycle hook failed")
		require.ErrorContains(t, err, "bookkeeping failed")
	})
}

func TestMeshHooks_AfterRunFailureBecomesTheRunError(t *testing.T) {
	// A run that would have succeeded reports the AfterRun failure; a run that
	// already failed keeps its own error (covered by AfterRunFiresOnAFailedRun).
	fm := twoCycleMesh(t)
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.AfterRun(func(context.Context, *fmesh.FMesh) error { return errors.New("cleanup failed") })
	})

	_, err := fm.Run(context.Background())

	require.ErrorContains(t, err, "afterRun hook failed")
	require.ErrorContains(t, err, "cleanup failed")
}
