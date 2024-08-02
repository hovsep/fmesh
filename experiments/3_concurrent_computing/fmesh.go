package main

import (
	"sync"
	"sync/atomic"
)

type FMesh struct {
	Components Components
	Pipes      Pipes
}

func (fm *FMesh) compute() int64 {
	var wg sync.WaitGroup
	var componentsTriggered int64
	for _, c := range fm.Components {
		wg.Add(1)
		c := c
		go func() {
			defer wg.Done()
			err := c.compute()

			if err == nil {
				atomic.AddInt64(&componentsTriggered, 1)
			}
		}()
	}

	wg.Wait()
	return componentsTriggered
}

func (fm *FMesh) flush() {
	for _, p := range fm.Pipes {
		p.flush()
	}
}

func (fm *FMesh) run() {
	for {
		componentsTriggered := fm.compute()

		if componentsTriggered == 0 {
			return
		}
		fm.flush()
	}
}
