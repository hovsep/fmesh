package fmesh

import (
	"context"
	"errors"
	"fmt"

	"github.com/hovsep/fmesh/cycle"
)

// mustStop defines when f-mesh must stop (it always checks only the last cycle).
//
// The order of the checks is load-bearing; see the comments on the ones where it
// is not obvious.
func (fm *FMesh) mustStop(ctx context.Context) (bool, error) {
	lastCycle := fm.runtimeInfo.Cycles.Last()

	// >= so the limit is the number of cycles that execute, not limit+1. The
	// HasActivatedComponents guard lets a mesh whose last allowed cycle was
	// already empty fall through to the natural stop instead of a false error.
	if (fm.config.CyclesLimit > 0) && (lastCycle.Number() >= fm.config.CyclesLimit) && lastCycle.HasActivatedComponents() {
		return fm.stop(ErrReachedMaxAllowedCycles)
	}

	if fm.config.TimeLimit > 0 && fm.runtimeInfo.Duration() >= fm.config.TimeLimit {
		return fm.stop(ErrTimeLimitExceeded)
	}

	// Before the error strategy: a canceled run makes activation functions return
	// ctx.Err(), which would otherwise be reported as ordinary activation errors
	// and hide why the mesh stopped.
	if err := fm.contextError(ctx); err != nil {
		return fm.stop(err)
	}

	// Before the natural stop check, so activation and hook errors are never
	// silently swallowed when nothing activated in the last cycle.
	if err := fm.strategyError(lastCycle); err != nil {
		return fm.stop(err)
	}

	if !lastCycle.HasActivatedComponents() {
		return fm.stop(nil)
	}

	// Components activated, but did the mesh actually move? After the natural stop
	// because a livelock is by definition a cycle that activated something, and
	// last because a real error is always the better explanation.
	if fm.livelock.detect(lastCycle) {
		return fm.stop(fm.livelock.error(lastCycle))
	}

	return false, nil
}

// stop logs why the run is ending and reports it. A nil err is a natural stop.
func (fm *FMesh) stop(err error) (bool, error) {
	if err == nil {
		fm.LogDebug("going to stop naturally")
	} else {
		fm.LogDebug("going to stop: %s", err)
	}
	return true, err
}

// strategyError returns the error that stops the mesh under the configured error
// handling strategy after lastCycle, or nil when the run may continue.
func (fm *FMesh) strategyError(lastCycle *cycle.Cycle) error {
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if lastCycle.HasActivationErrors() || lastCycle.HasActivationPanics() {
			return fmt.Errorf("%w, cycle # %d, %w",
				ErrHitAnErrorOrPanic, lastCycle.Number(), cycleFailures(lastCycle))
		}
	case StopOnFirstPanic:
		if lastCycle.HasActivationPanics() {
			return fmt.Errorf("%w, cycle # %d, %w",
				ErrHitAPanic, lastCycle.Number(), cycleFailures(lastCycle))
		}
	case IgnoreAll:
	}
	return nil
}

// cycleFailures joins the activation errors and panics of a cycle. Each part is
// guarded because a nil error passed to %w renders "%!w(<nil>)".
func cycleFailures(c *cycle.Cycle) error {
	var parts []error
	if activationErrors := c.AllErrorsCombined(); activationErrors != nil {
		parts = append(parts, fmt.Errorf("activation errors: %w", activationErrors))
	}
	if panics := c.AllPanicsCombined(); panics != nil {
		parts = append(parts, fmt.Errorf("activation panics: %w", panics))
	}
	return errors.Join(parts...)
}
