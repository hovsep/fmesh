package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// This example demonstrates how a component can have a pipe looped back into its own input,
// enabling a pattern that reactivates the component multiple times.
// By looping the output back into the input, the component can perform repeated calculations
// without explicit looping constructs in the code.
//
// For instance, this approach can be used to calculate Fibonacci numbers without needing
// traditional looping code. Instead, the loop is achieved by configuring ports and pipes,
// where each cycle processes a new Fibonacci term.
func main() {
	c1 := component.New("fibonacci number generator").
		WithInputs("i_cur", "i_prev").
		WithOutputs("o_cur", "o_prev").
		WithActivationFunc(func(this *component.Component) error {
			cur := this.InputByName("i_cur").FirstSignalPayloadOrDefault(0).(int)
			prev := this.InputByName("i_prev").FirstSignalPayloadOrDefault(0).(int)

			next := cur + prev

			// Hardcoded limit
			if next < 100 {
				fmt.Println(next)
				this.OutputByName("o_cur").PutSignals(signal.New(next))
				this.OutputByName("o_prev").PutSignals(signal.New(cur))
			}

			return nil
		})

	// Define pipes
	c1.Outputs().ByName("o_cur").PipeTo(c1.Inputs().ByName("i_cur"))
	c1.Outputs().ByName("o_prev").PipeTo(c1.Inputs().ByName("i_prev"))

	// Build mesh
	fm := fmesh.New("fibonacci example").WithComponents(c1)

	// Set inputs (first 2 Fibonacci numbers)
	f0, f1 := signal.New(0), signal.New(1)

	c1.Inputs().ByName("i_prev").PutSignals(f0)
	c1.Inputs().ByName("i_cur").PutSignals(f1)

	fmt.Println(f0.PayloadOrNil())
	fmt.Println(f1.PayloadOrNil())

	// Run the mesh
	_, err := fm.Run()

	if err != nil {
		fmt.Println(err)
	}

}
