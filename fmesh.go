package fmesh

import (
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
	name        string
	description string
	components  component.Collection
	config      Config
}

// New creates a new f-mesh
func New(name string) *FMesh {
	return &FMesh{
		name:       name,
		components: component.NewCollection(),
		config:     defaultConfig,
	}
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
		fm.components = fm.components.With(c)
	}
	return fm
}

// WithConfig sets the configuration and returns the f-mesh
func (fm *FMesh) WithConfig(config Config) *FMesh {
	fm.config = config
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
func (fm *FMesh) drainComponents(cycle *cycle.Cycle) {
	for _, c := range fm.Components() {
		activationResult := cycle.ActivationResults().ByComponentName(c.Name())

		if !activationResult.Activated() {
			continue
		}

		if component.IsWaitingForInput(activationResult) {
			if !component.WantsToKeepInputs(activationResult) {
				c.ClearInputs()
			}
			// Components waiting for inputs are not flushed
			continue
		}

		// Normally components are fully drained
		c.FlushOutputs()
		c.ClearInputs()
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

		fm.drainComponents(cycleResult)
	}
}

func (fm *FMesh) mustStop(cycleResult *cycle.Cycle, cycleNum int) (bool, error) {
	if (fm.config.CyclesLimit > 0) && (cycleNum > fm.config.CyclesLimit) {
		return true, ErrReachedMaxAllowedCycles
	}

	//Check if we are done (no components activated during the cycle => all inputs are processed)
	if !cycleResult.HasActivatedComponents() {
		return true, nil
	}

	//Check if mesh must stop because of configured error handling strategy
	switch fm.config.ErrorHandlingStrategy {
	case StopOnFirstErrorOrPanic:
		if cycleResult.HasErrors() || cycleResult.HasPanics() {
			return true, ErrHitAnErrorOrPanic
		}
		return false, nil
	case StopOnFirstPanic:
		if cycleResult.HasPanics() {
			return true, ErrHitAPanic
		}
		return false, nil
	case IgnoreAll:
		return false, nil
	default:
		return true, ErrUnsupportedErrorHandlingStrategy
	}
}
