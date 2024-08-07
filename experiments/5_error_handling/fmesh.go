package main

import (
	"fmt"
	"sync"
)

type FMesh struct {
	Components Components
	ErrorHandlingStrategy
}

func (fm *FMesh) activateComponents() *HopResult {
	hop := &HopResult{
		activationResults: make(map[string]error),
	}
	activationResultsChan := make(chan ActivationResult)
	doneChan := make(chan struct{})

	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case aRes := <-activationResultsChan:
				if aRes.activated {
					hop.Lock()
					hop.activationResults[aRes.componentName] = aRes.err
					hop.Unlock()
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
			activationResultsChan <- c.activate()
		}()
	}

	wg.Wait()
	doneChan <- struct{}{}
	return hop
}

func (fm *FMesh) flushPipes() {
	for _, c := range fm.Components {
		c.flushOutputs()
	}
}

func (fm *FMesh) run() ([]*HopResult, error) {
	hops := make([]*HopResult, 0)
	for {
		hopReport := fm.activateComponents()
		hops = append(hops, hopReport)

		if fm.ErrorHandlingStrategy == StopOnFirstError && hopReport.hasErrors() {
			return hops, fmt.Errorf("Hop #%d finished with errors. Stopping fmesh. Report: %v", len(hops), hopReport.activationResults)
		}

		if len(hopReport.activationResults) == 0 {
			return hops, nil
		}
		fm.flushPipes()
	}
}
