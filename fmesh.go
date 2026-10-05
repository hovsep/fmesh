package fmesh

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/internal/plugin"
	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/port"
)

// Option is a functional option for configuring an FMesh during construction.
type Option func(*FMesh) error

// FMesh is the functional mesh.
type FMesh struct {
	name        string
	description string
	meta        *meta.Meta
	components  *component.Collection
	runtimeInfo *RuntimeInfo
	logger      *log.Logger
	config      config
	hooks       *Hooks
	plugins     *plugin.Registry[*FMesh]

	// Livelock bookkeeping for the current run; reset by cleanUpPreviousRun.
	stalledCycles  int
	pendingSignals int
}

// New creates a new F-Mesh with the default configuration and applies any provided options.
func New(name string, opts ...Option) (*FMesh, error) {
	fm := &FMesh{
		name:        name,
		description: "",
		meta:        meta.New(),
		components:  component.NewCollection(),
		logger:      newDefaultLogger(name),
		config:      newDefaultConfig(),
		hooks:       newHooks(),
		plugins:     newPlugins(),
	}
	for _, opt := range opts {
		if err := opt(fm); err != nil {
			return nil, fmt.Errorf("fmesh %q option failed: %w", name, err)
		}
	}
	// Built after the options so the runtime info picks up the configured
	// history limit (Run rebuilds it the same way).
	fm.runtimeInfo = newRuntimeInfo(fm.config.CyclesHistoryLimit)

	// Last, so a plugin sees the fully configured mesh and can rely on anything
	// the options set up.
	if err := fm.initPlugins(); err != nil {
		return nil, err
	}
	return fm, nil
}

// Name returns the name of the F-Mesh.
func (fm *FMesh) Name() string {
	return fm.name
}

// Description returns the description of the F-Mesh.
func (fm *FMesh) Description() string {
	return fm.description
}

// Components returns all components in the mesh.
func (fm *FMesh) Components() *component.Collection {
	return fm.components
}

// ComponentByName returns a component by name.
func (fm *FMesh) ComponentByName(name string) *component.Component {
	return fm.Components().ByName(name)
}

// WithDescription is a constructor option that sets a description on the mesh.
func WithDescription(description string) Option {
	return func(fm *FMesh) error {
		fm.description = description
		return nil
	}
}

// Meta returns the mesh's metadata store.
func (fm *FMesh) Meta() *meta.Meta {
	return fm.meta
}

// WithMeta is a constructor option that adds or updates one metadata entry on the mesh.
func WithMeta[T meta.Value](key string, value T) Option {
	return func(fm *FMesh) error {
		fm.meta.Set(key, value)
		return nil
	}
}

// AddComponents adds components to the mesh. Returns an error if any component is invalid or has a
// duplicate name; then no component is added or changed. The OnComponentAdded hooks run after all
// components are added, for every component even when one fails, and a failing hook leaves them
// added.
func (fm *FMesh) AddComponents(components ...*component.Component) error {
	for _, c := range components {
		if err := c.ValidateBeforeAddingToMesh(); err != nil {
			return fmt.Errorf("failed to add component %q: %w", c.Name(), err)
		}
	}

	if err := fm.components.Add(components...); err != nil {
		return fmt.Errorf("failed to add components to mesh: %w", err)
	}

	for _, c := range components {
		c.SetParentMesh(fm)
		c.InheritLogger(fm.logger)
	}

	var hookErrs []error
	for _, c := range components {
		// Components are added outside a run, so there is no run context yet.
		if err := fm.hooks.onComponentAdded.TriggerAll(context.Background(), &ComponentAddedContext{FMesh: fm, Component: c}); err != nil {
			hookErrs = append(hookErrs, fmt.Errorf("onComponentAdded hook failed for component %q: %w", c.Name(), err))
		}
	}

	fm.LogDebug("%d components added to mesh", fm.Components().Len())
	return errors.Join(hookErrs...)
}

// SetupHooks configures hooks for the mesh using a closure.
func (fm *FMesh) SetupHooks(configure func(*Hooks)) *FMesh {
	configure(fm.hooks)
	return fm
}

