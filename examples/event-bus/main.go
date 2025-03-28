package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create publisher component
	publisher := component.New("publisher").
		WithDescription("Publishes events").
		WithOutputs("events").
		WithActivationFunc(func(this *component.Component) error {
			events := []string{
				"user.created",
				"order.placed",
				"payment.processed",
			}

			for _, event := range events {
				this.Logger().Printf("Publishing: %s", event)
				this.OutputByName("events").PutSignals(signal.New(event))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create subscriber components
	subscribers := make([]*component.Component, 3)
	for i := 0; i < 3; i++ {
		subscriberID := fmt.Sprintf("subscriber-%d", i+1)
		subscribers[i] = component.New(subscriberID).
			WithDescription(fmt.Sprintf("Subscribes to events (%s)", subscriberID)).
			WithInputs("events").
			WithActivationFunc(func(this *component.Component) error {
				for _, sig := range this.InputByName("events").AllSignalsOrNil() {
					event := sig.PayloadOrNil().(string)
					this.Logger().Printf("%s received: %s", this.Name(), event)
				}
				return nil
			})
	}

	// Create the mesh
	mesh := fmesh.New("event-bus-example").
		WithDescription("Demonstrates event bus pattern using FMesh").
		WithComponents(append([]*component.Component{publisher}, subscribers...)...)

	// Connect components
	for _, subscriber := range subscribers {
		publisher.OutputByName("events").PipeTo(subscriber.InputByName("events"))
	}

	// Start the mesh with initial signal
	publisher.InputByName("events").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
