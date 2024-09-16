package fmesh

import (
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"sync"
)

// FMesh is the functional mesh
type FMesh struct {
	name                  string
	description           string
	components            component.ComponentCollection
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

func (fm *FMesh) Components() component.ComponentCollection {
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

// runCycle runs one activation cycle (tries to activate all components)
func (fm *FMesh) runCycle() *cycle.Result {
	cycleResult := cycle.NewResult()

	if len(fm.components) == 0 {
		return cycleResult
	}

	activationResultsChan := make(chan *component.ActivationResult) //@TODO: close the channel
	doneChan := make(chan struct{})                                 //@TODO: close the channel

	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case aRes := <-activationResultsChan:
				//@TODO :check for closed channel
				cycleResult.Lock()
				cycleResult = cycleResult.WithActivationResults(aRes)
				cycleResult.Unlock()
			case <-doneChan:
				return
			}
		}
	}()

	for _, c := range fm.components {
		wg.Add(1)
		c := c //@TODO: check if this needed
		go func() {
			defer wg.Done()
			activationResultsChan <- c.MaybeActivate()
		}()
	}

	wg.Wait()
	doneChan <- struct{}{} //@TODO: no need to send close signal, just close the channel
	return cycleResult
}

// DrainComponents drains the data from all components outputs
func (fm *FMesh) drainComponents() {
	for _, c := range fm.components {
		c.FlushOutputs()
	}
}

// Run starts the computation until there is no component which activates (mesh has no unprocessed inputs)
func (fm *FMesh) Run() (cycle.Results, error) {
	allCycles := cycle.NewResults()
	cycleNumber := uint(0)
	for {
		cycleNumber++
		cycleResult := fm.runCycle().SetCycleNumber(cycleNumber)
		allCycles = allCycles.Add(cycleResult)

		mustStop, err := fm.mustStop(cycleResult)
		if mustStop {
			return allCycles, err
		}

		fm.drainComponents()
	}
}

func (fm *FMesh) mustStop(cycleResult *cycle.Result) (bool, error) {
	//Check if we are done (no components activated during the cycle => all inputs are processed)
	if !cycleResult.HasActivatedComponents() {
		return true, nil
	}

	//Check if mesh must stop because of configured error handling strategy
	switch fm.errorHandlingStrategy {
	case StopOnFirstError:
		return cycleResult.HasErrors(), ErrHitAnError
	case StopOnFirstPanic:
		return cycleResult.HasPanics(), ErrHitAPanic
	case IgnoreAll:
		return false, nil
	default:
		return true, ErrUnsupportedErrorHandlingStrategy
	}
}