// runCycle runs one activation cycle (tries to activate ready components).
// Returns any error that occurred.
// The cycle is always added to runtimeInfo even if an error occurred.
func (fm *FMesh) runCycle(ctx context.Context) (err error) {
	nextNumber := 1
	if lastCycle := fm.runtimeInfo.Cycles.Last(); lastCycle != nil {
		nextNumber = lastCycle.Number() + 1
	}
	newCycle := cycle.New().SetNumber(nextNumber)

	// Recorded however the cycle ends, so a failed cycle still shows up in the
	// run history. Deferred rather than repeated at each exit; it runs after the
	// afterCycle hook, which therefore sees the cycle in its CycleContext but not
	// yet in RuntimeInfo.Cycles.
	defer fm.runtimeInfo.Cycles.Add(newCycle)

	// AfterCycle runs however the cycle ends, like AfterRun and AfterActivation,
	// so a hook that pairs with BeforeCycle always sees the end of its cycle.
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

	// ForEach avoids cloning the component slice on every cycle (hot path)
	_ = fm.Components().ForEach(func(c *component.Component) error {
		wg.Go(func() {
			ar := c.MaybeActivate(ctx)
			// Components with no input produce pure noise in runtime info (in sparse
			// meshes it's most of the history), so their result is never recorded. A
			// missing result means "had no input"; the run loop treats absent results
			// as not-activated. Only NoInput is skipped here — WaitingForInputs*,
			// errors, panics, and HookFailed results are still recorded.
			if ar.Code() == component.ActivationCodeNoInput {
				return
			}
			newCycle.AddActivationResults(ar)
		})
		return nil
	})

	wg.Wait()

	return nil
}

