package fmesh

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/internal/plugin"
	"github.com/hovsep/fmesh/meta"
)

// Option is a functional option for configuring an FMesh during construction.
type Option func(*FMesh) error

// FMesh is the functional mesh.
type FMesh struct {
	name        string
	description string
	labels      *meta.Labels
	scalars     *meta.Scalars
	components  *component.Collection
	runtimeInfo *RuntimeInfo
	logger      *log.Logger
	config      Config
	hooks       *Hooks
	plugins     *plugin.Registry[*FMesh]
}

// New creates a new F-Mesh with the default configuration and applies any provided options.
func New(name string, opts ...Option) (*FMesh, error) {
	fm := &FMesh{
		name:        name,
		description: "",
		labels:      meta.NewLabels(),
		scalars:     meta.NewScalars(),
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

// Labels returns the mesh's labels store.
func (fm *FMesh) Labels() *meta.Labels {
	return fm.labels
}

// SetLabels replaces all labels.
func (fm *FMesh) SetLabels(labelMap map[string]string) *FMesh {
	fm.labels.Clear().SetMany(labelMap)
	return fm
}

// AddLabels adds or updates labels.
func (fm *FMesh) AddLabels(labelMap map[string]string) *FMesh {
	fm.labels.SetMany(labelMap)
	return fm
}

// AddLabel adds or updates a single label.
func (fm *FMesh) AddLabel(name, value string) *FMesh {
	fm.labels.Set(name, value)
	return fm
}

// ClearLabels removes all labels.
func (fm *FMesh) ClearLabels() *FMesh {
	fm.labels.Clear()
	return fm
}

// RemoveLabels removes specific labels.
func (fm *FMesh) RemoveLabels(names ...string) *FMesh {
	fm.labels.Remove(names...)
	return fm
}

// Scalars returns the mesh's scalars store.
func (fm *FMesh) Scalars() *meta.Scalars {
	return fm.scalars
}

// SetScalars replaces all scalars.
func (fm *FMesh) SetScalars(scalarsMap map[string]float64) *FMesh {
	fm.scalars.Clear().SetMany(scalarsMap)
	return fm
}

// AddScalars adds or updates scalars.
func (fm *FMesh) AddScalars(scalarsMap map[string]float64) *FMesh {
	fm.scalars.SetMany(scalarsMap)
	return fm
}

// AddScalar adds or updates a single scalar.
func (fm *FMesh) AddScalar(name string, value float64) *FMesh {
	fm.scalars.Set(name, value)
	return fm
}

// ClearScalars removes all scalars.
func (fm *FMesh) ClearScalars() *FMesh {
	fm.scalars.Clear()
	return fm
}

// RemoveScalars removes specific scalars.
func (fm *FMesh) RemoveScalars(names ...string) *FMesh {
	fm.scalars.Remove(names...)
	return fm
}

// WithLabel is a constructor option that adds or updates a single label on the mesh.
func WithLabel(name, value string) Option {
	return func(fm *FMesh) error {
		fm.labels.Set(name, value)
		return nil
	}
}

// WithScalar is a constructor option that adds or updates a single scalar on the mesh.
func WithScalar(name string, value float64) Option {
	return func(fm *FMesh) error {
		fm.scalars.Set(name, value)
		return nil
	}
}

// AddComponents adds components to the mesh. Returns an error if any component is invalid or has a duplicate name.
func (fm *FMesh) AddComponents(components ...*component.Component) error {
	for _, c := range components {
		if err := c.ValidateBeforeAddingToMesh(); err != nil {
			return fmt.Errorf("failed to add component %q: %w", c.Name(), err)
		}

		c.SetParentMesh(fm)
		c.InheritLogger(fm.logger)

		if err := fm.components.Add(c); err != nil {
			return fmt.Errorf("failed to add component %q to mesh: %w", c.Name(), err)
		}

		// Components are added outside a run, so there is no run context yet.
		if err := fm.hooks.onComponentAdded.Trigger(context.Background(), &ComponentAddedContext{FMesh: fm, Component: c}); err != nil {
			return fmt.Errorf("onComponentAdded hook failed for component %q: %w", c.Name(), err)
		}
	}

	fm.LogDebug("%d components added to mesh", fm.Components().Len())
	return nil
}

// SetupHooks configures hooks for the mesh using a closure.
func (fm *FMesh) SetupHooks(configure func(*Hooks)) *FMesh {
	configure(fm.hooks)
	return fm
}

// runCycle runs one activation cycle (tries to activate ready components).
// Returns any error that occurred.
// The cycle is always added to runtimeInfo even if an error occurred.
func (fm *FMesh) runCycle(ctx context.Context) error {
	nextNumber := 1
	if lastCycle := fm.runtimeInfo.Cycles.Last(); lastCycle != nil {
		nextNumber = lastCycle.Number() + 1
	}
	newCycle := cycle.New().SetNumber(nextNumber)

	if err := fm.hooks.beforeCycle.Trigger(ctx, &CycleContext{FMesh: fm, Cycle: newCycle}); err != nil {
		cycleErr := errors.Join(errFailedToRunCycle, fmt.Errorf("beforeCycle hook failed: %w", err))
		fm.runtimeInfo.Cycles.Add(newCycle)
		return cycleErr
	}

	fm.LogDebug("starting activation cycle #%d", newCycle.Number())

	if fm.Components().IsEmpty() {
		fm.runtimeInfo.Cycles.Add(newCycle)
		return errors.Join(errFailedToRunCycle, errNoComponents)
	}

	var wg sync.WaitGroup

	// ForEach avoids cloning the component map on every cycle (hot path)
	_ = fm.Components().ForEach(func(c *component.Component) error {
		wg.Add(1)
		go func(comp *component.Component, cyc *cycle.Cycle) {
			defer wg.Done()
			ar := comp.MaybeActivate(ctx)
			// Components with no input produce pure noise in runtime info (in sparse
			// meshes it's most of the history), so their result is never recorded. A
			// missing result means "had no input"; the run loop treats absent results
			// as not-activated. Only NoInput is skipped here — WaitingForInputs*,
			// errors, panics, and HookFailed results are still recorded.
			if ar.Code() == component.ActivationCodeNoInput {
				return
			}
			cyc.AddActivationResults(ar)
		}(c, newCycle)
		return nil
	})

	wg.Wait()

	if fm.IsDebug() {
		_ = newCycle.ActivationResults().ForEach(func(ar *component.ActivationResult) error {
			fm.LogDebug("activation result for component %s: activated: %t, code: %s, is error: %t, is panic: %t, error: %v",
				ar.ComponentName(), ar.Activated(), ar.Code(), ar.IsError(), ar.IsPanic(), ar.ActivationError())
			return nil
		})
	}

	if err := fm.hooks.afterCycle.Trigger(ctx, &CycleContext{FMesh: fm, Cycle: newCycle}); err != nil {
		cycleErr := errors.Join(errFailedToRunCycle, fmt.Errorf("afterCycle hook failed: %w", err))
		fm.runtimeInfo.Cycles.Add(newCycle)
		return cycleErr
	}

	fm.runtimeInfo.Cycles.Add(newCycle)
	return nil
}

// drainComponents drains the data from activated components.
// Components are processed in name order so fan-in signal order is deterministic.
func (fm *FMesh) drainComponents(ctx context.Context) error {
	components := fm.Components().AllOrdered()

	if err := fm.clearInputs(ctx, components); err != nil {
		return errors.Join(ErrFailedToDrain, err)
	}

	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByName(c.Name())

		if activationResult == nil || !activationResult.Activated() {
			// Component did not activate, so it did not create new output signals, hence nothing to drain
			continue
		}

		// Components waiting for inputs are never drained
		if component.IsWaitingForInput(activationResult) {
			continue
		}

		if err := c.FlushOutputs(ctx); err != nil {
			return errors.Join(ErrFailedToDrain, fmt.Errorf("failed to flush outputs of component %q: %w", c.Name(), err))
		}
	}
	return nil
}

// clearInputs clears all the input ports of all components activated in the latest cycle.
func (fm *FMesh) clearInputs(ctx context.Context, components []*component.Component) error {
	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByName(c.Name())

		if activationResult == nil || !activationResult.Activated() {
			// Component did not activate hence its inputs must be clear
			continue
		}

		if component.IsWaitingForInput(activationResult) && component.WantsToKeepInputs(activationResult) {
			// Component wants to keep inputs for the next cycle
			continue
		}

		if err := c.ClearInputs(ctx); err != nil {
			return errors.Join(errFailedToClearInputs, fmt.Errorf("component %q: %w", c.Name(), err))
		}
	}
	return nil
}

