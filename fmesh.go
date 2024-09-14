package fmesh

import (
	"fmt"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"sync"
)

// FMesh is the functional mesh
type FMesh struct {
	name                  string
	description           string
	components            component.Components
	errorHandlingStrategy ErrorHandlingStrategy
}

// New creates a new f-mesh
func New(name string) *FMesh {
	return &FMesh{name: name}
}

// WithDescription sets a description
func (fm *FMesh) WithDescription(description string) *FMesh {
	fm.description = description
	return fm
}

// WithComponents adds components to f-mesh
func (fm *FMesh) WithComponents(components ...*component.Component) *FMesh {
	if fm.components == nil {
		fm.components = component.NewComponents()
	}
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

// ActivateComponents runs one activation cycle (tries to activate all components)
func (fm *FMesh) activateComponents() *cycle.Result {
	cycleResult := cycle.NewResult()

	activationResultsChan := make(chan component.ActivationResult) //@TODO: close the channel
	doneChan := make(chan struct{})

	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case aRes := <-activationResultsChan:
				cycleResult.Lock()
				cycleResult.ActivationResults[aRes.ComponentName()] = aRes
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
	doneChan <- struct{}{}
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
	for {
		cycleResult := fm.activateComponents()

		if fm.shouldStop(cycleResult) {
			return allCycles, fmt.Errorf("cycle #%d finished with errors. Stopping fmesh. Report: %v", len(allCycles), cycleResult.ActivationResults)
		}

		if !cycleResult.HasActivatedComponents() {
			//No component activated in this cycle. FMesh is ready to stop
			return allCycles, nil
		}

		allCycles = append(allCycles, cycleResult)
		fm.drainComponents()
	}
}

func (fm *FMesh) shouldStop(cycleResult *cycle.Result) bool {
	switch fm.errorHandlingStrategy {
	case StopOnFirstError:
		if cycleResult.HasErrors() {
			return true
		}
	case IgnoreAll:
		return false
	default:
		panic("unsupported error handling strategy")
	}
	return false
}
