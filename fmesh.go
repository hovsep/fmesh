package fmesh

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
)

// FMesh is the functional mesh.
type FMesh struct {
	name         string
	description  string
	chainableErr error
	components   *component.Collection
	runtimeInfo  *RuntimeInfo
	config       *Config
	hooks        *Hooks
}

// New creates a new F-Mesh with default configuration.
func New(name string) *FMesh {
	return &FMesh{
		name:         name,
		description:  "",
		chainableErr: nil,
		components:   component.NewCollection(),
		runtimeInfo: &RuntimeInfo{
			Cycles:   cycle.NewGroup(),
			Duration: 0,
		},
		config: defaultConfig,
		hooks:  NewHooks(),
	}
}

// Name returns the name of the F-Mesh.
func (fm *FMesh) Name() string {
	return fm.name
}

// Description returns the description of the F-Mesh.
func (fm *FMesh) Description() string {
	return fm.description
}

// NewWithConfig creates a new F-Mesh with custom configuration.
func NewWithConfig(name string, config *Config) *FMesh {
	return New(name).withConfig(config)
}

// Components returns all components in the mesh.
func (fm *FMesh) Components() *component.Collection {
	if fm.HasChainableErr() {
		return component.NewCollection().WithChainableErr(fm.ChainableErr())
	}
	return fm.components
}

// ComponentByName returns a component by name.
func (fm *FMesh) ComponentByName(name string) *component.Component {
	return fm.Components().ByName(name)
}

// WithDescription sets a description.
func (fm *FMesh) WithDescription(description string) *FMesh {
	if fm.HasChainableErr() {
		return fm
	}

	fm.description = description
	return fm
}

// AddComponents adds components to the mesh and returns the mesh for chaining.
func (fm *FMesh) AddComponents(components ...*component.Component) *FMesh {
	if fm.HasChainableErr() {
		return fm
	}

	for _, c := range components {
		// Inherit logger from fm if the component does not have its own
		if c.Logger() == nil {
			c = c.WithLogger(fm.Logger())
		}
		fm.components = fm.components.Add(c.WithParentMesh(fm))
		if c.HasChainableErr() {
			return fm.WithChainableErr(c.ChainableErr())
		}
	}

	fm.LogDebug(fmt.Sprintf("%d components added to mesh", fm.Components().Len()))
	return fm
}

// SetupHooks configures hooks for the mesh using a closure.
// All hook registration happens inside the provided function.
func (fm *FMesh) SetupHooks(configure func(*Hooks)) *FMesh {
	if fm.HasChainableErr() {
		return fm
	}
	configure(fm.hooks)
	return fm
}

// runCycle runs one activation cycle (tries to activate ready components).
func (fm *FMesh) runCycle() {
	newCycle := cycle.New().WithNumber(fm.runtimeInfo.Cycles.Len() + 1)

	if err := fm.hooks.cycleBegin.Trigger(&CycleContext{FMesh: fm, Cycle: newCycle}); err != nil {
		newCycle.WithChainableErr(errors.Join(errFailedToRunCycle, fmt.Errorf("cycleBegin hook failed: %w", err)))
		fm.WithChainableErr(newCycle.ChainableErr())
		fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(newCycle)
		return
	}

	fm.LogDebug(fmt.Sprintf("starting activation cycle #%d", newCycle.Number()))

	if fm.HasChainableErr() {
		newCycle.WithChainableErr(fm.ChainableErr())
	}

	if fm.Components().Len() == 0 {
		newCycle.WithChainableErr(errors.Join(errFailedToRunCycle, errNoComponents))
	}

	var wg sync.WaitGroup

	components, err := fm.Components().All()
	if err != nil {
		newCycle.WithChainableErr(errors.Join(errFailedToRunCycle, err))
	}

	for _, c := range components {
		if c.HasChainableErr() {
			fm.WithChainableErr(c.ChainableErr())
		}
		wg.Add(1)

		go func(component *component.Component, cycle *cycle.Cycle) {
			defer wg.Done()

			cycle.ActivationResults().Add(component.MaybeActivate())
		}(c, newCycle)
	}

	wg.Wait()

	// Bubble up chain errors from activation results
	activationResults, err := newCycle.ActivationResults().All()
	if err != nil {
		newCycle.WithChainableErr(err)
	} else {
		for _, ar := range activationResults {
			if ar.HasChainableErr() {
				newCycle.WithChainableErr(ar.ChainableErr())
				break
			}
		}
	}

	if newCycle.HasChainableErr() {
		fm.WithChainableErr(newCycle.ChainableErr())
	}

	if fm.IsDebug() {
		newCycle.ActivationResults().ForEach(func(ar *component.ActivationResult) {
			fm.LogDebug(fmt.Sprintf("activation result for component %s : activated: %t, , code: %s, is error: %t, is panic: %t, error: %v", ar.ComponentName(), ar.Activated(), ar.Code(), ar.IsError(), ar.IsPanic(), ar.ActivationError()))
		})
	}

	if err := fm.hooks.cycleEnd.Trigger(&CycleContext{FMesh: fm, Cycle: newCycle}); err != nil {
		newCycle.WithChainableErr(errors.Join(errFailedToRunCycle, fmt.Errorf("cycleEnd hook failed: %w", err)))
		fm.WithChainableErr(newCycle.ChainableErr())
	}

	fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(newCycle)
}

