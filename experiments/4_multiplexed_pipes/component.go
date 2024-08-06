package main

import "errors"

type Component struct {
	name    string
	inputs  Ports
	outputs Ports
	handler func(inputs Ports, outputs Ports) error
}

type Components []*Component

func (c *Component) activate() error {
	if !c.inputs.anyHasValue() {
		//No inputs set, nothing to activateComponents
		return errors.New("no inputs set")
	}
	//Run the computation
	err := c.handler(c.inputs, c.outputs)
	if err != nil {
		return err
	}

	//Clear inputs
	c.inputs.clearAll()
	return nil
}

func (c *Component) flushOutputs() {
	for _, out := range c.outputs {
		if !out.hasValue() || len(out.pipes) == 0 {
			continue
		}

		for _, pipe := range out.pipes {
			//Multiplexing
			pipe.To.setValue(out.getValue())
		}
		out.clearValue()
	}
}
