package fmesh

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
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
}

// New creates a new F-Mesh with the default configuration and applies any provided options.
func New(name string, opts ...Option) (*FMesh, error) {
	fm := &FMesh{
		name:        name,
		description: "",
		labels:      meta.NewLabels(),
		scalars:     meta.NewScalars(),
		components:  component.NewCollection(),
		runtimeInfo: newRuntimeInfo(),
		logger:      newDefaultLogger(name),
		config:      newDefaultConfig(),
		hooks:       newHooks(),
	}
	for _, opt := range opts {
		if err := opt(fm); err != nil {
			return nil, fmt.Errorf("fmesh %q option failed: %w", name, err)
		}
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

		if err := fm.components.Add(c); err != nil {
			return fmt.Errorf("failed to add component %q to mesh: %w", c.Name(), err)
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
func (fm *FMesh) runCycle() error {
	newCycle := cycle.New().SetNumber(fm.runtimeInfo.Cycles.Len() + 1)

	if err := fm.hooks.cycleBegin.Trigger(&CycleContext{FMesh: fm, Cycle: newCycle}); err != nil {
		cycleErr := errors.Join(errFailedToRunCycle, fmt.Errorf("cycleBegin hook failed: %w", err))
		fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(newCycle)
		return cycleErr
	}

	fm.LogDebug("starting activation cycle #%d", newCycle.Number())

	components := fm.Components().All()
	if len(components) == 0 {
		fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(newCycle)
		return errors.Join(errFailedToRunCycle, errNoComponents)
	}

	var wg sync.WaitGroup

	for _, c := range components {
		wg.Add(1)
		go func(comp *component.Component, cyc *cycle.Cycle) {
			defer wg.Done()
			cyc.ActivationResults().Add(comp.MaybeActivate())
		}(c, newCycle)
	}

	wg.Wait()

	if fm.IsDebug() {
		_ = newCycle.ActivationResults().ForEach(func(ar *component.ActivationResult) error {
			fm.LogDebug("activation result for component %s: activated: %t, code: %s, is error: %t, is panic: %t, error: %v",
				ar.ComponentName(), ar.Activated(), ar.Code(), ar.IsError(), ar.IsPanic(), ar.ActivationError())
			return nil
		})
	}

	if err := fm.hooks.cycleEnd.Trigger(&CycleContext{FMesh: fm, Cycle: newCycle}); err != nil {
		cycleErr := errors.Join(errFailedToRunCycle, fmt.Errorf("cycleEnd hook failed: %w", err))
		fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(newCycle)
		return cycleErr
	}

	fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(newCycle)
	return nil
}

// drainComponents drains the data from activated components.
func (fm *FMesh) drainComponents() error {
	if err := fm.clearInputs(); err != nil {
		return errors.Join(ErrFailedToDrain, err)
	}

	components := fm.Components().All()

	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByName(c.Name())

		if !activationResult.Activated() {
			// Component did not activate, so it did not create new output signals, hence nothing to drain
			continue
		}

		// Components waiting for inputs are never drained
		if component.IsWaitingForInput(activationResult) {
			continue
		}

		if err := c.FlushOutputs(); err != nil {
			return errors.Join(ErrFailedToDrain, fmt.Errorf("failed to flush outputs of component %q: %w", c.Name(), err))
		}
	}
	return nil
}

// @TODO: we can inline this into drainComponents
// clearInputs clears all the input ports of all components activated in the latest cycle.
func (fm *FMesh) clearInputs() error {
	components := fm.Components().All()

	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByName(c.Name())

		if !activationResult.Activated() {
			// Component did not activate hence its inputs must be clear
			continue
		}

		if component.IsWaitingForInput(activationResult) && component.WantsToKeepInputs(activationResult) {
			// Component wants to keep inputs for the next cycle
			continue
		}

		if err := c.ClearInputs(); err != nil {
			return errors.Join(errFailedToClearInputs, fmt.Errorf("component %q: %w", c.Name(), err))
		}
	}
	return nil
}

func (fm *FMesh) cleanUpPreviousRun() error {
	// Clear all output ports to prevent signal accumulation between runs
	if err := fm.Components().ForEach(func(c *component.Component) error {
		return c.ClearOutputs()
	}); err != nil {
		return err
	}

	// Init runtime info
	fm.runtimeInfo = newRuntimeInfo()
	fm.runtimeInfo.MarkStarted()
	return nil
}

// Run executes the mesh, activating components until completion or cycle limit.
func (fm *FMesh) Run() (ri *RuntimeInfo, runErr error) {
	if err := fm.cleanUpPreviousRun(); err != nil {
		return nil, err
	}

	ri = fm.runtimeInfo

	defer func() {
		fm.runtimeInfo.MarkStopped()
		if err := fm.hooks.afterRun.Trigger(fm); err != nil {
			if runErr == nil {
				runErr = fmt.Errorf("afterRun hook failed: %w", err)
			}
		}
	}()

	if err := fm.hooks.beforeRun.Trigger(fm); err != nil {
		runErr = fmt.Errorf("beforeRun hook failed: %w", err)
		return ri, runErr
	}

	if err := fm.validateBeforeRun(); err != nil {
		runErr = err
		return ri, runErr
	}

	for {
		cycleErr := fm.runCycle()
		if cycleErr != nil {
			runErr = cycleErr
			return ri, runErr
		}

		if mustStop, err := fm.mustStop(); mustStop {
			runErr = err
			return ri, runErr
		}

		if err := fm.drainComponents(); err != nil {
			runErr = err
			return ri, runErr
		}
	}
}

// mustStop defines when f-mesh must stop (it always checks only the last cycle).
func (fm *FMesh) mustStop() (bool, error) {
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

	// Check if the mesh finished naturally (no component activated during the last cycle)
	if !lastCycle.HasActivatedComponents() {
		fm.LogDebug("going to stop naturally")
		return true, nil
	}

	// Check if mesh must stop because of the configured error handling strategy
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if lastCycle.HasActivationErrors() || lastCycle.HasActivationPanics() {
			runError := fmt.Errorf("%w, cycle # %d, activation errors: %w, activation panics: %w",
				ErrHitAnErrorOrPanic, lastCycle.Number(), lastCycle.AllErrorsCombined(), lastCycle.AllPanicsCombined())
			fm.LogDebug("going to stop: %s", runError)
			return true, runError
		}
		return false, nil
	case StopOnFirstPanic:
		if lastCycle.HasActivationPanics() {
			runError := fmt.Errorf("%w, cycle # %d, activation panics: %w",
				ErrHitAPanic, lastCycle.Number(), lastCycle.AllPanicsCombined())
			fm.LogDebug("going to stop: %s", runError)
			return true, runError
		}
		return false, nil
	case IgnoreAll:
		return false, nil
	default:
		fm.LogDebug("going to stop: %s", ErrUnsupportedErrorHandlingStrategy)
		return true, ErrUnsupportedErrorHandlingStrategy
	}
}