// DrainComponents drains the data from activated components.
func (fm *FMesh) drainComponents() {
	if fm.HasChainableErr() {
		fm.WithChainableErr(errors.Join(ErrFailedToDrain, fm.ChainableErr()))
		return
	}

	fm.clearInputs()
	if fm.HasChainableErr() {
		return
	}

	components, err := fm.Components().All()
	if err != nil {
		fm.WithChainableErr(errors.Join(ErrFailedToDrain, err))
		return
	}

	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByName(c.Name())

		if activationResult.HasChainableErr() {
			fm.WithChainableErr(errors.Join(ErrFailedToDrain, activationResult.ChainableErr()))
			return
		}

		if !activationResult.Activated() {
			// Component did not activate, so it did not create new output signals, hence nothing to drain
			continue
		}

		// Components waiting for inputs are never drained
		if component.IsWaitingForInput(activationResult) {
			// @TODO: maybe we should additionally clear outputs
			// because it is technically possible to set some output signals and then return errWaitingForInput in AF
			continue
		}

		c.FlushOutputs()
	}
}

// clearInputs clears all the input ports of all components activated in the latest cycle.
func (fm *FMesh) clearInputs() {
	if fm.HasChainableErr() {
		return
	}

	components, err := fm.Components().All()
	if err != nil {
		fm.WithChainableErr(errors.Join(errFailedToClearInputs, err))
		return
	}

	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByName(c.Name())

		if activationResult.HasChainableErr() {
			fm.WithChainableErr(errors.Join(errFailedToClearInputs, activationResult.ChainableErr()))
		}

		if !activationResult.Activated() {
			// Component did not activate hence it's inputs must be clear
			continue
		}

		if component.IsWaitingForInput(activationResult) && component.WantsToKeepInputs(activationResult) {
			// Component want to keep inputs for the next cycle
			// @TODO: add fine grained control on which ports to keep
			continue
		}

		c.ClearInputs()
	}
}

// Run executes the mesh, activating components until completion or cycle limit.
func (fm *FMesh) Run() (*RuntimeInfo, error) {
	fm.runtimeInfo.StartedAt = time.Now()
	defer func() {
		fm.runtimeInfo.StoppedAt = time.Now()
		fm.runtimeInfo.Duration = time.Since(fm.runtimeInfo.StartedAt)
		if err := fm.hooks.afterRun.Trigger(fm); err != nil {
			// Don't overwrite existing chainable error
			if !fm.HasChainableErr() {
				fm.WithChainableErr(fmt.Errorf("afterRun hook failed: %w", err))
			}
		}
	}()

	if err := fm.hooks.beforeRun.Trigger(fm); err != nil {
		return fm.runtimeInfo, fmt.Errorf("beforeRun hook failed: %w", err)
	}

	if fm.HasChainableErr() {
		return fm.runtimeInfo, fm.ChainableErr()
	}

	validationErr := fm.validate()

	if validationErr != nil {
		return fm.WithChainableErr(validationErr).runtimeInfo, validationErr
	}

	for {
		fm.runCycle()

		if mustStop, err := fm.mustStop(); mustStop {
			return fm.runtimeInfo, err
		}

		fm.drainComponents()
		if fm.HasChainableErr() {
			return fm.runtimeInfo, fm.ChainableErr()
		}
	}
}

