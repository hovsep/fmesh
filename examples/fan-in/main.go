package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create source components
	sources := make([]*component.Component, 3)
	for i := 0; i < 3; i++ {
		sourceID := fmt.Sprintf("source-%d", i+1)
		sources[i] = component.New(sourceID).
			WithDescription(fmt.Sprintf("Data source %s", sourceID)).
			WithOutputs("data").
			WithActivationFunc(func(this *component.Component) error {
				for j := 0; j < 3; j++ {
					msg := fmt.Sprintf("Message %d from %s", j+1, this.Name())
					this.Logger().Printf("Generating: %s", msg)
					this.OutputByName("data").PutSignals(signal.New(msg))
					time.Sleep(100 * time.Millisecond)
				}
				return nil
			})
	}

	// Create merger component
	merger := component.New("merger").
		WithDescription("Merges data from multiple sources").
		WithInputs("input-1", "input-2", "input-3").
		WithOutputs("merged").
		WithActivationFunc(func(this *component.Component) error {
			for i := 1; i <= 3; i++ {
				inputName := fmt.Sprintf("input-%d", i)
				for _, sig := range this.InputByName(inputName).AllSignalsOrNil() {
					msg := sig.PayloadOrNil().(string)
					this.Logger().Printf("Received from %s: %s", inputName, msg)
					this.OutputByName("merged").PutSignals(signal.New(msg))
				}
			}
			return nil
		})

	// Create consumer component
	consumer := component.New("consumer").
		WithDescription("Consumes merged data").
		WithInputs("data").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("data").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				this.Logger().Printf("Consuming: %s", msg)
			}
			return nil
		})

	// Create the mesh
	mesh := fmesh.New("fan-in-example").
		WithDescription("Demonstrates fan-in pattern using FMesh").
		WithComponents(append(sources, merger, consumer)...)

	// Connect components
	for i, src := range sources {
		src.OutputByName("data").PipeTo(merger.InputByName(fmt.Sprintf("input-%d", i+1)))
	}
	merger.OutputByName("merged").PipeTo(consumer.InputByName("data"))

	// Start the mesh with initial signals
	for _, src := range sources {
		src.InputByName("data").PutSignals(signal.New("start"))
	}

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
