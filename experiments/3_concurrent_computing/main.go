package main

import (
	"fmt"
)

// This experiment demonstrates ability to run computations of N components (which are in the same "hop") concurrently
func main() {
	//Define components
	m1 := &Component{
		name: "m1",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			v1 := *inputs.byName("i1").getValue()
			v1 = v1 * 2
			outputs.byName("o1").setValue(&v1)
			return nil
		},
	}

	m2 := &Component{
		name: "m2",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			v1 := *inputs.byName("i1").getValue()
			v1 = v1 * 3
			outputs.byName("o1").setValue(&v1)
			return nil
		},
	}

	a1 := &Component{
		name: "a1",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			v1 := *inputs.byName("i1").getValue()
			v1 = v1 + 50
			outputs.byName("o1").setValue(&v1)
			return nil
		},
	}

	a2 := &Component{
		name: "a2",
		inputs: Ports{
			"i1": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			v1 := *inputs.byName("i1").getValue()
			v1 = v1 + 100
			outputs.byName("o1").setValue(&v1)
			return nil
		},
	}

	comb := &Component{
		name: "combiner",
		inputs: Ports{
			"i1": &Port{},
			"i2": &Port{},
		},
		outputs: Ports{
			"o1": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			v1, v2 := *inputs.byName("i1").getValue(), *inputs.byName("i2").getValue()
			r := v1 + v2
			outputs.byName("o1").setValue(&r)
			return nil
		},
	}

	//Define pipes
	pipes := Pipes{
		&Pipe{
			From: m1.outputs.byName("o1"),
			To:   a1.inputs.byName("i1"),
		},
		&Pipe{
			From: m2.outputs.byName("o1"),
			To:   a2.inputs.byName("i1"),
		},
		&Pipe{
			From: a1.outputs.byName("o1"),
			To:   comb.inputs.byName("i1"),
		},
		&Pipe{
			From: a2.outputs.byName("o1"),
			To:   comb.inputs.byName("i2"),
		},
	}

	//Build mesh
	fm := FMesh{
		Components: Components{
			m1, m2, a1, a2, comb,
		},
		Pipes: pipes,
	}

	//Set inputs
	a, b := 10, 20
	m1.inputs.byName("i1").setValue(&a)
	m2.inputs.byName("i1").setValue(&b)

	//Run the mesh
	fm.run()

	//Read outputs
	res := *comb.outputs.byName("o1").getValue()
	fmt.Printf("Result is %v", res)
}
