package fmesh

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
)

// RuntimeInfo contains information about the runtime of the f-mesh.
type RuntimeInfo struct {
	Cycles    *cycle.Group
	StartedAt time.Time
	StoppedAt time.Time
	Duration  time.Duration
}

// FMesh is the functional mesh.
type FMesh struct {
	name        string
	description string
	*common.Chainable
	components  *component.Collection
	runtimeInfo *RuntimeInfo
	config      *Config
}

// New creates a new f-mesh with default config.
func New(name string) *FMesh {
	return &FMesh{
		name:        name,
		description: "",
		Chainable:   common.NewChainable(),
		components:  component.NewCollection(),
		runtimeInfo: &RuntimeInfo{
			Cycles:   cycle.NewGroup(),
			Duration: 0,
		},
		config: defaultConfig,
	}
}

// Name getter.
func (fm *FMesh) Name() string {
	return fm.name
}

// Description getter.
func (fm *FMesh) Description() string {
	return fm.description
}

// NewWithConfig creates a new f-mesh with custom config
func NewWithConfig(name string, config *Config) *FMesh {
	return New(name).withConfig(config)
}

// Components getter.
func (fm *FMesh) Components() *component.Collection {
	if fm.HasErr() {
		return component.NewCollection().WithErr(fm.Err())
	}
	return fm.components
}

// ComponentByName shortcut method.
func (fm *FMesh) ComponentByName(name string) *component.Component {
	return fm.Components().ByName(name)
}

// WithDescription sets a description.
func (fm *FMesh) WithDescription(description string) *FMesh {
	if fm.HasErr() {
		return fm
	}

	fm.description = description
	return fm
}

// WithComponents adds components to f-mesh.
func (fm *FMesh) WithComponents(components ...*component.Component) *FMesh {
	if fm.HasErr() {
		return fm
	}

	for _, c := range components {
		if c.Logger() == nil {
			// Inherit logger from fm if component does not have its own
			c = c.WithLogger(fm.Logger())
		}
		fm.components = fm.components.With(c)
		if c.HasErr() {
			return fm.WithErr(c.Err())
		}
	}

	fm.LogDebug(fmt.Sprintf("%d components added to mesh", fm.Components().Len()))
	return fm
}

// runCycle runs one activation cycle (tries to activate ready components).
func (fm *FMesh) runCycle() {
	newCycle := cycle.New().WithNumber(fm.runtimeInfo.Cycles.Len() + 1)

	fm.LogDebug(fmt.Sprintf("starting activation cycle #%d", newCycle.Number()))

	if fm.HasErr() {
		newCycle.SetErr(fm.Err())
	}

	if fm.Components().Len() == 0 {
		newCycle.SetErr(errors.Join(errFailedToRunCycle, errNoComponents))
	}

	var wg sync.WaitGroup

	components, err := fm.Components().Components()
	if err != nil {
		newCycle.SetErr(errors.Join(errFailedToRunCycle, err))
	}

	for _, c := range components {
		if c.HasErr() {
			fm.SetErr(c.Err())
		}
		wg.Add(1)

		go func(component *component.Component, cycle *cycle.Cycle) {
			defer wg.Done()

			cycle.ActivationResults().Add(component.MaybeActivate())
		}(c, newCycle)
	}

	wg.Wait()

	// Bubble up chain errors from activation results
	for _, ar := range newCycle.ActivationResults().All() {
		if ar.HasErr() {
			newCycle.SetErr(ar.Err())
			break
		}
	}

	if newCycle.HasErr() {
		fm.SetErr(newCycle.Err())
	}

	if fm.IsDebug() {
		for _, ar := range newCycle.ActivationResults().All() {
			fm.LogDebug(fmt.Sprintf("activation result for component %s : activated: %t, , code: %s, is error: %t, is panic: %t, error: %v", ar.ComponentName(), ar.Activated(), ar.Code(), ar.IsError(), ar.IsPanic(), ar.ActivationError()))
		}
	}

	fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.With(newCycle)
}

// DrainComponents drains the data from activated components.
func (fm *FMesh) drainComponents() {
	if fm.HasErr() {
		fm.SetErr(errors.Join(ErrFailedToDrain, fm.Err()))
		return
	}

	fm.clearInputs()
	if fm.HasErr() {
		return
	}

	components, err := fm.Components().Components()
	if err != nil {
		fm.SetErr(errors.Join(ErrFailedToDrain, err))
		return
	}

	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByComponentName(c.Name())

		if activationResult.HasErr() {
			fm.SetErr(errors.Join(ErrFailedToDrain, activationResult.Err()))
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
	if fm.HasErr() {
		return
	}

	components, err := fm.Components().Components()
	if err != nil {
		fm.SetErr(errors.Join(errFailedToClearInputs, err))
		return
	}

	lastCycle := fm.runtimeInfo.Cycles.Last()

	for _, c := range components {
		activationResult := lastCycle.ActivationResults().ByComponentName(c.Name())

		if activationResult.HasErr() {
			fm.SetErr(errors.Join(errFailedToClearInputs, activationResult.Err()))
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

// Run starts the computation until there is no component that activates (mesh has no unprocessed inputs).
func (fm *FMesh) Run() (*RuntimeInfo, error) {
	fm.runtimeInfo.StartedAt = time.Now()
	defer func() {
		fm.runtimeInfo.StoppedAt = time.Now()
		fm.runtimeInfo.Duration = time.Since(fm.runtimeInfo.StartedAt)
	}()

	if fm.HasErr() {
		return fm.runtimeInfo, fm.Err()
	}

	for {
		fm.runCycle()

		if mustStop, err := fm.mustStop(); mustStop {
			return fm.runtimeInfo, err
		}

		fm.drainComponents()
		if fm.HasErr() {
			return fm.runtimeInfo, fm.Err()
		}
	}
}

// mustStop defines when f-mesh must stop (it always checks only the last cycle).
func (fm *FMesh) mustStop() (bool, error) {
	if fm.HasErr() {
		return false, nil
	}

	lastCycle := fm.runtimeInfo.Cycles.Last()

	// Check if cycles limit is hit
	if (fm.config.CyclesLimit > 0) && (lastCycle.Number() > fm.config.CyclesLimit) {
		fm.LogDebug(fmt.Sprintf("going to stop: %s", ErrReachedMaxAllowedCycles))

		return true, ErrReachedMaxAllowedCycles
	}

	// Check if time constraint is hit
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

	// Check if mesh must stop because of configured error handling strategy
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if lastCycle.HasErrors() || lastCycle.HasPanics() {
			runError := fmt.Errorf("%w, cycle # %d, activation errors: %w, activation panics: %w", ErrHitAnErrorOrPanic, lastCycle.Number(), lastCycle.AllErrorsCombined(), lastCycle.AllPanicsCombined())
			fm.LogDebug(fmt.Sprintf("going to stop: %s", runError))

			return true, runError
		}
		return false, nil
	case StopOnFirstPanic:
		if lastCycle.HasPanics() {
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

// WithErr returns f-mesh with an error.
func (fm *FMesh) WithErr(err error) *FMesh {
	fm.SetErr(err)
	return fm
}
