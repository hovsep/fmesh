package fmesh

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"sync"
)

const UnlimitedCycles = 0

type Config struct {
	// ErrorHandlingStrategy defines how f-mesh will handle errors and panics
	ErrorHandlingStrategy ErrorHandlingStrategy
	// CyclesLimit defines max number of activation cycles, 0 means no limit
	CyclesLimit int
}

var defaultConfig = Config{
	ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
	CyclesLimit:           1000,
}

// FMesh is the functional mesh
type FMesh struct {
	common.NamedEntity
	common.DescribedEntity
	*common.Chainable
	components *component.Collection
	config     Config
}

// New creates a new f-mesh
func New(name string) *FMesh {
	return &FMesh{
		NamedEntity:     common.NewNamedEntity(name),
		DescribedEntity: common.NewDescribedEntity(""),
		Chainable:       common.NewChainable(),
		components:      component.NewCollection(),
		config:          defaultConfig,
	}
}

// Components getter
func (fm *FMesh) Components() *component.Collection {
	if fm.HasErr() {
		return component.NewCollection().WithErr(fm.Err())
	}
	return fm.components
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
		fm.components = fm.components.With(c)
		if c.HasErr() {
			return fm.WithErr(c.Err())
		}
	}
	return fm
}

// WithConfig sets the configuration and returns the f-mesh
func (fm *FMesh) WithConfig(config Config) *FMesh {
	if fm.HasErr() {
		return fm
	}

	fm.config = config
	return fm
}

// runCycle runs one activation cycle (tries to activate ready components)
func (fm *FMesh) runCycle() *cycle.Cycle {
	newCycle := cycle.New()

	if fm.HasErr() {
		return newCycle.WithErr(fm.Err())
	}

	if fm.Components().Len() == 0 {
		fm.SetErr(errors.New("failed to run cycle: no components found"))
		return newCycle.WithErr(fm.Err())
	}

	var wg sync.WaitGroup

	components, err := fm.Components().Components()
	if err != nil {
		fm.SetErr(fmt.Errorf("failed to run cycle: %w", err))
		return newCycle.WithErr(fm.Err())
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

	return newCycle
}

// DrainComponents drains the data from activated components
func (fm *FMesh) drainComponents(cycle *cycle.Cycle) error {
	if fm.HasErr() {
		return fm.Err()
	}

	components, err := fm.Components().Components()
	if err != nil {
		return fmt.Errorf("failed to drain components: %w", err)
	}

	for _, c := range components {
		activationResult := cycle.ActivationResults().ByComponentName(c.Name())

		if activationResult.HasErr() {
			return activationResult.Err()
		}

		if !activationResult.Activated() {
			// Component did not activate, so it did not create new output signals, hence nothing to drain
			continue
		}

		// By default, all outputs are flushed and all inputs are cleared
		shouldFlushOutputs := true
		shouldClearInputs := true

		if component.IsWaitingForInput(activationResult) {
			// @TODO: maybe we should clear outputs
			// in order to prevent leaking outputs from previous cycle
			// (if outputs were set before returning errWaitingForInputs)
			shouldFlushOutputs = false
			shouldClearInputs = !component.WantsToKeepInputs(activationResult)
		}

		if shouldClearInputs {
			c.ClearInputs()
		}

		if shouldFlushOutputs {
			c.FlushOutputs()
		}

	}
	return nil
}

// Run starts the computation until there is no component which activates (mesh has no unprocessed inputs)
func (fm *FMesh) Run() (cycle.Cycles, error) {
	if fm.HasErr() {
		return nil, fm.Err()
	}

	allCycles := cycle.NewGroup()
	cycleNumber := 0
	for {
		cycleResult := fm.runCycle().WithNumber(cycleNumber)

		if cycleResult.HasErr() {
			fm.SetErr(cycleResult.Err())
			return nil, fmt.Errorf("chain error occurred in cycle #%d : %w", cycleResult.Number(), cycleResult.Err())
		}

		allCycles = allCycles.With(cycleResult)

		mustStop, chainError, stopError := fm.mustStop(cycleResult)
		if chainError != nil {
			return nil, chainError
		}

		if mustStop {
			cycles, err := allCycles.Cycles()
			if err != nil {
				return nil, err
			}
			return cycles, stopError
		}

		err := fm.drainComponents(cycleResult)
		if err != nil {
			return nil, err
		}
		cycleNumber++
	}
}

// mustStop defines when f-mesh must stop after activation cycle
func (fm *FMesh) mustStop(cycleResult *cycle.Cycle) (bool, error, error) {
	if fm.HasErr() {
		return false, fm.Err(), nil
	}

	if (fm.config.CyclesLimit > 0) && (cycleResult.Number() > fm.config.CyclesLimit) {
		return true, nil, ErrReachedMaxAllowedCycles
	}

	//Check if we are done (no components activated during the cycle => all inputs are processed)
	if !cycleResult.HasActivatedComponents() {
		return true, nil, nil
	}

	//Check if mesh must stop because of configured error handling strategy
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if cycleResult.HasErrors() || cycleResult.HasPanics() {
			return true, nil, ErrHitAnErrorOrPanic
		}
		return false, nil, nil
	case StopOnFirstPanic:
		if cycleResult.HasPanics() {
			return true, nil, ErrHitAPanic
		}
		return false, nil, nil
	case IgnoreAll:
		return false, nil, nil
	default:
		return true, nil, ErrUnsupportedErrorHandlingStrategy
	}
}

// WithErr returns f-mesh with error
func (fm *FMesh) WithErr(err error) *FMesh {
	fm.SetErr(err)
	return fm
}
