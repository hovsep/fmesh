package main

import (
	"fmt"
	"strconv"
)

// This experiment shows how a component can have a pipe looped in to it's input.
// This pattern allows to activate components multiple time using control plane (special output with looped in pipe)
func main() {
	//Define components
	c1 := &Component{
		name: "c1",
		inputs: Ports{
			"i1": &Port{}, //Data plane
			"i2": &Port{}, //Control plane
		},
		outputs: Ports{
			"o1": &Port{}, //Data plane
			"o2": &Port{}, //Control plane (loop)
		},
		handler: func(inputs Ports, outputs Ports) error {
			i1 := inputs.byName("i1").getValue()

			v1 := (i1).(*SingleSignal).GetInt()

			if v1 > 100 {
				//Signal is ready to go to next component, breaking the loop
				outputs.byName("o1").setValue(&SingleSignal{
					val: v1,
				})
			} else {
				//Loop in
				outputs.byName("o2").setValue(&SingleSignal{
					val: v1 + 5,
				})
			}

			return nil
		},
	}

	c2 := &Component{
		name: "c2",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			var resSignal *SingleSignal
			i1 := inputs.byName("i1").getValue()

			//Bypass i1->o1
			if i1.IsSingle() {
				v1 := (i1).(*SingleSignal).GetInt()
				resSignal = &SingleSignal{
					val: strconv.Itoa(v1) + " suffix added once",
				}
			}

			outputs.byName("o1").setValue(resSignal)
			return nil
		},
	}

	//Define pipes
	c1.outputs.byName("o1").CreatePipeTo(c2.inputs.byName("i1"))
	c1.outputs.byName("o2").CreatePipeTo(c1.inputs.byName("i1")) //Loop

	//Build mesh
	fm := FMesh{
		Components: Components{
			c1, c2,
		},
		ErrorHandlingStrategy: StopOnFirstError,
	}

	//Set inputs
	a := &SingleSignal{
		val: 10,
	}

	c1.inputs.byName("i1").setValue(a)

	//Run the mesh
	hops, err := fm.run()
	_ = hops

	if err != nil {
		fmt.Println(err)
	}

	//Read outputs
	res := c2.outputs.byName("o1").getValue()
	fmt.Printf("Result is %v", res)
}
