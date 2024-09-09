package fmesh

import (
	"fmt"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/hop"
	"sync"
)

type FMesh struct {
	Name        string
	Description string
	Components  component.Components
	ErrorHandlingStrategy
}

func (fm *FMesh) ActivateComponents() *hop.HopResult {
	hopResult := &hop.HopResult{
		ActivationResults: make(map[string]error),
	}
	activationResultsChan := make(chan hop.ActivationResult)
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

	for _, c := range fm.Components {
		wg.Add(1)
		c := c
		go func() {
			defer wg.Done()
			activationResultsChan <- c.Activate()
		}()
	}

	wg.Wait()
	doneChan <- struct{}{}
	return hopResult
}

func (fm *FMesh) FlushPipes() {
	for _, c := range fm.Components {
		c.FlushOutputs()
	}
}

func (fm *FMesh) Run() ([]*hop.HopResult, error) {
	hops := make([]*hop.HopResult, 0)
	for {
		hopReport := fm.ActivateComponents()
		hops = append(hops, hopReport)

		if fm.ErrorHandlingStrategy == StopOnFirstError && hopReport.HasErrors() {
			return hops, fmt.Errorf("Hop #%d finished with errors. Stopping fmesh. Report: %v", len(hops), hopReport.ActivationResults)
		}

		if len(hopReport.ActivationResults) == 0 {
			return hops, nil
		}
		fm.FlushPipes()
	}
}
