package main

import (
	"fmt"
)

// This example demonstrates the ability of meshes to be nested (a component of mesh can be a mesh itself and nesting depth is unlimited)
func main() {
	//Define components
	c1 := &Component{
		name:        "math",
		description: "a * b + c",
		inputs: Ports{
			"a": &Port{},
			"b": &Port{},
			"c": &Port{},
		},
		outputs: Ports{
			"out": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			if !inputs.manyByName("a", "b", "c").allHaveValue() {
				return errWaitingForInputKeepInputs
			}

			//This component is using a nested mesh for multiplication
			multiplierWithLogger := getSubMesh()

			//Pass inputs
			forwardSignal(inputs.byName("a"), multiplierWithLogger.Components.byName("Multiplier").inputs.byName("a"))
			forwardSignal(inputs.byName("b"), multiplierWithLogger.Components.byName("Multiplier").inputs.byName("b"))

			//Run submesh inside a component
			multiplierWithLogger.run()

			//Read the multiplication result
			multiplicationResult := multiplierWithLogger.Components.byName("Multiplier").outputs.byName("result").getSignal().GetValue().(int)

			//Do the rest of calculation
			res := multiplicationResult + inputs.byName("c").getSignal().GetValue().(int)

			outputs.byName("out").putSignal(newSignal(res))
			return nil
		},
	}

	c2 := &Component{
		name:        "add constant",
		description: "a + 35",
		inputs: Ports{
			"a": &Port{},
		},
		outputs: Ports{
			"out": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			if inputs.byName("a").hasSignal() {
				a := inputs.byName("a").getSignal().GetValue().(int)
				outputs.byName("out").putSignal(newSignal(a + 35))
			}
			return nil
		},
	}

	//Define pipes
	c1.outputs.byName("out").CreatePipesTo(c2.inputs.byName("a"))

	//Build mesh
	fm := &FMesh{
		Components:            Components{c1, c2},
		ErrorHandlingStrategy: StopOnFirstError,
	}

	//Set inputs

	c1.inputs.byName("a").putSignal(newSignal(2))
	c1.inputs.byName("b").putSignal(newSignal(3))
	c1.inputs.byName("c").putSignal(newSignal(4))

	//Run the mesh
	hops, err := fm.run()
	if err != nil {
		fmt.Println(err)
	}
	_ = hops

	res := c2.outputs.byName("out").getSignal().GetValue()

	fmt.Printf("outter fmesh result %v", res)
}

func getSubMesh() *FMesh {
	multiplier := &Component{
		name:        "Multiplier",
		description: "This component multiplies numbers on it's inputs",
		inputs: Ports{
			"a": &Port{},
			"b": &Port{},
		},
		outputs: Ports{
			"bypass_a": &Port{},
			"bypass_b": &Port{},

			"result": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			//@TODO: simplify waiting API
			if !inputs.manyByName("a", "b").allHaveValue() {
				return errWaitingForInputKeepInputs
			}

			//Bypass input signals, so logger can get them
			forwardSignal(inputs.byName("a"), outputs.byName("bypass_a"))
			forwardSignal(inputs.byName("b"), outputs.byName("bypass_b"))

			a, b := inputs.byName("a").getSignal().GetValue().(int), inputs.byName("b").getSignal().GetValue().(int)

			outputs.byName("result").putSignal(newSignal(a * b))
			return nil
		},
	}

	logger := &Component{
		name:        "Logger",
		description: "This component logs inputs of multiplier",
		inputs: Ports{
			"a": &Port{},
			"b": &Port{},
		},
		outputs: nil, //No output
		handler: func(inputs Ports, outputs Ports) error {
			if inputs.byName("a").hasSignal() {
				fmt.Println(fmt.Sprintf("Inner logger says: a is %v", inputs.byName("a").getSignal().GetValue()))
			}

			if inputs.byName("b").hasSignal() {
				fmt.Println(fmt.Sprintf("Inner logger says: b is %v", inputs.byName("b").getSignal().GetValue()))
			}

			return nil
		},
	}

	multiplier.outputs.byName("bypass_a").CreatePipesTo(logger.inputs.byName("a"))
	multiplier.outputs.byName("bypass_b").CreatePipesTo(logger.inputs.byName("b"))

	return &FMesh{
		Name:                  "Logged multiplicator",
		Description:           "multiply 2 numbers and log inputs into std out",
		Components:            Components{multiplier, logger},
		ErrorHandlingStrategy: StopOnFirstError,
	}
}
