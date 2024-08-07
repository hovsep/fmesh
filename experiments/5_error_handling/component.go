package main

import (
	"fmt"
)

type Component struct {
	name    string
	inputs  Ports
	outputs Ports
	handler func(inputs Ports, outputs Ports) error
}

type Components []*Component

func (c *Component) activate() ActivationResult {
	if !c.inputs.anyHasValue() {
		//No inputs set, stop here
		return ActivationResult{
			activated:     false,
			componentName: c.name,
			err:           nil,
		}
	}
	//Run the computation
	err := c.handler(c.inputs, c.outputs)

	//Clear inputs
	c.inputs.clearAll()

	if err != nil {
		return ActivationResult{
			activated:     true,
			componentName: c.name,
			err:           fmt.Errorf("failed to activate component: %w", err),
		}
	}

	return ActivationResult{
		activated:     true,
		componentName: c.name,
		err:           nil,
	}
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
