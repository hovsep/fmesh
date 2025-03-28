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
		WithDescription("Generates data to be filtered").
		WithOutputs("data").
		WithActivationFunc(func(this *component.Component) error {
			messages := []string{
				"Hello World",
				"Error: connection failed",
				"Processing data",
				"Error: timeout",
				"Success!",
			}

			for _, msg := range messages {
				this.Logger().Printf("Generating: %s", msg)
				this.OutputByName("data").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create filter component
	filter := component.New("filter").
		WithDescription("Filters error messages").
		WithInputs("data").
		WithOutputs("errors", "non-errors").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("data").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				if strings.Contains(msg, "Error:") {
					this.Logger().Printf("Filtering error: %s", msg)
					this.OutputByName("errors").PutSignals(signal.New(msg))
				} else {
					this.Logger().Printf("Filtering non-error: %s", msg)
					this.OutputByName("non-errors").PutSignals(signal.New(msg))
				}
			}
			return nil
		})

	// Create error handler
	errorHandler := component.New("error-handler").
		WithDescription("Handles error messages").
		WithInputs("errors").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("errors").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				this.Logger().Printf("Handling error: %s", msg)
			}
			return nil
		})

	// Create success handler
	successHandler := component.New("success-handler").
		WithDescription("Handles non-error messages").
		WithInputs("messages").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("messages").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				this.Logger().Printf("Processing success: %s", msg)
			}
			return nil
		})

	// Create the mesh
	mesh := fmesh.New("filter-example").
		WithDescription("Demonstrates filter pattern using FMesh").
		WithComponents(generator, filter, errorHandler, successHandler)

	// Connect components
	generator.OutputByName("data").PipeTo(filter.InputByName("data"))
	filter.OutputByName("errors").PipeTo(errorHandler.InputByName("errors"))
	filter.OutputByName("non-errors").PipeTo(successHandler.InputByName("messages"))

	// Start the mesh with initial signal
	generator.InputByName("data").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
