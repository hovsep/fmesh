package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"os"
	"strings"
)

// This example is used in readme.md
func main() {
	fm := fmesh.New("hello world").
		WithComponents(
			component.New("concat").
				WithInputs("i1", "i2").
				WithOutputs("res").
				WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
					word1 := inputs.ByName("i1").FirstSignalPayloadOrDefault("").(string)
					word2 := inputs.ByName("i2").FirstSignalPayloadOrDefault("").(string)

					outputs.ByName("res").PutSignals(signal.New(word1 + word2))
					return nil
				}),
			component.New("case").
				WithInputs("i1").
				WithOutputs("res").
				WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
					inputString := inputs.ByName("i1").FirstSignalPayloadOrDefault("").(string)

					outputs.ByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
					return nil
				})).
		WithConfig(fmesh.Config{
			ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
			CyclesLimit:           10,
		})

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

	//Extract results
	results := fm.Components().ByName("case").OutputByName("res").FirstSignalPayloadOrNil()
	fmt.Printf("Result is : %v", results)
}