// validateBeforeRun does pre-run checks using plain loops (no nested ForEach chains).
func (fm *FMesh) validateBeforeRun() error {
	components := fm.Components().All()

	for _, c := range components {
		if err := c.ValidateBeforeActivating(); err != nil {
			return fmt.Errorf("invalid component %q: %w", c.Name(), err)
		}

		if c.ParentMesh() != fm {
			return fmt.Errorf("component %q has invalid parent mesh", c.Name())
		}

		outputPorts := c.Outputs().All()

		for _, p := range outputPorts {
			if err := p.ValidateBeforeActivation(); err != nil {
				return fmt.Errorf("invalid port %q in component %q: %w", p.Name(), c.Name(), err)
			}

			if p.ParentComponent() != c {
				return fmt.Errorf("port %q in component %q has invalid parent component", p.Name(), c.Name())
			}

			destPorts := p.Pipes().All()
			for _, destPort := range destPorts {
				if err := destPort.ValidateBeforeActivation(); err != nil {
					return fmt.Errorf("invalid pipe destination port %q from port %q: %w", destPort.Name(), p.Name(), err)
				}

				parent := destPort.ParentComponent()
				destComponent, ok := parent.(*component.Component)
				if !ok || destComponent == nil {
					return fmt.Errorf("destination port %q has invalid parent component", destPort.Name())
				}

				if err := destComponent.ValidateBeforeActivating(); err != nil {
					return fmt.Errorf("invalid component %q (destination): %w", destComponent.Name(), err)
				}

				if destComponent.ParentMesh() != fm {
					return fmt.Errorf("component %q has invalid parent mesh", destComponent.Name())
				}
			}
		}
	}
	return nil
}
