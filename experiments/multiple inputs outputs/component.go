package main

type Component struct {
	name    string
	inputs  Ports
	outputs Ports
	handler func(inputs Ports, outputs Ports)
}

type Components []*Component

func (c *Component) compute() {
	if !c.inputs.anyHasValue() {
		//No inputs set, nothing to compute
		return
	}
	//Run the computation
	c.handler(c.inputs, c.outputs)

	//Clear inputs
	c.inputs.setAll(nil)
}
