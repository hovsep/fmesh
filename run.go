package fmesh

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
)

// Run executes the mesh, activating components until completion or cycle limit.
//
// The context is passed to every activation function and hook, and to the drain
// that moves signals between cycles. Canceling it stops the mesh at the next
// cycle boundary with ErrRunCanceled; a running cycle is never interrupted,
// because Go cannot preempt a goroutine — an activation function that ignores
// its context blocks the mesh for as long as it runs.
//
// A configured TimeLimit becomes a deadline on this context, so it reaches
// activation functions instead of only being checked between cycles.
//
// Run is not reentrant: it resets per-run state on the mesh, so a mesh must
// finish one Run before starting another. Between runs, output ports are
// cleared but input ports are not — signals a failed or interrupted run left
// on inputs become part of the next run's seed.
func (fm *FMesh) Run(ctx context.Context) (ri *RuntimeInfo, runErr error) {
	if err := fm.cleanUpPreviousRun(ctx); err != nil {
		return nil, err
	}

	// AfterRun gets the caller's context: once the time limit has fired, the run
	// context is done, and an AfterRun that exports or flushes would fail at once.
	callerCtx := ctx

	// After cleanUpPreviousRun starts the run clock, never before: contextError
	// tells the time limit from a caller's cancellation by the run's duration,
	// so the deadline must not fire before that duration reaches the limit.
	if fm.config.TimeLimit > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, fm.config.TimeLimit)
		defer cancel()
	}

	ri = fm.runtimeInfo

	defer func() {
		fm.runtimeInfo.markStopped()
		if err := fm.hooks.afterRun.TriggerAll(callerCtx, fm); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("afterRun hook failed: %w", err))
		}
		// The history now belongs to the caller: holding it here would keep a
		// discarded run's cycles alive until the next Run.
		fm.runtimeInfo = newRuntimeInfo(fm.config.CyclesHistoryLimit)
	}()

	if err := fm.hooks.beforeRun.Trigger(ctx, fm); err != nil {
		return ri, fmt.Errorf("beforeRun hook failed: %w", err)
	}

	// An already-canceled context must not run a single cycle.
	if err := fm.contextError(ctx); err != nil {
		return ri, err
	}

	for {
		if err := fm.runCycle(ctx); err != nil {
			return ri, err
		}

		if mustStop, err := fm.mustStop(ctx); mustStop {
			return ri, err
		}

		if err := fm.drainComponents(ctx); err != nil {
			return ri, err
		}

		fm.livelock.takeBaseline()
	}
}

func (fm *FMesh) cleanUpPreviousRun(ctx context.Context) error {
	for c := range fm.components.Each {
		if err := c.ClearOutputs(ctx); err != nil {
			return fmt.Errorf("failed to clear outputs of component %q: %w", c.Name(), err)
		}
	}

	fm.runtimeInfo = newRuntimeInfo(fm.config.CyclesHistoryLimit)
	fm.runtimeInfo.markStarted()

	fm.livelock = newLivelockDetector(fm.components, fm.config.LivelockThreshold)
	return nil
}

// contextError translates a canceled context into a mesh error.
//
// A deadline can come from two places: the mesh time limit, which Run turns into
// a deadline of its own, and the caller's own context. They are told apart by the
// elapsed time — the mesh's own limit cannot fire before it has elapsed — so a
// caller passing a shorter deadline is reported as a cancellation rather than
// being blamed on a time limit it did not hit.
func (fm *FMesh) contextError(ctx context.Context) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) &&
		fm.config.TimeLimit > 0 && fm.runtimeInfo.Duration() >= fm.config.TimeLimit {
		return ErrTimeLimitExceeded
	}

	return fmt.Errorf("%w: %w", ErrRunCanceled, err)
}

// runCycle activates every component that has input, concurrently, and records
// the cycle in runtimeInfo however it ends.
func (fm *FMesh) runCycle(ctx context.Context) (err error) {
	nextNumber := 1
	if lastCycle := fm.runtimeInfo.Cycles.Last(); lastCycle != nil {
		nextNumber = lastCycle.Number() + 1
	}
	newCycle := cycle.New().SetNumber(nextNumber)

	// Runs after the afterCycle hook, which therefore sees the cycle in its
	// CycleContext but not yet in RuntimeInfo.Cycles.
	defer fm.runtimeInfo.Cycles.Add(newCycle)

	// AfterCycle runs however the cycle ends, so it always pairs with BeforeCycle.
	defer func() {
		if hookErr := fm.hooks.afterCycle.TriggerAll(ctx, &CycleContext{FMesh: fm, Cycle: newCycle}); hookErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to run cycle: afterCycle hook failed: %w", hookErr))
		}
	}()

	if err := fm.hooks.beforeCycle.Trigger(ctx, &CycleContext{FMesh: fm, Cycle: newCycle}); err != nil {
		return fmt.Errorf("failed to run cycle: beforeCycle hook failed: %w", err)
	}

	fm.LogDebug("starting activation cycle #%d", newCycle.Number())

	var wg sync.WaitGroup

	for c := range fm.components.Each {
		// No goroutine for a component with nothing to read: in a sparse mesh
		// that is most of them, and MaybeActivate would only report NoInput.
		if !c.Inputs().AnyHasSignals() {
			continue
		}
		wg.Go(func() {
			ar := c.MaybeActivate(ctx)
			// A missing result means "had no input": recording NoInput would fill
			// the history of a sparse mesh with noise.
			if ar.Code() == component.ActivationCodeNoInput {
				return
			}
			newCycle.AddActivationResults(ar)
		})
	}

	wg.Wait()

	return nil
}

// drainComponents drains the data from activated components.
// Components are processed in name order so fan-in signal order is deterministic.
//
// Every input is cleared before any output is flushed, and the two passes cannot
// be merged into one: flushing delivers signals into downstream input ports, so
// a single pass would clear away what an earlier component had just delivered.
func (fm *FMesh) drainComponents(ctx context.Context) error {
	if err := fm.clearInputs(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToDrain, err)
	}

	return fm.forEachActivatedComponent(func(c *component.Component, activationResult *component.ActivationResult) error {
		if component.IsWaitingForInput(activationResult) {
			return nil
		}

		if err := c.FlushOutputs(ctx); err != nil {
			return fmt.Errorf("%w: failed to flush outputs of component %q: %w", ErrFailedToDrain, c.Name(), err)
		}
		return nil
	})
}

// clearInputs clears all the input ports of all components activated in the latest cycle.
func (fm *FMesh) clearInputs(ctx context.Context) error {
	return fm.forEachActivatedComponent(func(c *component.Component, activationResult *component.ActivationResult) error {
		if component.WantsToKeepInputs(activationResult) {
			return nil
		}

		if err := c.ClearInputs(ctx); err != nil {
			return fmt.Errorf("failed to clear input ports: component %q: %w", c.Name(), err)
		}
		return nil
	})
}

// forEachActivatedComponent applies action to every component that activated in
// the last cycle, paired with its activation result, and stops at the first error.
// It walks the name-ordered components rather than the map-backed activation
// results: drain order decides fan-in order and must be deterministic. A
// component with no result did not activate at all.
func (fm *FMesh) forEachActivatedComponent(
	action func(*component.Component, *component.ActivationResult) error,
) error {
	results := fm.runtimeInfo.Cycles.Last().ActivationResults()
	for c := range fm.components.Each {
		activationResult := results.ByName(c.Name())
		if activationResult == nil || !activationResult.Activated() {
			continue
		}
		if err := action(c, activationResult); err != nil {
			return err
		}
	}
	return nil
}