// forEachActivatedComponent applies action to every component that activated in
// the last cycle, paired with its activation result, and stops at the first error.
// It walks the name-ordered components rather than the map-backed activation
// results: drain order decides fan-in order and must be deterministic. A
// component with no result did not activate at all.
func (fm *FMesh) forEachActivatedComponent(
	components []*component.Component,
	action func(*component.Component, *component.ActivationResult) error,
) error {
	results := fm.runtimeInfo.Cycles.Last().ActivationResults()
	for _, c := range components {
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

// drainComponents drains the data from activated components.
// Components are processed in name order so fan-in signal order is deterministic.
//
// Every input is cleared before any output is flushed, and the two passes cannot
// be merged into one: flushing delivers signals into downstream input ports, so
// a single pass would clear away what an earlier component had just delivered.
func (fm *FMesh) drainComponents(ctx context.Context) error {
	components := fm.Components().AllOrdered()

	if err := fm.clearInputs(ctx, components); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToDrain, err)
	}

	return fm.forEachActivatedComponent(components, func(c *component.Component, activationResult *component.ActivationResult) error {
		// Components waiting for inputs are never drained
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
func (fm *FMesh) clearInputs(ctx context.Context, components []*component.Component) error {
	return fm.forEachActivatedComponent(components, func(c *component.Component, activationResult *component.ActivationResult) error {
		if component.WantsToKeepInputs(activationResult) {
			// Component wants to keep inputs for the next cycle
			return nil
		}

		if err := c.ClearInputs(ctx); err != nil {
			return fmt.Errorf("failed to clear input ports: component %q: %w", c.Name(), err)
		}
		return nil
	})
}

func (fm *FMesh) cleanUpPreviousRun(ctx context.Context) error {
	// Clear all output ports to prevent signal accumulation between runs
	if err := fm.Components().ForEach(func(c *component.Component) error {
		if err := c.ClearOutputs(ctx); err != nil {
			return fmt.Errorf("failed to clear outputs of component %q: %w", c.Name(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	// Init runtime info
	fm.runtimeInfo = newRuntimeInfo(fm.config.CyclesHistoryLimit)
	fm.runtimeInfo.markStarted()

	// Seeded inputs are the baseline the first cycle is compared against, so a
	// mesh that stalls immediately is caught in threshold cycles rather than
	// threshold+1.
	fm.stalledCycles = 0
	fm.pendingSignals = fm.countPendingSignals()
	return nil
}

// countPendingSignals totals the signals sitting on input ports across the mesh.
// It is the cheapest thing that answers "did anything move?" — see detectLivelock.
func (fm *FMesh) countPendingSignals() int {
	total := 0
	_ = fm.Components().ForEach(func(c *component.Component) error {
		return c.Inputs().ForEach(func(p *port.Port) error {
			total += p.Signals().Len()
			return nil
		})
	})
	return total
}

// detectLivelock reports whether the mesh has stopped making progress, and
// updates the stall bookkeeping. Called once per cycle.
//
// A cycle is stalled when every component that activated was waiting and keeping
// its inputs *and* the pending signal count is unchanged — both halves are
// needed: the first alone flags legitimate input accumulation, the second alone
// flags a busy-but-idempotent mesh. A waiter that drops its inputs never stalls:
// the count is taken before the drain clears them, so it would look unchanged.
func (fm *FMesh) detectLivelock(lastCycle *cycle.Cycle) bool {
	if fm.config.LivelockThreshold <= 0 {
		return false
	}

	pending := fm.countPendingSignals()
	allKeeping := lastCycle.HasActivatedComponents() && lastCycle.ActivationResults().Every(component.WantsToKeepInputs)
	if allKeeping && pending == fm.pendingSignals {
		fm.stalledCycles++
	} else {
		fm.stalledCycles = 0
	}
	fm.pendingSignals = pending

	return fm.stalledCycles >= fm.config.LivelockThreshold
}

// cycleFailures describes what went wrong in a cycle, joining only the parts
// that actually happened: a nil error passed to %w renders "%!w(<nil>)", so
// each part is guarded even though callers only arrive with at least one.
func cycleFailures(c *cycle.Cycle) error {
	var parts []error
	if activationErrors := c.AllErrorsCombined(); activationErrors != nil {
		parts = append(parts, fmt.Errorf("activation errors: %w", activationErrors))
	}
	if panics := c.AllPanicsCombined(); panics != nil {
		parts = append(parts, fmt.Errorf("activation panics: %w", panics))
	}
	if len(parts) == 0 {
		return errors.New("no activation errors or panics recorded")
	}
	return errors.Join(parts...)
}

// livelockError explains the stall by naming who is stuck and on what: the
// empty input ports of each waiting component point at the pipe that was never
// wired.
func (fm *FMesh) livelockError(lastCycle *cycle.Cycle) error {
	// Name enough components to see the pattern; count the rest.
	const maxNamed = 5

	var detail strings.Builder
	starved, named, waiting := 0, 0, 0
	for _, c := range fm.Components().AllOrdered() {
		activationResult := lastCycle.ActivationResults().ByName(c.Name())
		if activationResult == nil || !component.IsWaitingForInput(activationResult) {
			// A component with no input is not recorded at all. Counting them is
			// worth the line: in a mutual wait only the component that happens to
			// hold the signal shows up above, and the one starving it is invisible.
			if activationResult == nil {
				starved++
			}
			continue
		}

		var empty, holding []string
		for _, p := range c.Inputs().AllOrdered() {
			if p.HasSignals() {
				holding = append(holding, p.Name())
			} else {
				empty = append(empty, p.Name())
			}
		}

		waiting++
		if named >= maxNamed {
			continue
		}
		named++

		mode := "dropping inputs"
		if component.WantsToKeepInputs(activationResult) {
			mode = "keeping inputs"
		}
		fmt.Fprintf(&detail, "\n  %q is waiting (%s): empty input ports %v, holding signals on %v",
			c.Name(), mode, empty, holding)
	}

	if waiting > named {
		fmt.Fprintf(&detail, "\n  ...and %d more waiting component(s)", waiting-named)
	}

	if starved > 0 {
		fmt.Fprintf(&detail, "\n  %d other component(s) never activated: no signals ever reached them", starved)
	}

	return fmt.Errorf("%w: no progress for %d consecutive cycles, stopped at cycle #%d. "+
		"Every component that activated is waiting for inputs and no signals moved, "+
		"so every following cycle would be identical to this one%s",
		ErrLivelockDetected, fm.stalledCycles, lastCycle.Number(), detail.String())
}

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
	}
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
	if fm.detectLivelock(lastCycle) {
		return fm.stop(fm.livelockError(lastCycle))
	}

	return false, nil
}
