package fmesh

import (
	"errors"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"sync"
)

// @TODO: move this to fm.Config
const maxCyclesAllowed = 100 //Dev mode

// FMesh is the functional mesh
type FMesh struct {
	name                  string
	description           string
	components            component.Collection
	errorHandlingStrategy ErrorHandlingStrategy
}

// New creates a new f-mesh
func New(name string) *FMesh {
	return &FMesh{name: name, components: component.NewComponentCollection()}
}

// Name getter
func (fm *FMesh) Name() string {
	return fm.name
}

// Description getter
func (fm *FMesh) Description() string {
	return fm.description
}

func (fm *FMesh) Components() component.Collection {
	return fm.components
}

// WithDescription sets a description
func (fm *FMesh) WithDescription(description string) *FMesh {
	fm.description = description
	return fm
}

// WithComponents adds components to f-mesh
func (fm *FMesh) WithComponents(components ...*component.Component) *FMesh {
	for _, c := range components {
		fm.components.Add(c)
	}
	return fm
}

// WithErrorHandlingStrategy defines how the mesh will handle errors
func (fm *FMesh) WithErrorHandlingStrategy(strategy ErrorHandlingStrategy) *FMesh {
	fm.errorHandlingStrategy = strategy
	return fm
}

// runCycle runs one activation cycle (tries to activate ready components)
func (fm *FMesh) runCycle() *cycle.Cycle {
	newCycle := cycle.New()

	if len(fm.components) == 0 {
		return newCycle
	}

	var wg sync.WaitGroup

	for _, c := range fm.components {
		wg.Add(1)

		go func(component *component.Component, cycle *cycle.Cycle) {
			defer wg.Done()

			cycle.Lock()
			cycle.ActivationResults().Add(c.MaybeActivate())
			cycle.Unlock()
		}(c, newCycle)
	}

	wg.Wait()
	return newCycle
}

// DrainComponents drains the data from activated components
func (fm *FMesh) drainComponentsAfterCycle(cycle *cycle.Cycle) {
	for _, c := range fm.components {
		activationResult := cycle.ActivationResults().ByComponentName(c.Name())

		if !activationResult.Activated() {
			continue
		}

		c.FlushOutputs(activationResult)
		c.FlushInputs(activationResult, c.WantsToKeepInputs(activationResult)) // Inputs are a bit trickier
	}
}

// Run starts the computation until there is no component which activates (mesh has no unprocessed inputs)
func (fm *FMesh) Run() (cycle.Collection, error) {
	allCycles := cycle.NewCollection()
	for {
		cycleResult := fm.runCycle()
		allCycles = allCycles.With(cycleResult)

		mustStop, err := fm.mustStop(cycleResult, len(allCycles))
		if mustStop {
			return allCycles, err
		}

		fm.drainComponentsAfterCycle(cycleResult)
	}
}

func (fm *FMesh) mustStop(cycleResult *cycle.Cycle, cycleNum int) (bool, error) {
	if cycleNum >= maxCyclesAllowed {
		return true, errors.New("reached max allowed cycles")
	}

	//Check if we are done (no components activated during the cycle => all inputs are processed)
	if !cycleResult.HasActivatedComponents() {
		return true, nil
	}

	//Check if mesh must stop because of configured error handling strategy
	switch fm.errorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		return cycleResult.HasErrors() || cycleResult.HasPanics(), ErrHitAnErrorOrPanic
	case StopOnFirstPanic:
		return cycleResult.HasPanics(), ErrHitAPanic
	case IgnoreAll:
		return false, nil
	default:
		return true, ErrUnsupportedErrorHandlingStrategy
	}
}
