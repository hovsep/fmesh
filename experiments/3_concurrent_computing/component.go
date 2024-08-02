package main

import "errors"

type Component struct {
	name    string
	inputs  Ports
	outputs Ports
	handler func(inputs Ports, outputs Ports) error
}

type Components []*Component

func (c *Component) compute() error {
	if !c.inputs.anyHasValue() {
		//No inputs set, nothing to compute
		return errors.New("no inputs set")
	}
	//Run the computation
	c.handler(c.inputs, c.outputs)

	//Clear inputs
	c.inputs.setAll(nil)
	return nil
}
