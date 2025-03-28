package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create broadcaster component
	broadcaster := component.New("broadcaster").
		WithDescription("Broadcasts messages to all receivers").
		WithOutputs("messages").
		WithInputs("acks").
		WithActivationFunc(func(this *component.Component) error {
			// Send broadcast messages
			messages := []string{
				"Hello everyone!",
				"This is a broadcast message",
				"Goodbye!",
			}

			for _, msg := range messages {
				this.Logger().Printf("Broadcasting: %s", msg)
				this.OutputByName("messages").PutSignals(signal.New(msg))
				time.Sleep(200 * time.Millisecond)
			}

			// Process acknowledgments
			for _, sig := range this.InputByName("acks").AllSignalsOrNil() {
				ack := sig.PayloadOrNil().(string)
				this.Logger().Printf("Received acknowledgment: %s", ack)
			}
			return nil
		})

	// Create receiver components
	receivers := make([]*component.Component, 3)
	for i := 0; i < 3; i++ {
		receiverID := fmt.Sprintf("receiver-%d", i+1)
		receivers[i] = component.New(receiverID).
			WithDescription(fmt.Sprintf("Receives broadcast messages (%s)", receiverID)).
			WithInputs("messages").
			WithOutputs("acks").
			WithActivationFunc(func(this *component.Component) error {
				for _, sig := range this.InputByName("messages").AllSignalsOrNil() {
					msg := sig.PayloadOrNil().(string)
					this.Logger().Printf("%s received: %s", this.Name(), msg)

					// Send acknowledgment
					ack := fmt.Sprintf("%s received: %s", this.Name(), msg)
					this.OutputByName("acks").PutSignals(signal.New(ack))
					time.Sleep(100 * time.Millisecond)
				}
				return nil
			})
	}

	// Create the mesh
	mesh := fmesh.New("broadcast-example").
		WithDescription("Demonstrates broadcast pattern with acknowledgments using FMesh").
		WithComponents(append([]*component.Component{broadcaster}, receivers...)...)

	// Connect components
	for _, receiver := range receivers {
		// Connect broadcaster to each receiver
		broadcaster.OutputByName("messages").PipeTo(receiver.InputByName("messages"))
		// Connect acknowledgments back to broadcaster
		receiver.OutputByName("acks").PipeTo(broadcaster.InputByName("acks"))
	}

	// Run the mesh
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