func (fm *FMesh) cleanUpPreviousRun(ctx context.Context) error {
	// Clear all output ports to prevent signal accumulation between runs
	if err := fm.Components().ForEach(func(c *component.Component) error {
		return c.ClearOutputs(ctx)
	}); err != nil {
		return err
	}

	// Init runtime info
	fm.runtimeInfo = newRuntimeInfo(fm.config.CyclesHistoryLimit)
	fm.runtimeInfo.markStarted()
	return nil
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
func (fm *FMesh) Run(ctx context.Context) (ri *RuntimeInfo, runErr error) {
	if fm.config.TimeLimit > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, fm.config.TimeLimit)
		defer cancel()
	}

	if err := fm.cleanUpPreviousRun(ctx); err != nil {
		return nil, err
	}

	ri = fm.runtimeInfo

	defer func() {
		fm.runtimeInfo.markStopped()
		if err := fm.hooks.afterRun.Trigger(ctx, fm); err != nil {
			if runErr == nil {
				runErr = fmt.Errorf("afterRun hook failed: %w", err)
			}
		}
	}()

	if err := fm.hooks.beforeRun.Trigger(ctx, fm); err != nil {
		runErr = fmt.Errorf("beforeRun hook failed: %w", err)
		return ri, runErr
	}

	// An already-canceled context must not run a single cycle.
	if err := fm.contextError(ctx); err != nil {
		runErr = err
		return ri, runErr
	}

	for {
		cycleErr := fm.runCycle(ctx)
		if cycleErr != nil {
			runErr = cycleErr
			return ri, runErr
		}

		if mustStop, err := fm.mustStop(ctx); mustStop {
			runErr = err
			return ri, runErr
		}

		if err := fm.drainComponents(ctx); err != nil {
			runErr = err
			return ri, runErr
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

// mustStop defines when f-mesh must stop (it always checks only the last cycle).
func (fm *FMesh) mustStop(ctx context.Context) (bool, error) {
	lastCycle := fm.runtimeInfo.Cycles.Last()

	// Check if cycles limit is hit
	if (fm.config.CyclesLimit > 0) && (lastCycle.Number() > fm.config.CyclesLimit) {
		fm.LogDebug("going to stop: %s", ErrReachedMaxAllowedCycles)
		return true, ErrReachedMaxAllowedCycles
	}

	// Check if the time constraint is hit
	if fm.config.TimeLimit > 0 {
		if fm.runtimeInfo.Duration() >= fm.config.TimeLimit {
			fm.LogDebug("going to stop: %s", ErrTimeLimitExceeded)
			return true, ErrTimeLimitExceeded
		}
	}

	// Check the context before the error strategy: when a run is canceled,
	// activation functions typically return ctx.Err() and would otherwise be
	// reported as ordinary activation errors, hiding why the mesh stopped.
	if err := fm.contextError(ctx); err != nil {
		fm.LogDebug("going to stop: %s", err)
		return true, err
	}

	// Check if mesh must stop because of the configured error handling strategy.
	// This is evaluated before the natural stop check so activation and hook errors
	// are never silently swallowed when nothing activated in the last cycle.
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if lastCycle.HasActivationErrors() || lastCycle.HasActivationPanics() {
			runError := fmt.Errorf("%w, cycle # %d, activation errors: %w, activation panics: %w",
				ErrHitAnErrorOrPanic, lastCycle.Number(), lastCycle.AllErrorsCombined(), lastCycle.AllPanicsCombined())
			fm.LogDebug("going to stop: %s", runError)
			return true, runError
		}
	case StopOnFirstPanic:
		if lastCycle.HasActivationPanics() {
			runError := fmt.Errorf("%w, cycle # %d, activation panics: %w",
				ErrHitAPanic, lastCycle.Number(), lastCycle.AllPanicsCombined())
			fm.LogDebug("going to stop: %s", runError)
			return true, runError
		}
	case IgnoreAll:
	default:
		fm.LogDebug("going to stop: %s", ErrUnsupportedErrorHandlingStrategy)
		return true, ErrUnsupportedErrorHandlingStrategy
	}

	// Check if the mesh finished naturally (no component activated during the last cycle)
	if !lastCycle.HasActivatedComponents() {
		fm.LogDebug("going to stop naturally")
		return true, nil
	}

	return false, nil
}
