package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"os"
)

// This example shows a very basic program written with FMesh
// All it does is passes an integer into a simple f-mesh which consists of 2 components, the first one adds 2 to the
// initial number, and the second one doubles the result. (result must be 102)
func main() {
	c1 := component.NewComponent("adder").
		WithDescription("adds 2 to the input").
		WithInputs("num").
		WithOutputs("res").
		WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
			num := inputs.ByName("num").Signal().Payload().(int)
			outputs.ByName("res").PutSignal(signal.New(num + 2))
			return nil
		})

	c2 := component.NewComponent("multiplier").
		WithDescription("multiplies by 3").
		WithInputs("num").
		WithOutputs("res").
		WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
			num := inputs.ByName("num").Signal().Payload().(int)
			outputs.ByName("res").PutSignal(signal.New(num * 3))
			return nil
		})

	c1.Outputs().ByName("res").PipeTo(c2.Inputs().ByName("num"))

	fm := fmesh.New("basic fmesh").WithComponents(c1, c2).WithErrorHandlingStrategy(fmesh.StopOnFirstError)

	c1.Inputs().ByName("num").PutSignal(signal.New(32))

	_, err := fm.Run()

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	res := c2.Outputs().ByName("res").Signal().Payload()

	fmt.Println("FMesh calculation result:", res)

}
