package main

import (
	"errors"
	"fmt"
)

// This experiment enhances error handling capabilities
func main() {
	//Define components

	c0 := &Component{
		name: "this component just brakes on start",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			return errors.New("i'm feelin baaad")
		},
	}

	c1 := &Component{
		name: "add 3",
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
		name: "mul 2",
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

	c3 := &Component{
		name: "add 5",
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

	c4 := &Component{
		name: "agg sum",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			var resSignal *SingleSignal
			i1 := inputs.byName("i1").getValue()

			if i1.IsAggregate() {
				a1 := (i1).(*AggregateSignal)

				sum := 0
				for _, v := range a1.val {
					sum += v.GetInt()
				}

				resSignal = &SingleSignal{
					val: sum,
				}
			}

			outputs.byName("o1").setValue(resSignal)
			return nil
		},
	}

	//Build mesh
	fm := FMesh{
		ErrorHandlingStrategy: StopOnFirstError,
		Components: Components{
			c0, c1, c2, c3, c4,
		},
	}

	//Define pipes
	c1.outputs.byName("o1").CreatePipeTo(c2.inputs.byName("i1"))
	c1.outputs.byName("o1").CreatePipeTo(c3.inputs.byName("i1"))
	c2.outputs.byName("o1").CreatePipeTo(c4.inputs.byName("i1"))
	c3.outputs.byName("o1").CreatePipeTo(c4.inputs.byName("i1"))

	//Set inputs
	a := &SingleSignal{
		val: 10,
	}
	c1.inputs.byName("i1").setValue(a)
	c0.inputs.byName("i1").setValue(&SingleSignal{
		val: 34,
	})

	//Run the mesh
	hops, err := fm.run()
	_ = hops

	if err != nil {
		fmt.Println(err)
	}

	//Read outputs
	res := c4.outputs.byName("o1").getValue()
	fmt.Printf("Result is %v", res)
}
