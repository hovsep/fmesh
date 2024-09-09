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
// All it does is passes an integer into a simple fmesh which consists of 2 components, the first one adds 2 to the
// initial number, and the second one doubles the result.
func main() {
	c1 := &component.Component{
		Name:        "adder",
		Description: "adds 2 to the input",
		Inputs: port.Ports{
			"num": &port.Port{},
		},
		Outputs: port.Ports{
			"res": &port.Port{},
		},
		ActivationFunc: func(inputs port.Ports, outputs port.Ports) error {
			num := inputs.ByName("num").GetSignal().GetPayload().(int)
			outputs.ByName("res").PutSignal(signal.New(num + 2))
			return nil
		},
	}

	c2 := &component.Component{
		Name:        "multiplier",
		Description: "multiplies by 3",
		Inputs: port.Ports{
			"num": &port.Port{},
		},
		Outputs: port.Ports{
			"res": &port.Port{},
		},
		ActivationFunc: func(inputs port.Ports, outputs port.Ports) error {
			num := inputs.ByName("num").GetSignal().GetPayload().(int)
			outputs.ByName("res").PutSignal(signal.New(num * 3))
			return nil
		},
	}

	c1.Outputs.ByName("res").PipeTo(c2.Inputs.ByName("num"))

	fm := &fmesh.FMesh{
		Name:                  "basic fmesh",
		Description:           "",
		Components:            component.Components{c1, c2},
		ErrorHandlingStrategy: fmesh.StopOnFirstError,
	}

	c1.Inputs.ByName("num").PutSignal(signal.New(32))

	_, err := fm.Run()

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	res := c2.Outputs.ByName("res").GetSignal().GetPayload().(int)

	fmt.Println("FMesh calculation result:", res)

}
