package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

// This example shows how a component can have a pipe looped into it's input.
// This pattern allows to activate components multiple time using control plane (special output with looped-in pipe)
// For example we can calculate Fibonacci numbers without actually having a code for loop (the loop is implemented by ports and pipes)
func main() {
	c1 := component.New("fibonacci number generator").
		WithInputs("i_cur", "i_prev").
		WithOutputs("o_cur", "o_prev").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
			cur := inputs.ByName("i_cur").FirstSignalPayloadOrDefault(0).(int)
			prev := inputs.ByName("i_prev").FirstSignalPayloadOrDefault(0).(int)

			next := cur + prev

			//Hardcoded limit
			if next < 100 {
				fmt.Println(next)
				outputs.ByName("o_cur").PutSignals(signal.New(next))
				outputs.ByName("o_prev").PutSignals(signal.New(cur))
			}

			return nil
		})

	//Define pipes
	c1.Outputs().ByName("o_cur").PipeTo(c1.Inputs().ByName("i_cur"))
	c1.Outputs().ByName("o_prev").PipeTo(c1.Inputs().ByName("i_prev"))

	//Build mesh
	fm := fmesh.New("fibonacci example").WithComponents(c1)

	//Set inputs (first 2 Fibonacci numbers)
	f0, f1 := signal.New(0), signal.New(1)

	c1.Inputs().ByName("i_prev").PutSignals(f0)
	c1.Inputs().ByName("i_cur").PutSignals(f1)

	fmt.Println(f0.PayloadOrNil())
	fmt.Println(f1.PayloadOrNil())

	//Run the mesh
	_, err := fm.Run()

	if err != nil {
		fmt.Println(err)
	}

}
