package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Observer Pattern Example
// This example demonstrates how to implement the observer pattern using FMesh components.
// It shows how to:
// 1. Create a subject component that maintains state and notifies observers
// 2. Create multiple observer components that react to state changes
// 3. Broadcast state changes to all observers simultaneously
// The pattern is useful for implementing event handling systems where multiple
// components need to react to changes in another component's state.
// Common use cases include:
// - GUI event handling
// - System monitoring
// - Event logging
func main() {
	// Create subject component
	subject := component.New("subject").
		WithDescription("Subject that notifies observers of state changes").
		WithInputs("start").
		WithOutputs("state-changes").
		WithActivationFunc(func(this *component.Component) error {
			states := []string{
				"Initializing",
				"Running",
				"Paused",
				"Running",
				"Stopped",
			}

			for _, state := range states {
				this.Logger().Printf("Subject state changed to: %s", state)
				this.OutputByName("state-changes").PutSignals(signal.New(state))
				time.Sleep(200 * time.Millisecond)
			}
			return nil
		})

	// Create observer components
	observers := make([]*component.Component, 3)
	for i := 0; i < 3; i++ {
		observerID := fmt.Sprintf("observer-%d", i+1)
		observers[i] = component.New(observerID).
			WithDescription(fmt.Sprintf("Observer %s that reacts to state changes", observerID)).
			WithInputs("state-changes").
			WithActivationFunc(func(this *component.Component) error {
				for _, sig := range this.InputByName("state-changes").AllSignalsOrNil() {
					state := sig.PayloadOrNil().(string)
					this.Logger().Printf("%s received state change: %s", this.Name(), state)
					time.Sleep(100 * time.Millisecond)
				}
				return nil
			})
	}

	// Create the mesh
	mesh := fmesh.New("observer-example").
		WithDescription("Demonstrates observer pattern using FMesh").
		WithComponents(append([]*component.Component{subject}, observers...)...)

	// Connect components
	for _, observer := range observers {
		subject.OutputByName("state-changes").PipeTo(observer.InputByName("state-changes"))
	}

	// Start the mesh with initial signal
	subject.InputByName("start").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
