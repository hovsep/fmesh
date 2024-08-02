package main

import "fmt"

// This experiment aims to demonstrate components with multiple inputs and outputs
// In this case the component acts like a multivariable function
func main() {

	//Define components
	c1 := &Component{
		name: "mul 2, mul 10",
		inputs: Ports{
			"i1": &Port{},
			"i2": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
			"o2": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) {
			i1, i2 := inputs.byName("i1"), inputs.byName("i2")
			if i1.hasValue() && i2.hasValue() {
				//Merge 2 input signals and put it onto single output
				c := *i1.getValue()*2 + *i2.getValue()*10
				outputs.byName("o1").setValue(&c)
			}

			//We can generate output signal without any input
			c4 := 4
			outputs.byName("o2").setValue(&c4)
		},
	}

	c2 := &Component{
		name: "add 3 or input",
		inputs: Ports{
			"i1": &Port{},
			"i2": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) {
			c := 3
			if inputs.byName("i2").hasValue() {
				c = *inputs.byName("i2").getValue()
			}
			t := *inputs.byName("i1").getValue() + c
			outputs.byName("o1").setValue(&t)
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
		handler: func(inputs Ports, outputs Ports) {
			t := *inputs.byName("i1").getValue() + 5
			outputs.byName("o1").setValue(&t)
		},
	}

	//Define pipes
	pipes := Pipes{
		&Pipe{
			In:  c1.outputs.byName("o1"),
			Out: c2.inputs.byName("i1"),
		},
		&Pipe{
			In:  c1.outputs.byName("o2"),
			Out: c2.inputs.byName("i2"),
		},
		&Pipe{
			In:  c2.outputs.byName("o1"),
			Out: c3.inputs.byName("i1"),
		},
	}

	//Build mesh
	fm := FMesh{
		Components: Components{
			c1, c2, c3,
		},
		Pipes: pipes,
	}

	//Set inputs
	a, b := 10, 20
	c1.inputs.byName("i1").setValue(&a)
	c1.inputs.byName("i2").setValue(&b)

	//Run the mesh
	fm.run()

	//Read outputs
	res := *c3.outputs.byName("o1").getValue()
	fmt.Printf("Result is %v", res)
}
