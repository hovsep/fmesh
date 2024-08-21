package main

import (
	"fmt"
)

// This experiment shows how components can "wait" for particular inputs.
// Here we have c5 component which just sums up the outcomes of c4 and the chain of c1->c2->c3
// As c4 output is ready in second hop the c5 is triggered, but the problem is the other part coming from the chain is not ready yet,
// because in second hop only c2 is triggered. So c5 has to wait somehow while both inputs will be ready.
// By sending special sentinel errors we instruct mesh how to treat given component.
// So starting from this version component can tell the mesh if it's inputs must be reset or kept.
func main() {
	//Define components
	c1 := &Component{
		name: "c1",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			var resSignal *SingleSignal
			i1 := inputs.byName("i1").getValue()

			if i1.IsSingle() {
				v1 := (i1).(*SingleSignal).GetInt()
				resSignal = &SingleSignal{
					val: v1 + 3,
				}
			}

			outputs.byName("o1").setValue(resSignal)
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

			if i1.IsSingle() {
				v1 := (i1).(*SingleSignal).GetInt()
				resSignal = &SingleSignal{
					val: v1 + 5,
				}
			}

			outputs.byName("o1").setValue(resSignal)
			return nil
		},
	}

	c3 := &Component{
		name: "c3",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			var resSignal *SingleSignal
			i1 := inputs.byName("i1").getValue()

			if i1.IsSingle() {
				v1 := (i1).(*SingleSignal).GetInt()
				resSignal = &SingleSignal{
					val: v1 * 2,
				}
			}

			outputs.byName("o1").setValue(resSignal)
			return nil
		},
	}

	c4 := &Component{
		name: "c4",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			var resSignal *SingleSignal
			i1 := inputs.byName("i1").getValue()

			if i1.IsSingle() {
				v1 := (i1).(*SingleSignal).GetInt()
				resSignal = &SingleSignal{
					val: v1 + 10,
				}
			}

			outputs.byName("o1").setValue(resSignal)
			return nil
		},
	}

	c5 := &Component{
		name: "c5",
		inputs: Ports{
			"i1": &Port{},
			"i2": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {

			// This component is a basic summator, it must trigger only when both inputs are set
			if !inputs.manyByName("i1", "i2").allHaveValue() {
				return errWaitingForInputKeepInputs
			}

			i1 := inputs.byName("i1").getValue()
			i2 := inputs.byName("i2").getValue()

			resSignal := &SingleSignal{
				val: (i1).(*SingleSignal).GetInt() + (i2).(*SingleSignal).GetInt(),
			}

			outputs.byName("o1").setValue(resSignal)
			return nil
		},
	}

	//Build mesh
	fm := FMesh{
		Components: Components{
			c1, c2, c3, c4, c5,
		},
		ErrorHandlingStrategy: StopOnFirstError,
	}

	//Define pipes
	c1.outputs.byName("o1").CreatePipeTo(c2.inputs.byName("i1"))
	c2.outputs.byName("o1").CreatePipeTo(c3.inputs.byName("i1"))
	c3.outputs.byName("o1").CreatePipeTo(c5.inputs.byName("i1"))
	c4.outputs.byName("o1").CreatePipeTo(c5.inputs.byName("i2"))

	//Set inputs
	a := &SingleSignal{
		val: 10,
	}

	b := &SingleSignal{
		val: 2,
	}

	c1.inputs.byName("i1").setValue(a)
	c4.inputs.byName("i1").setValue(b)

	//Run the mesh
	hops, err := fm.run()
	_ = hops

	if err != nil {
		fmt.Println(err)
	}

	//Read outputs
	res := c5.outputs.byName("o1").getValue()
	fmt.Printf("Result is %v", res)
}
