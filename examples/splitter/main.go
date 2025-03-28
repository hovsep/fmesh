package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create data generator
	generator := component.New("generator").
		WithDescription("Generates data to be split").
		WithOutputs("data").
		WithActivationFunc(func(this *component.Component) error {
			messages := []string{
				"Hello World",
				"Processing data",
				"Testing system",
				"Deploying updates",
			}

			for _, msg := range messages {
				this.Logger().Printf("Generating: %s", msg)
				this.OutputByName("data").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create splitter component
	splitter := component.New("splitter").
		WithDescription("Splits messages into words").
		WithInputs("data").
		WithOutputs("words").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("data").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				words := strings.Fields(msg)
				for _, word := range words {
					this.Logger().Printf("Splitting word: %s", word)
					this.OutputByName("words").PutSignals(signal.New(word))
				}
			}
			return nil
		})

	// Create word processor
	processor := component.New("processor").
		WithDescription("Processes individual words").
		WithInputs("words").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("words").AllSignalsOrNil() {
				word := sig.PayloadOrNil().(string)
				this.Logger().Printf("Processing word: %s", word)
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create the mesh
	mesh := fmesh.New("splitter-example").
		WithDescription("Demonstrates splitter pattern using FMesh").
		WithComponents(generator, splitter, processor)

	// Connect components
	generator.OutputByName("data").PipeTo(splitter.InputByName("data"))
	splitter.OutputByName("words").PipeTo(processor.InputByName("words"))

	// Start the mesh with initial signal
	generator.InputByName("data").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
