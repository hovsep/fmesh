package constraints

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
)

// counterMesh is a component looping back into itself, so the mesh runs until
// something stops it. activations counts how many times it ran.
func counterMesh(t *testing.T, activations *atomic.Int64, work func(ctx context.Context) error, opts ...fmesh.Option) *fmesh.FMesh {
	t.Helper()

	c := testutil.MustComponent("counter",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
			activations.Add(1)
			if work != nil {
				if err := work(ctx); err != nil {
					return err
				}
			}
			return this.OutputByName("out").PutSignals(signal.New(1))
		}))
	require.NoError(t, c.LoopbackPipe("out", "in"))

	fm, err := fmesh.New("ctx-mesh", opts...)
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	return fm
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Run("canceling the context stops the mesh", func(t *testing.T) {
		var activations atomic.Int64
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel from the activation function itself once enough cycles have run,
		// which is the same path an external cancel takes: the mesh notices at the
		// next cycle boundary.
		fm := counterMesh(t, &activations, func(context.Context) error {
			if activations.Load() >= 5 {
				cancel()
			}
			return nil
		}, fmesh.WithUnlimitedCycles(), fmesh.WithUnlimitedTime())
		defer cancel()

		_, err := fm.Run(ctx)

		require.ErrorIs(t, err, fmesh.ErrRunCanceled)
		require.ErrorIs(t, err, context.Canceled, "the underlying context error must stay reachable")
		assert.Less(t, activations.Load(), int64(10), "the mesh must stop promptly, not run on")
	})

	t.Run("an already canceled context runs nothing", func(t *testing.T) {
		var activations atomic.Int64
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		fm := counterMesh(t, &activations, nil)
		ri, err := fm.Run(ctx)

		require.ErrorIs(t, err, fmesh.ErrRunCanceled)
		assert.Zero(t, activations.Load(), "no component may activate under a dead context")
		assert.Zero(t, ri.Cycles.Len())
	})

	t.Run("the caller's deadline is reported as cancellation, not as a time limit", func(t *testing.T) {
		var activations atomic.Int64
		// Caller's deadline is much shorter than the mesh time limit, so the mesh
		// must not claim its own limit was hit.
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		fm := counterMesh(t, &activations, func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Millisecond):
				return nil
			}
		}, fmesh.WithTimeLimit(30*time.Second), fmesh.WithUnlimitedCycles())

		_, err := fm.Run(ctx)

		require.ErrorIs(t, err, fmesh.ErrRunCanceled)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		assert.NotErrorIs(t, err, fmesh.ErrTimeLimitExceeded)
	})
}

func TestRun_TimeLimitReachesActivationFunctions(t *testing.T) {
	// The point of deriving the time limit as a context deadline: an activation
	// function that respects its context is interrupted by the limit, instead of
	// the limit only being noticed after the blocking call returns.
	const timeLimit = 100 * time.Millisecond

	var observedDeadline atomic.Bool
	var activations atomic.Int64

	fm := counterMesh(t, &activations, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			observedDeadline.Store(true)
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return errors.New("activation was not interrupted by the mesh time limit")
		}
	}, fmesh.WithTimeLimit(timeLimit), fmesh.WithUnlimitedCycles())

	start := time.Now()
	_, err := fm.Run(context.Background())
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.True(t, observedDeadline.Load(), "the activation function must see the deadline through its context")
	assert.Less(t, elapsed, time.Second,
		"the mesh must stop near its time limit, not wait for the blocking call")
	assert.ErrorIs(t, err, fmesh.ErrTimeLimitExceeded)
}

func TestRun_ContextIsIgnorable(t *testing.T) {
	// Ignoring the context stays valid: the mesh still finishes normally.
	var activations atomic.Int64
	fm := counterMesh(t, &activations, nil, fmesh.WithCyclesLimit(3))

	_, err := fm.Run(context.Background())

	require.ErrorIs(t, err, fmesh.ErrReachedMaxAllowedCycles)
	assert.Positive(t, activations.Load())
}

func TestRun_HooksReceiveTheRunContext(t *testing.T) {
	type seen struct {
		beforeRun, beforeCycle, afterCycle, afterRun, beforeActivation, afterActivation atomic.Bool
	}
	var got seen

	key := struct{ name string }{"probe"}
	ctx := context.WithValue(context.Background(), key, "carried")

	carries := func(ctx context.Context) bool { return ctx.Value(key) == "carried" }

	c := testutil.MustComponent("c",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
			require.True(t, carries(ctx), "activation function must receive the run context")
			return this.OutputByName("out").PutSignals(signal.New(1))
		}))
	c.SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(ctx context.Context, _ *component.Component) error {
			got.beforeActivation.Store(carries(ctx))
			return nil
		})
		h.AfterActivation(func(ctx context.Context, _ *component.ActivationContext) error {
			got.afterActivation.Store(carries(ctx))
			return nil
		})
	})

	fm, err := fmesh.New("hooks-ctx")
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(c))
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeRun(func(ctx context.Context, _ *fmesh.FMesh) error {
			got.beforeRun.Store(carries(ctx))
			return nil
		})
		h.AfterRun(func(ctx context.Context, _ *fmesh.FMesh) error {
			got.afterRun.Store(carries(ctx))
			return nil
		})
		h.BeforeCycle(func(ctx context.Context, _ *fmesh.CycleContext) error {
			got.beforeCycle.Store(carries(ctx))
			return nil
		})
		h.AfterCycle(func(ctx context.Context, _ *fmesh.CycleContext) error {
			got.afterCycle.Store(carries(ctx))
			return nil
		})
	})
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))

	_, err = fm.Run(ctx)
	require.NoError(t, err)

	assert.True(t, got.beforeRun.Load(), "BeforeRun")
	assert.True(t, got.beforeCycle.Load(), "BeforeCycle")
	assert.True(t, got.afterCycle.Load(), "AfterCycle")
	assert.True(t, got.afterRun.Load(), "AfterRun")
	assert.True(t, got.beforeActivation.Load(), "BeforeActivation")
	assert.True(t, got.afterActivation.Load(), "AfterActivation")
}
