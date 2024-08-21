package main

import (
	"fmt"
)

// This experiment shows how a component can have a pipe looped in to it's input.
// This pattern allows to activate components multiple time using control plane (special output with looped in pipe)
func main() {
	//Define components
	c1 := &Component{
		name: "c1",
		inputs: Ports{
			"i_cur":  &Port{},
			"i_prev": &Port{},
		},
		outputs: Ports{
			"o_cur":  &Port{},
			"o_prev": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			iCur := inputs.byName("i_cur").getValue()
			iPrev := inputs.byName("i_prev").getValue()

			vCur := iCur.(*SingleSignal).GetInt()
			vPrev := iPrev.(*SingleSignal).GetInt()

			vNext := vCur + vPrev

			if vNext < 150 {
				fmt.Println(vNext)
				outputs.byName("o_cur").setValue(&SingleSignal{val: vNext})
				outputs.byName("o_prev").setValue(&SingleSignal{val: vCur})
			}

			return nil
		},
	}

	//Define pipes
	c1.outputs.byName("o_cur").CreatePipeTo(c1.inputs.byName("i_cur"))
	c1.outputs.byName("o_prev").CreatePipeTo(c1.inputs.byName("i_prev"))

	//Build mesh
	fm := FMesh{
		Components: Components{
			c1,
		},
		ErrorHandlingStrategy: StopOnFirstError,
	}

	//Set inputs
	f0 := &SingleSignal{
		val: 0,
	}

	f1 := &SingleSignal{
		val: 1,
	}

	c1.inputs.byName("i_prev").setValue(f0)
	c1.inputs.byName("i_cur").setValue(f1)

	fmt.Println(f0.val)
	fmt.Println(f1.val)

	//Run the mesh
	hops, err := fm.run()
	_ = hops

	if err != nil {
		fmt.Println(err)
	}

}