// mustStop defines when f-mesh must stop (it always checks only the last cycle).
func (fm *FMesh) mustStop() (bool, error) {
	if fm.HasChainableErr() {
		return false, nil
	}

	lastCycle := fm.runtimeInfo.Cycles.Last()

	// Check if cycles limit is hit
	if (fm.config.CyclesLimit > 0) && (lastCycle.Number() > fm.config.CyclesLimit) {
		fm.LogDebug(fmt.Sprintf("going to stop: %s", ErrReachedMaxAllowedCycles))

		return true, ErrReachedMaxAllowedCycles
	}

	// Check if the time constraint is hit
	if fm.config.TimeLimit != UnlimitedTime {
		if time.Since(fm.runtimeInfo.StartedAt) >= fm.config.TimeLimit {
			fm.LogDebug(fmt.Sprintf("going to stop: %s", ErrTimeLimitExceeded))

			return true, ErrTimeLimitExceeded
		}
	}

	// Check if the mesh finished naturally (no component activated during the last cycle)
	if !lastCycle.HasActivatedComponents() {
		// Stop naturally (no components activated during the cycle => all inputs are processed)
		fm.LogDebug("going to stop naturally")

		return true, nil
	}

	// Check if mesh must stop because of the configured error handling strategy
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if lastCycle.HasActivationErrors() || lastCycle.HasActivationPanics() {
			runError := fmt.Errorf("%w, cycle # %d, activation errors: %w, activation panics: %w", ErrHitAnErrorOrPanic, lastCycle.Number(), lastCycle.AllErrorsCombined(), lastCycle.AllPanicsCombined())
			fm.LogDebug(fmt.Sprintf("going to stop: %s", runError))

			return true, runError
		}
		return false, nil
	case StopOnFirstPanic:
		if lastCycle.HasActivationPanics() {
			runError := fmt.Errorf("%w, cycle # %d, activation panics: %w", ErrHitAPanic, lastCycle.Number(), lastCycle.AllPanicsCombined())
			fm.LogDebug(fmt.Sprintf("going to stop: %s", runError))

			return true, runError
		}
		return false, nil
	case IgnoreAll:
		return false, nil
	default:
		fm.LogDebug(fmt.Sprintf("going to stop: %s", ErrUnsupportedErrorHandlingStrategy))

		return true, ErrUnsupportedErrorHandlingStrategy
	}
}

// WithChainableErr returns f-mesh with an error.
// The error is automatically joined with the mesh's name as context.
func (fm *FMesh) WithChainableErr(err error) *FMesh {
	if err == nil {
		fm.chainableErr = nil
		return fm
	}

	contextErr := fmt.Errorf("error in fmesh '%s'", fm.name)
	fm.chainableErr = errors.Join(contextErr, err)
	return fm
}

// HasChainableErr returns true when a chainable error is set.
func (fm *FMesh) HasChainableErr() bool {
	return fm.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (fm *FMesh) ChainableErr() error {
	return fm.chainableErr
}

func (fm *FMesh) validate() error {
	if fm.HasChainableErr() {
		return fmt.Errorf("failed to validate fmesh: %w", fm.ChainableErr())
	}

	components, err := fm.Components().All()
	if err != nil {
		return fmt.Errorf("failed to get components: %w", err)
	}

	for _, c := range components {
		if c.HasChainableErr() {
			return fmt.Errorf("failed to validate component %s: %w", c.Name(), c.ChainableErr())
		}

		if c.ParentMesh() == nil {
			return fmt.Errorf("component %s his not registered in the mesh", c.Name())
		}

		if c.ParentMesh() != fm {
			return fmt.Errorf("component %s has invalid parent mesh", c.Name())
		}

		outputs, err := c.Outputs().All()
		if err != nil {
			return fmt.Errorf("failed to get outputs for component %s: %w", c.Name(), err)
		}

		for _, p := range outputs {
			if p.HasChainableErr() {
				return fmt.Errorf("failed to validate port %s in component %s: %w", p.Name(), c.Name(), p.ChainableErr())
			}

			if p.ParentComponent() == nil {
				return fmt.Errorf("port %s in component %s has not parent component set", p.Name(), c.Name())
			}

			if p.ParentComponent() != c {
				return fmt.Errorf("port %s in component %s has invalid parent component", p.Name(), c.Name())
			}

			pipes, err := p.Pipes().All()
			if err != nil {
				return fmt.Errorf("failed to get pipes for port %s in component %s: %w", p.Name(), c.Name(), err)
			}

			for _, pipe := range pipes {
				if pipe.ParentComponent() == nil {
					return fmt.Errorf("pipe leads to unregistered port %s in component %s", pipe.Name(), c.Name())
				}

				destComponent := fm.components.ByName(pipe.ParentComponent().Name())
				if destComponent.HasChainableErr() {
					return fmt.Errorf("pipe leads to absent component %s: %w", pipe.ParentComponent().Name(), destComponent.ChainableErr())
				}

				if destComponent.ParentMesh() == nil {
					return fmt.Errorf("pipe leads to unregistered component %s", pipe.ParentComponent().Name())
				}

				if destComponent.ParentMesh() != fm {
					return fmt.Errorf("pipe leads to port %s in component %s that has invalid parent mesh", pipe.Name(), c.Name())
				}
			}
		}
	}

	return nil
}
