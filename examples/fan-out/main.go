package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create producer component
	producer := component.New("producer").
		WithDescription("Produces messages to be distributed").
		WithOutputs("messages").
		WithActivationFunc(func(this *component.Component) error {
			messages := []string{
				"Message 1",
				"Message 2",
				"Message 3",
				"Message 4",
			}

			for _, msg := range messages {
				this.Logger().Printf("Producing: %s", msg)
				this.OutputByName("messages").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create worker components
	workers := make([]*component.Component, 3)
	for i := 0; i < 3; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		workers[i] = component.New(workerID).
			WithDescription(fmt.Sprintf("Processes messages (%s)", workerID)).
			WithInputs("messages").
			WithActivationFunc(func(this *component.Component) error {
				for _, sig := range this.InputByName("messages").AllSignalsOrNil() {
					msg := sig.PayloadOrNil().(string)
					this.Logger().Printf("%s processing: %s", this.Name(), msg)
					time.Sleep(200 * time.Millisecond)
				}
				return nil
			})
	}

	// Create the mesh
	mesh := fmesh.New("fan-out-example").
		WithDescription("Demonstrates fan-out pattern using FMesh").
		WithComponents(append([]*component.Component{producer}, workers...)...)

	// Connect components - fan out to all workers
	for _, worker := range workers {
		producer.OutputByName("messages").PipeTo(worker.InputByName("messages"))
	}

	// Start the mesh with initial signal
	producer.InputByName("messages").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
