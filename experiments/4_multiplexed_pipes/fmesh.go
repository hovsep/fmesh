package main

import (
	"sync"
	"sync/atomic"
)

type FMesh struct {
	Components Components
}

func (fm *FMesh) activateComponents() int64 {
	var wg sync.WaitGroup
	var componentsTriggered int64
	for _, c := range fm.Components {
		wg.Add(1)
		c := c
		go func() {
			defer wg.Done()
			err := c.activate()

			if err == nil {
				atomic.AddInt64(&componentsTriggered, 1)
			}
		}()
	}

	wg.Wait()
	return componentsTriggered
}

func (fm *FMesh) flushPipes() {
	for _, c := range fm.Components {
		c.flushOutputs()
	}
}

func (fm *FMesh) run() {
	for {
		componentsTriggered := fm.activateComponents()

		if componentsTriggered == 0 {
			return
		}
		fm.flushPipes()
	}
}
