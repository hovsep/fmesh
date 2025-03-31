package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"os"
	"strings"
)

// This example is used in readme.md
func main() {
	fm := fmesh.NewWithConfig("hello world", &fmesh.Config{
		ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
		CyclesLimit:           10,
	}).
		WithComponents(
			component.New("concat").
				WithInputs("i1", "i2").
				WithOutputs("res").
				WithActivationFunc(func(this *component.Component) error {
					word1 := this.InputByName("i1").FirstSignalPayloadOrDefault("").(string)
					word2 := this.InputByName("i2").FirstSignalPayloadOrDefault("").(string)

					this.OutputByName("res").PutSignals(signal.New(word1 + word2))
					return nil
				}),
			component.New("case").
				WithInputs("i1").
				WithOutputs("res").
				WithActivationFunc(func(this *component.Component) error {
					inputString := this.InputByName("i1").FirstSignalPayloadOrDefault("").(string)

					this.OutputByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
					return nil
				}))

	fm.Components().ByName("concat").Outputs().ByName("res").PipeTo(
		fm.Components().ByName("case").Inputs().ByName("i1"),
	)

	// Init inputs
	fm.Components().ByName("concat").InputByName("i1").PutSignals(signal.New("hello "))
	fm.Components().ByName("concat").InputByName("i2").PutSignals(signal.New("world !"))

	// Run the mesh
	_, err := fm.Run()

	// Check for errors
	if err != nil {
		fmt.Println("F-Mesh returned an error")
		os.Exit(1)
	}

	// Extract results
	results := fm.Components().ByName("case").OutputByName("res").FirstSignalPayloadOrNil()
	fmt.Printf("Result is : %v", results)
}
