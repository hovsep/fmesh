package fmesh

import (
	"fmt"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/hop"
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

// ActivateComponents tries to activate all components
func (fm *FMesh) activateComponents() *hop.HopResult {
	hopResult := &hop.HopResult{
		ActivationResults: make(map[string]error),
	}
	activationResultsChan := make(chan hop.ActivationResult) //@TODO: close the channel
	doneChan := make(chan struct{})

	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case aRes := <-activationResultsChan:
				if aRes.Activated {
					hopResult.Lock()
					hopResult.ActivationResults[aRes.ComponentName] = aRes.Err
					hopResult.Unlock()
				}
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
	return hopResult
}

// DrainComponents drains the data from all components outputs
func (fm *FMesh) drainComponents() {
	for _, c := range fm.components {
		c.FlushOutputs()
	}
}

// Run starts the computation until there is no component which activates (mesh has no unprocessed inputs)
func (fm *FMesh) Run() ([]*hop.HopResult, error) {
	hops := make([]*hop.HopResult, 0)
	for {
		hopReport := fm.activateComponents()
		hops = append(hops, hopReport) //@TODO:add collection abstraction

		//@TODO:simplify check
		if fm.errorHandlingStrategy == StopOnFirstError && hopReport.HasErrors() {
			return hops, fmt.Errorf("hop #%d finished with errors. Stopping fmesh. Report: %v", len(hops), hopReport.ActivationResults)
		}

		//@TODO:Add method
		if len(hopReport.ActivationResults) == 0 {
			//No component activated in this cycle. FMesh is ready to stop
			return hops, nil
		}
		fm.drainComponents()
	}
}
