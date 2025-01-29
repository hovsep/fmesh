package fmesh

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"log"
	"os"
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
	cycles     *cycle.Group
	config     Config
	logger     *log.Logger
}

// New creates a new f-mesh
func New(name string) *FMesh {
	return &FMesh{
		NamedEntity:     common.NewNamedEntity(name),
		DescribedEntity: common.NewDescribedEntity(""),
		Chainable:       common.NewChainable(),
		components:      component.NewCollection(),
		cycles:          cycle.NewGroup(),
		config:          defaultConfig,
		logger:          getDefaultLogger(),
	}
}

// Components getter
func (fm *FMesh) Components() *component.Collection {
	if fm.HasErr() {
		return component.NewCollection().WithErr(fm.Err())
	}
	return fm.components
}

//@TODO: add shortcut method: ComponentByName()

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
func (fm *FMesh) runCycle() {
	newCycle := cycle.New().WithNumber(fm.cycles.Len() + 1)

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

	fm.cycles = fm.cycles.With(newCycle)

	return
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
			return true, fmt.Errorf("%w, cycle # %d", ErrHitAnErrorOrPanic, lastCycle.Number())
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

func (fm *FMesh) WithLogger(logger *log.Logger) *FMesh {
	if fm.HasErr() {
		return fm
	}

	fm.logger = logger
	return fm
}

func (fm *FMesh) Logger() *log.Logger {
	return fm.logger
}

func getDefaultLogger() *log.Logger {
	logger := log.Default()
	logger.SetOutput(os.Stdout)
	logger.SetFlags(log.LstdFlags | log.Lmsgprefix)
	return logger
}
