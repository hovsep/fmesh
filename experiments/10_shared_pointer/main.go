package main

import (
	"fmt"
)

// This example demonstrates a multiplexing problem in current implementation (signals are shared pointers when multiplexed)
func main() {
	//Define components
	gen := &Component{
		name: "number generator",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"num": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			outputs.byName("num").setValue(inputs.byName("i1").getValue())
			return nil
		},
	}

	r1 := &Component{
		name: "modifies input signal ",
		inputs: Ports{
			"i1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			sig := inputs.byName("i1").getValue()
			sig.(*SingleSignal).val = 666 //This modifies the signals for all receivers, as signal is a shared pointer
			return nil
		},
	}

	r2 := &Component{
		name: "receives multiplexed input signal ",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			//r2 expects 10 to come from "gen", but actually it is getting modified by r1 caused undesired behaviour
			outputs.byName("o1").setValue(inputs.byName("i1").getValue())
			return nil
		},
	}

	//Define pipes
	gen.outputs.byName("num").CreatePipesTo(r1.inputs.byName("i1"), r2.inputs.byName("i1"))

	//Build mesh
	fm := FMesh{
		Components:            Components{gen, r1, r2},
		ErrorHandlingStrategy: StopOnFirstError,
	}

	//Set inputs

	gen.inputs.byName("i1").setValue(&SingleSignal{val: 10})

	//Run the mesh
	hops, err := fm.run()
	_ = hops

	res := r2.outputs.byName("o1").getValue()

	fmt.Printf("r2 received : %v", res.(*SingleSignal).GetVal())

	if err != nil {
		fmt.Println(err)
	}

}
