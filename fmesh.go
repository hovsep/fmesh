package fmesh

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"sync"
)

// FMesh is the functional mesh
type FMesh struct {
	common.NamedEntity
	common.DescribedEntity
	*common.Chainable
	components *component.Collection
	cycles     *cycle.Group
	config     *Config
}

// New creates a new f-mesh with default config
func New(name string) *FMesh {
	return &FMesh{
		NamedEntity:     common.NewNamedEntity(name),
		DescribedEntity: common.NewDescribedEntity(""),
		Chainable:       common.NewChainable(),
		components:      component.NewCollection(),
		cycles:          cycle.NewGroup(),
		config:          defaultConfig,
	}
}

// NewWithConfig creates a new f-mesh with custom config
func NewWithConfig(name string, config *Config) *FMesh {
	return New(name).withConfig(config)
}

// Components getter
func (fm *FMesh) Components() *component.Collection {
	if fm.HasErr() {
		return component.NewCollection().WithErr(fm.Err())
	}
	return fm.components
}

// ComponentByName shortcut method
func (fm *FMesh) ComponentByName(name string) *component.Component {
	return fm.Components().ByName(name)
}

// WithDescription sets a description
func (fm *FMesh) WithDescription(description string) *FMesh {
	if fm.HasErr() {
		return fm
	}

	fm.DescribedEntity = common.NewDescribedEntity(description)
	return fm
}

// WithComponents adds components to f-mesh
func (fm *FMesh) WithComponents(components ...*component.Component) *FMesh {
	if fm.HasErr() {
		return fm
	}

	for _, c := range components {
		fm.components = fm.components.With(c.WithLogger(fm.Logger()))
		if c.HasErr() {
			return fm.WithErr(c.Err())
		}
	}

	fm.LogDebug(fmt.Sprintf("%d components added to mesh", fm.Components().Len()))
	return fm
}

// runCycle runs one activation cycle (tries to activate ready components)
func (fm *FMesh) runCycle() {
	newCycle := cycle.New().WithNumber(fm.cycles.Len() + 1)

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

			cycle.Lock()
			cycle.ActivationResults().Add(c.MaybeActivate())
			cycle.Unlock()
		}(c, newCycle)
	}

	wg.Wait()

	//Bubble up chain errors from activation results
	for _, ar := range newCycle.ActivationResults() {
		if ar.HasErr() {
			newCycle.SetErr(ar.Err())
			break
		}
	}

	if newCycle.HasErr() {
		fm.SetErr(newCycle.Err())
	}

	if fm.IsDebug() {
		for _, ar := range newCycle.ActivationResults() {
			fm.LogDebug(fmt.Sprintf("activation result for component %s : activated: %t, , code: %s, is error: %t, is panic: %t, error: %v", ar.ComponentName(), ar.Activated(), ar.Code(), ar.IsError(), ar.IsPanic(), ar.ActivationError()))
		}
	}

	fm.cycles = fm.cycles.With(newCycle)
}

// DrainComponents drains the data from activated components
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

	lastCycle := fm.cycles.Last()

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

// clearInputs clears all the input ports of all components activated in latest cycle
func (fm *FMesh) clearInputs() {
	if fm.HasErr() {
		return
	}

	components, err := fm.Components().Components()
	if err != nil {
		fm.SetErr(errors.Join(errFailedToClearInputs, err))
		return
	}

	lastCycle := fm.cycles.Last()

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
			//@TODO: add fine grained control on which ports to keep
			continue
		}

		c.ClearInputs()
	}
}

// Run starts the computation until there is no component which activates (mesh has no unprocessed inputs)
func (fm *FMesh) Run() (cycle.Cycles, error) {
	if fm.HasErr() {
		return nil, fm.Err()
	}

	for {
		fm.runCycle()

		if mustStop, err := fm.mustStop(); mustStop {
			return fm.cycles.CyclesOrNil(), err
		}

		fm.drainComponents()
		if fm.HasErr() {
			return nil, fm.Err()
		}
	}
}

// mustStop defines when f-mesh must stop (it always checks only last cycle)
func (fm *FMesh) mustStop() (bool, error) {
	if fm.HasErr() {
		return false, nil
	}

	lastCycle := fm.cycles.Last()

	if (fm.config.CyclesLimit > 0) && (lastCycle.Number() > fm.config.CyclesLimit) {
		return true, ErrReachedMaxAllowedCycles
	}

	if !lastCycle.HasActivatedComponents() {
		// Stop naturally (no components activated during the cycle => all inputs are processed)
		return true, nil
	}

	//Check if mesh must stop because of configured error handling strategy
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if lastCycle.HasErrors() || lastCycle.HasPanics() {
			//@TODO: add failing components names to error
			return true, fmt.Errorf("%w, cycle # %d, activation errors: %w", ErrHitAnErrorOrPanic, lastCycle.Number(), lastCycle.AllErrorsCombined())
		}
		return false, nil
	case StopOnFirstPanic:
		// @TODO: add more context to error
		if lastCycle.HasPanics() {
			return true, ErrHitAPanic
		}
		return false, nil
	case IgnoreAll:
		return false, nil
	default:
		return true, ErrUnsupportedErrorHandlingStrategy
	}
}

// WithErr returns f-mesh with error
func (fm *FMesh) WithErr(err error) *FMesh {
	fm.SetErr(err)
	return fm
}
